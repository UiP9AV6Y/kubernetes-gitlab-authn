package template

import (
	"fmt"
	htmltemplate "html/template"
	"net"
	"strconv"
	"time"
)

// Functions contains various rendering functions suitable for
// usage with Golang templates.
var Functions = htmltemplate.FuncMap{
	"comment":   HTMLComment,
	"hostName":  HostName,
	"hostPort":  HostPort,
	"rfcDate":   RFCDate,
	"unixEpoch": UnixEpoch,
}

// HTMLComment renders the input parameters as string
// using [fmt.Sprint] and encloses them in HTML comment brackets.
func HTMLComment(v ...interface{}) htmltemplate.HTML {
	comment := fmt.Sprint(v...)

	return htmltemplate.HTML("<!-- " + comment + " -->")
}

// HostName returns the hostname part of [net.SplitHostPort]
func HostName(v string) (name string, err error) {
	name, _, err = net.SplitHostPort(v)

	return
}

// HostName returns the port part of [net.SplitHostPort]
func HostPort(v string) (port string, err error) {
	_, port, err = net.SplitHostPort(v)

	return
}

// RFCDate renders the provided time using [time.RFC3339]
func RFCDate(v time.Time) string {
	return v.Format(time.RFC3339)
}

// UnixEpoch returns the elapsed time since the beginning
// of the UNIX epoch in seconds as string.
func UnixEpoch(v time.Time) string {
	return strconv.FormatInt(v.Unix(), 10)
}
