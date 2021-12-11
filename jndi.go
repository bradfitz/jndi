// Package jndi lets Gophers participate in the log4j fun (CVE-2021-44228).
//
// It would be irresponsible to use this package.
package jndi

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// NewLogger returns a new Logger that does expansion and evalation of jndi expression
// within the user-influenceable log text.
func NewLogger() *log.Logger {
	return log.New(Wrap(os.Stderr), "", log.LstdFlags)
}

func Wrap(w io.Writer) io.Writer {
	return writer{w, realEnv}
}

var realEnv = env{
	transport: http.DefaultTransport,
	getEnv:    os.Getenv,
}

type env struct {
	transport http.RoundTripper
	getEnv    func(string) string
}

type writer struct {
	ww io.Writer
	e  env
}

func (w writer) Write(p []byte) (n int, err error) {
	n = len(p)
	_, err = io.WriteString(w.ww, w.e.subst(string(p)))
	return n, err // close enough
}

var opRx = regexp.MustCompile(`\$\{(\w+?):(?:[^}\$]|(\$[^\{]))+}`)

func (e env) subst(s string) string {
	for {
		s2 := opRx.ReplaceAllStringFunc(s, func(sub string) string {
			i := strings.Index(sub, ":")
			return e.lookup(sub[2:i], sub[i+1:len(sub)-1])
		})
		if s2 == s {
			return s2
		}
		s = s2
	}
}

// see https://logging.apache.org/log4j/2.x/manual/lookups.html
func (e env) lookup(op, arg string) string {
	switch op {
	case "lower":
		return strings.ToLower(arg)
	case "upper":
		return strings.ToUpper(arg)
	case "env":
		if s := e.getEnv(strings.ToLower(arg)); s != "" {
			return s
		}
		if s := e.getEnv(strings.ToUpper(arg)); s != "" {
			return s
		}
	case "jndi":
		// I looked at gopkg.in/ldap.v2 and got scared.
		// A GET request is enough to do some DNS data and leak
		// some environment variable secrets.
		urlStr := strings.Replace(arg, "ldap://", "http://", 1) // oh well

		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return err.Error()
		}
		res, err := e.transport.RoundTrip(req)
		if err != nil {
			return err.Error()
		}
		all, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err.Error()
		}
		return string(all)
	}
	return ""
}
