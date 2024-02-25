// Package auth is a lightweight credential store.
// It provides functionality for loading credentials, as well as validating credentials.
package auth

import (
	"os"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// AllUsers is the username that indicates all users, even anonymous users (requests without
	// any BasicAuth information).
	AllUsers = "*"
)

// BasicAuther is the interface an object must support to return basic auth information.
type BasicAuther interface {
	BasicAuth() (string, string, bool)
}

// CredentialsStore stores authentication and authorization information for all users.
type CredentialsStore struct {
	store map[string]string
	perms map[string]map[v1.Credential_Permission]struct{}
}

// NewCredentialsStore returns a new instance of a CredentialStore.
func NewCredentialsStore() *CredentialsStore {
	return &CredentialsStore{
		store: make(map[string]string),
		perms: make(map[string]map[v1.Credential_Permission]struct{}),
	}
}

// NewCredentialsStoreFromFile returns a new instance of a CredentialStore loaded from a file.
func NewCredentialsStoreFromFile(path string) (*CredentialsStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ls v1.Credential_List
	err = protojson.Unmarshal(data, &ls)
	if err != nil {
		return nil, err
	}
	c := NewCredentialsStore()
	return c, c.Load(&ls)
}

// Load loads credential information from a reader.
func (c *CredentialsStore) Load(ls *v1.Credential_List) error {
	for _, e := range ls.Items {
		c.store[e.Username] = e.Password
		p := make(map[v1.Credential_Permission]struct{})
		for _, v := range e.Perms {
			p[v] = struct{}{}
		}
		c.perms[e.Username] = p
	}
	return nil
}

// Check returns true if the password is correct for the given username.
func (c *CredentialsStore) Check(username, password string) bool {
	pw, ok := c.store[username]
	return ok && pw == password
}

// Password returns the password for the given user.
func (c *CredentialsStore) Password(username string) (string, bool) {
	pw, ok := c.store[username]
	return pw, ok
}

// CheckRequest returns true if b contains a valid username and password.
func (c *CredentialsStore) CheckRequest(b BasicAuther) bool {
	username, password, ok := b.BasicAuth()
	if !ok || !c.Check(username, password) {
		return false
	}
	return true
}

// HasPerm returns true if username has the given perm, either directly or
// via AllUsers. It does not perform any password checking.
func (c *CredentialsStore) HasPerm(username string, perm v1.Credential_Permission) bool {
	if m, ok := c.perms[username]; ok {
		if _, ok := m[perm]; ok {
			return true
		}
	}

	if m, ok := c.perms[AllUsers]; ok {
		if _, ok := m[perm]; ok {
			return true
		}
	}

	return false
}

// HasAnyPerm returns true if username has at least one of the given perms,
// either directly, or via AllUsers. It does not perform any password checking.
func (c *CredentialsStore) HasAnyPerm(username string, perm ...v1.Credential_Permission) bool {
	for i := range perm {
		if c.HasPerm(username, perm[i]) {
			return true
		}
	}
	return false
}

// AA authenticates and checks authorization for the given username and password
// for the given perm. If the credential store is nil, then this function always
// returns true. If AllUsers have the given perm, authentication is not done.
// Only then are the credentials checked, and then the perm checked.
func (c *CredentialsStore) AA(username, password string, perm v1.Credential_Permission) bool {
	// No credential store? Auth is not even enabled.
	if c == nil {
		return true
	}

	// Is the required perm granted to all users, including anonymous users?
	if c.HasAnyPerm(AllUsers, perm, v1.Credential_ALL) {
		return true
	}

	// At this point a username needs to have been supplied.
	if username == "" {
		return false
	}

	// Authenticate the user.
	if !c.Check(username, password) {
		return false
	}

	// Is the specified user authorized?
	return c.HasAnyPerm(username, perm, v1.Credential_ALL)
}

// HasPermRequest returns true if the username returned by b has the givem perm.
// It does not perform any password checking, but if there is no username
// in the request, it returns false.
func (c *CredentialsStore) HasPermRequest(b BasicAuther, perm v1.Credential_Permission) bool {
	username, _, ok := b.BasicAuth()
	return ok && c.HasPerm(username, perm)
}
