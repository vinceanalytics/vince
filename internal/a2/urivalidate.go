package a2

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// error returned when validation don't match
type UriValidationError string

func (e UriValidationError) Error() string {
	return string(e)
}

func newUriValidationError(msg string, base string, redirect string) UriValidationError {
	return UriValidationError(fmt.Sprintf("%s: %s / %s", msg, base, redirect))
}

// ParseUrls parses URLs, resolving references to the base url
func ParseUrls(baseUrl, redirectUrl string) (retBaseUrl, retRedirectUrl *url.URL, err error) {
	var base, redirect *url.URL
	// parse base url
	if base, err = url.Parse(baseUrl); err != nil {
		return nil, nil, err
	}

	// parse redirect url
	if redirect, err = url.Parse(redirectUrl); err != nil {
		return nil, nil, err
	}

	// must not have fragment
	if base.Fragment != "" || redirect.Fragment != "" {
		return nil, nil, newUriValidationError("url must not include fragment.", baseUrl, redirectUrl)
	}

	// Scheme must match
	if redirect.Scheme != base.Scheme {
		return nil, nil, newUriValidationError("scheme mismatch", baseUrl, redirectUrl)
	}

	var (
		redirectMatch bool
		host          string
	)

	// verify the hosts match
	if redirect.Host == base.Host {
		redirectMatch = true
		host = base.Host
	} else if baseIP := net.ParseIP(base.Host); baseIP != nil && baseIP.IsLoopback() && base.Scheme == "http" {
		// if the registered redirect URI is lookback, then match the input redirect without the port.
		if redirectIP := net.ParseIP(redirect.Hostname()); redirectIP != nil && redirectIP.IsLoopback() {
			redirectMatch = true
			host = redirect.Host
		}
	}

	// Hosts must match
	if !redirectMatch {
		return nil, nil, newUriValidationError("host mismatch", baseUrl, redirectUrl)
	}

	// resolve references to base url
	retBaseUrl = (&url.URL{Scheme: base.Scheme, Host: host}).ResolveReference(&url.URL{Path: base.Path})
	retRedirectUrl = (&url.URL{Scheme: base.Scheme, Host: host}).ResolveReference(&url.URL{Path: redirect.Path, RawQuery: redirect.RawQuery})
	return
}

// ValidateUriList validates that redirectUri is contained in baseUriList.
// baseUriList may be a string separated by separator.
// If separator is blank, validate only 1 URI.
func ValidateUriList(baseUriList string, redirectUri string, separator string) (realRedirectUri string, err error) {
	// make a list of uris
	var slist []string
	if separator != "" {
		slist = strings.Split(baseUriList, separator)
	} else {
		slist = make([]string, 0)
		slist = append(slist, baseUriList)
	}

	for _, sitem := range slist {
		realRedirectUri, err = ValidateUri(sitem, redirectUri)
		// validated, return no error
		if err == nil {
			return realRedirectUri, nil
		}

		// if there was an error that is not a validation error, return it
		if _, iok := err.(UriValidationError); !iok {
			return "", err
		}
	}

	return "", newUriValidationError("urls don't validate", baseUriList, redirectUri)
}

// ValidateUri validates that redirectUri is contained in baseUri
func ValidateUri(baseUri string, redirectUri string) (realRedirectUri string, err error) {
	if baseUri == "" || redirectUri == "" {
		return "", errors.New("urls cannot be blank.")
	}

	base, redirect, err := ParseUrls(baseUri, redirectUri)
	if err != nil {
		return "", err
	}

	// allow exact path matches
	if base.Path == redirect.Path {
		return redirect.String(), nil
	}

	// ensure prefix matches are actually subpaths
	requiredPrefix := strings.TrimRight(base.Path, "/") + "/"
	if !strings.HasPrefix(redirect.Path, requiredPrefix) {
		return "", newUriValidationError("path prefix doesn't match", baseUri, redirectUri)
	}

	return redirect.String(), nil
}

// FirstUri returns the first uri from a URI list
func FirstUri(baseUriList string, separator string) string {
	if separator != "" {
		slist := strings.Split(baseUriList, separator)
		if len(slist) > 0 {
			return slist[0]
		}
	} else {
		return baseUriList
	}

	return ""
}
