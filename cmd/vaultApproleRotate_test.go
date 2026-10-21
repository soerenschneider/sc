package cmd

import (
	"context"
	"errors"
	"net"
	"slices"
	"testing"
)

func TestLookupHostCidrs(t *testing.T) {
	lookup := func(_ context.Context, host string) ([]net.IPAddr, error) {
		switch host {
		case "a.example.com":
			return []net.IPAddr{{IP: net.ParseIP("192.168.1.10")}, {IP: net.ParseIP("2001:db8::1")}}, nil
		case "b.example.com":
			return []net.IPAddr{{IP: net.ParseIP("192.168.1.10")}, {IP: net.ParseIP("10.0.0.1")}}, nil
		case "empty.example.com":
			return nil, nil
		}
		return nil, errors.New("no such host")
	}

	got, err := lookupHostCidrs(context.Background(), lookup, []string{"a.example.com", "b.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"192.168.1.10/32", "2001:db8::1/128", "10.0.0.1/32"}; !slices.Equal(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}

	for _, host := range []string{"unknown.example.com", "empty.example.com"} {
		if _, err := lookupHostCidrs(context.Background(), lookup, []string{host}); err == nil {
			t.Errorf("expected an error for %q", host)
		}
	}
}
