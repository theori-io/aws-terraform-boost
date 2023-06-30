package utils

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
)

const xmlHeader = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"

// StringifyXML returns the string representation of a value.
// This is an incomplete XML port of prettify in aws-sdk-go
func StringifyXML(i interface{}, rootName string) string {
	var buf bytes.Buffer

	buf.WriteString(xmlHeader)
	stringify(reflect.ValueOf(i), 0, &buf, rootName)

	return buf.String()
}

func stringify(v reflect.Value, indent int, buf *bytes.Buffer, rootName string) {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		strtype := v.Type().String()
		if strtype == "time.Time" {
			fmt.Fprintf(buf, "%s", v.Interface())
			break
		} else if strings.HasPrefix(strtype, "io.") {
			buf.WriteString("<buffer>")
			break
		}

		if len(rootName) != 0 {
			buf.WriteString("<" + rootName + ">")
		}

		names := []string{}
		for i := 0; i < v.Type().NumField(); i++ {
			name := v.Type().Field(i).Name
			f := v.Field(i)
			if name[0:1] == strings.ToLower(name[0:1]) {
				continue // ignore unexported fields
			}
			if (f.Kind() == reflect.Ptr || f.Kind() == reflect.Slice || f.Kind() == reflect.Map) && f.IsNil() {
				continue // ignore unset fields
			}
			names = append(names, name)
		}

		for i, n := range names {
			val := v.FieldByName(n)
			ft, ok := v.Type().FieldByName(n)
			if !ok {
				panic(fmt.Sprintf("expected to find field %v on type %v, but was not found", n, v.Type()))
			}

			buf.WriteString(strings.Repeat(" ", indent+2))
			buf.WriteString("<" + ft.Tag.Get("locationName") + ">")

			stringify(val, indent+2, buf, ft.Tag.Get("locationNameList"))
			buf.WriteString("</" + ft.Tag.Get("locationName") + ">")

			if i < len(names)-1 {
				buf.WriteString("\n")
			}
		}

		if len(rootName) != 0 {
			buf.WriteString("\n" + strings.Repeat(" ", indent) + "</" + rootName + ">")
		}
	case reflect.Slice:
		strtype := v.Type().String()
		if strtype == "[]uint8" {
			fmt.Fprintf(buf, "<binary> len %d", v.Len())
			break
		}

		nl, id, id2 := "", "", ""
		if v.Len() > 3 {
			nl, id, id2 = "\n", strings.Repeat(" ", indent), strings.Repeat(" ", indent+2)
		}
		buf.WriteString(nl)
		for i := 0; i < v.Len(); i++ {
			buf.WriteString(id2)
			stringify(v.Index(i), indent+2, buf, rootName)

			if i < v.Len()-1 {
				buf.WriteString(nl)
			}
		}

		buf.WriteString(nl + id)
	case reflect.Map:
		buf.WriteString("<map>")

		for i, k := range v.MapKeys() {
			buf.WriteString(strings.Repeat(" ", indent+2))
			buf.WriteString(k.String() + ": ")
			stringify(v.MapIndex(k), indent+2, buf, "")

			if i < v.Len()-1 {
				buf.WriteString("\n")
			}
		}

		buf.WriteString("\n" + strings.Repeat(" ", indent) + "</map>")
	default:
		if !v.IsValid() {
			fmt.Fprint(buf, "<invalid value>")
			return
		}
		switch v.Interface().(type) {
		case string:
			xml.Escape(buf, []byte(v.Interface().(string)))
		default:
			xml.Escape(buf, []byte(fmt.Sprintf("%v", v.Interface())))
		}
	}
}
