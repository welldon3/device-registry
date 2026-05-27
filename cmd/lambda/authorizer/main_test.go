package main

import (
	"testing"
)

func TestIsAuthorized(t *testing.T) {
	orig := apiSecret
	t.Cleanup(func() { apiSecret = orig })
	apiSecret = "test-secret"

	cases := []struct {
		header string
		want   bool
	}{
		{"Bearer test-secret", true},
		{"bearer test-secret", true},
		{"BEARER test-secret", true},
		{"Bearer wrong-secret", false},
		{"Bearer ", false},
		{"", false},
		{"Basic test-secret", false},
		{"test-secret", false},
	}

	for _, c := range cases {
		got := isAuthorized(c.header)
		if got != c.want {
			t.Errorf("isAuthorized(%q) = %v, want %v", c.header, got, c.want)
		}
	}
}

func TestIsAuthorized_EmptySecret(t *testing.T) {
	orig := apiSecret
	t.Cleanup(func() { apiSecret = orig })
	apiSecret = ""

	if isAuthorized("Bearer anything") {
		t.Fatal("expected false when apiSecret is empty")
	}
}