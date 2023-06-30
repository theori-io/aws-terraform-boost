package main

import (
	"flag"
	"log"
	"os/user"
	"path/filepath"

	aws_terraform_boost "theori.io/aws-terraform-boost"
)

// Returns ~/.aws/credentials
func defaultAwsCredentialPath() string {
	x, _ := user.Current()
	credentialsFile := filepath.Join(x.HomeDir, ".aws/credentials")
	return credentialsFile
}

func main() {
	var pemPath string
	flag.StringVar(&pemPath, "pem", "certs/fullchain.pem", "path to pem file")

	var keyPath string
	flag.StringVar(&keyPath, "key", "certs/server.key", "path to key file")

	var addr_plain string
	flag.StringVar(&addr_plain, "addr", ":10001", "Listening address (HTTP; use this as HTTPS_PROXY)")

	var addr_ssl string
	flag.StringVar(&addr_ssl, "addr_ssl", ":10002", "Listening address (SSL)")

	var aws_profile string
	flag.StringVar(&aws_profile, "aws-profile", "default", "AWS Profile Name")

	var credentials_file string
	flag.StringVar(&credentials_file, "credentials-file", defaultAwsCredentialPath(), "credentials file Name")

	flag.Parse()
	server_plain, server_https := aws_terraform_boost.NewServer(
		addr_plain, addr_ssl, credentials_file, aws_profile)

	go func() { log.Fatal(server_plain.ListenAndServe()) }()
	log.Fatal(server_https.ListenAndServeTLS(pemPath, keyPath))
}
