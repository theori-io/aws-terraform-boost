package aws_terraform_boost

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"theori.io/aws-terraform-boost/internal/interceptor"
	"theori.io/aws-terraform-boost/internal/types"
)

func handleTunneling(w http.ResponseWriter, addr string, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	log.Printf("Connect: %s", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func Proxy(account *types.Account) *ReverseProxy {
	cache := make(map[string]*types.ActionCache)

	actionHandlers := interceptor.AWSInterceptors()
	for actionName := range actionHandlers {
		cache[actionName] = types.NewActionCache()
	}

	return &ReverseProxy{
		Rewrite: func(r *ProxyRequest) *http.Response {
			r.Out.URL.Host = r.In.Host
			r.Out.URL.Scheme = "https"
			r.Out.Host = r.In.Host

			body, _ := io.ReadAll(r.In.Body)
			q, _ := url.ParseQuery(string(body[:]))

			// Extract region from the request
			region := regionFromHost(r.In.Host)
			overridenConfig := aws.NewConfig()
			overridenConfig.Region = &region

			action := q.Get("Action")
			if handler, ok := actionHandlers[action]; ok {
				filtered_body := handler(q, cache[action], account, overridenConfig)
				return intercept(filtered_body, r)
			}

			r.Out.Body = io.NopCloser(bytes.NewBuffer(body))
			return nil
		},
	}
}

// Extract "us-east-1" from "ec2.us-east-1.amazonaws.com"
func regionFromHost(host string) string {
	end := strings.LastIndex(host, ".amazonaws.com")
	return host[strings.LastIndex(host[:end], ".")+1 : end]
}

func intercept(filtered_body []byte, r *ProxyRequest) *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header: http.Header{
			"Content-Length": {fmt.Sprintf("%d", len(filtered_body))},
			"Content-Type":   {"application/xml"},
		},
		Body:             io.NopCloser(bytes.NewReader(filtered_body)),
		ContentLength:    int64(len(filtered_body)),
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     true,
		Trailer:          map[string][]string{},
		Request:          r.Out,
		TLS:              r.Out.TLS,
	}
}

func NewServer(addr string, addrSsl string, credentialsFile string, awsProfile string) (*http.Server, *http.Server) {
	serverPlain := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("HTTP | method: %s", r.Method)
			if r.Method == http.MethodConnect {
				handleTunneling(w, addrSsl, r)
			}
		}),

		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	rev := Proxy(types.NewAccount(credentialsFile, awsProfile))
	serverHttps := &http.Server{
		Addr: addrSsl,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("HTTPS | method: %s | host: %s | path: %s", r.Method, r.Host, r.URL)
			rev.ServeHTTP(w, r)
		}),
	}
	return serverPlain, serverHttps
}
