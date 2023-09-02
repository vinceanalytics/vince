package a2

import (
	"fmt"
	"testing"
)

func TestURIValidate(t *testing.T) {
	valid := []struct {
		name              string
		clientRedirectURI string
		inputRedirectURI  string
		normalized        string
	}{
		{
			"Exact match",
			"http://localhost:14000/appauth",
			"http://localhost:14000/appauth",
			"http://localhost:14000/appauth",
		},
		{
			"Only domain, exact match",
			"http://google.com",
			"http://google.com",
			"http://google.com",
		},
		{
			"Only domain, with subpath",
			"http://google.com",
			"http://google.com/redirect",
			"http://google.com/redirect",
		},
		{
			"Trailing slash",
			"http://www.google.com/myapp",
			"http://www.google.com/myapp/",
			"http://www.google.com/myapp/",
		},
		{
			"Exact match with trailing slash",
			"http://www.google.com/myapp/",
			"http://www.google.com/myapp/",
			"http://www.google.com/myapp/",
		},
		{
			"Subpath",
			"http://www.google.com/myapp",
			"http://www.google.com/myapp/interface/implementation",
			"http://www.google.com/myapp/interface/implementation",
		},
		{
			"Subpath with trailing slash",
			"http://www.google.com/myapp/",
			"http://www.google.com/myapp/interface/implementation",
			"http://www.google.com/myapp/interface/implementation",
		},
		{
			"Subpath with things that are close to path traversals, but aren't",
			"http://www.google.com/myapp",
			"http://www.google.com/myapp/.../..implementation../...",
			"http://www.google.com/myapp/.../..implementation../...",
		},
		{
			"If the allowed basepath contains path traversals, allow them?",
			"http://www.google.com/traversal/../allowed",
			"http://www.google.com/traversal/../allowed/with/subpath",
			"http://www.google.com/allowed/with/subpath",
		},
		{
			"Backslashes",
			"https://mysafewebsite.com/secure/redirect",
			"https://mysafewebsite.com/secure/redirect/\\../\\../\\../evil",
			"https://mysafewebsite.com/secure/redirect/%5C../%5C../%5C../evil",
		},
		{
			"Query string must be kept",
			"http://www.google.com/myapp/redir",
			"http://www.google.com/myapp/redir?a=1&b=2",
			"http://www.google.com/myapp/redir?a=1&b=2",
		},
		{
			"IPv4 loopback address",
			"http://127.0.0.1/callback",
			"http://127.0.0.1:8081/callback",
			"http://127.0.0.1:8081/callback",
		},
		{
			"Uncommon IPv4 loopback address",
			"http://127.0.0.1/callback",
			"http://127.0.0.8:8081/callback",
			"http://127.0.0.8:8081/callback",
		},
		{
			"IPv6 loopback address",
			"http://127.0.0.1/callback",
			"http://[::1]:8081/callback",
			"http://[::1]:8081/callback",
		},
		{
			"No port in IPv4 loopback address",
			"http://127.0.0.1/callback",
			"http://127.0.0.1/callback",
			"http://127.0.0.1/callback",
		},
		{
			"No port in IPv6 loopback address",
			"http://127.0.0.1/callback",
			"http://[0:0:0:0:0:0:0:1]/callback",
			"http://[0:0:0:0:0:0:0:1]/callback",
		},
	}
	for _, v := range valid {
		t.Run(fmt.Sprintf("valid/%s", v.name), func(t *testing.T) {
			if realRedirectUri, err := ValidateUri(v.clientRedirectURI, v.inputRedirectURI); err != nil {
				t.Errorf("Expected ValidateUri(%s, %s) to succeed, got %v", v.clientRedirectURI, v.inputRedirectURI, err)
			} else if realRedirectUri != v.normalized {
				t.Errorf("Expected ValidateUri(%s, %s) to return uri %s, got %s", v.clientRedirectURI, v.inputRedirectURI, v.normalized, realRedirectUri)
			}
		})
	}

	invalid := []struct {
		name              string
		clientRedirectURI string
		inputRedirectURI  string
	}{
		{
			"Doesn't satisfy base path",
			"http://localhost:14000/appauth",
			"http://localhost:14000/app",
		},
		{
			"Doesn't satisfy base path",
			"http://localhost:14000/app/",
			"http://localhost:14000/app",
		},
		{
			"Not a subpath of base path",
			"http://localhost:14000/appauth",
			"http://localhost:14000/appauthmodifiedpath",
		},
		{
			"Host mismatch",
			"http://www.google.com/myapp",
			"http://www2.google.com/myapp",
		},
		{
			"Scheme mismatch",
			"http://www.google.com/myapp",
			"https://www.google.com/myapp",
		},
		{
			"Path traversal",
			"http://www.google.com/myapp",
			"http://www.google.com/myapp/..",
		},
		{
			"Embedded path traversal",
			"http://www.google.com/myapp",
			"http://www.google.com/myapp/../test",
		},
		{
			"Not a subpath",
			"http://www.google.com/myapp",
			"http://www.google.com/myapp../test",
		},
		{
			"Backslashes",
			"https://mysafewebsite.com/secure/redirect",
			"https://mysafewebsite.com/secure%2fredirect/../evil",
		},
		{
			"Mismatching ports",
			"https://example.com:8081/redirect",
			"https://example.com:8080/redirect",
		},
		{
			"Non loopback address",
			"http://127.0.0.1/callback",
			"http://localhost:8080/callback",
		},
		{
			"Invalid input redirect URI",
			"http://127.0.0.1/callback",
			"http://127.0.0.1:abc/callback",
		},
		{
			"Non http scheme in redirect URI",
			"custom://127.0.0.1/callback",
			"custom://127.0.0.1:8080/callback",
		},
		{
			"Redirect URI is loopback, input is a domain name with port",
			"http://127.0.0.1/callback",
			"http://example.com:8081/callback",
		},
		{
			"Redirect URI is loopback, input is a domain name without port",
			"http://127.0.0.1/callback",
			"http://example.com/callback",
		},
	}
	for _, v := range invalid {
		t.Run(fmt.Sprintf("invalid/%s", v.name), func(t *testing.T) {
			if _, err := ValidateUri(v.clientRedirectURI, v.inputRedirectURI); err == nil {
				t.Errorf("Expected ValidateUri(%s, %s) to fail", v.clientRedirectURI, v.inputRedirectURI)
			}
		})
	}
}

func TestURIListValidate(t *testing.T) {
	// V1
	if _, err := ValidateUriList("http://localhost:14000/appauth", "http://localhost:14000/appauth", ""); err != nil {
		t.Errorf("V1: %s", err)
	}

	// V2
	if _, err := ValidateUriList("http://localhost:14000/appauth", "http://localhost:14000/app", ""); err == nil {
		t.Error("V2 should have failed")
	}

	// V3
	if _, err := ValidateUriList("http://xxx:14000/appauth;http://localhost:14000/appauth", "http://localhost:14000/appauth", ";"); err != nil {
		t.Errorf("V3: %s", err)
	}

	// V4
	if _, err := ValidateUriList("http://xxx:14000/appauth;http://localhost:14000/appauth", "http://localhost:14000/app", ";"); err == nil {
		t.Error("V4 should have failed")
	}
}
