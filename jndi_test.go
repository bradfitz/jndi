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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "got-path=%v", r.URL.Path)
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
		in, want string
	}{
		{"foo${env:user}bar", "fooalicebar"},
		{"lower ${lower:FOO}", "lower foo"},
		{"upper ${upper:foo}", "upper FOO"},
		{"nested ${env:${lower:u}ser}", "nested alice"},
		{"hit network: ${jndi:ldap://foo.com/bar}", "hit network: got-path=/bar"},
	}
	for _, tt := range tests {
		if got := e.subst(tt.in); got != tt.want {
			t.Errorf("subst(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
