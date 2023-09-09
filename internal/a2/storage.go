package a2

import (
	"context"
	"errors"
)

var (
	// ErrNotFound is the error returned by Storage Get<...> and Load<...> functions in case
	// no entity is found in the storage. E.g. Storage.GetClient() returns ErrNotFound when
	// client is not found. All other returned errors must be treated as storage-specific errors,
	// like "connection lost", "connection refused", etc.
	ErrNotFound = errors.New("Entity not found")
)

// Storage interface
type Storage interface {
	// Clone the storage if needed. For example, using mgo, you can clone the session with session.Clone
	// to avoid concurrent access problems.
	// This is to avoid cloning the connection at each method access.
	// Can return itself if not a problem.
	Clone(context.Context) Storage

	// Close the resources the Storage potentially holds (using Clone for example)
	Close(context.Context)

	// GetClient loads the client by id (client_id)
	GetClient(ctx context.Context, id string) (Client, error)

	// SaveAuthorize saves authorize data.
	SaveAuthorize(context.Context, *AuthorizeData) error

	// LoadAuthorize looks up AuthorizeData by a code.
	// Client information MUST be loaded together.
	// Optionally can return error if expired.
	LoadAuthorize(ctx context.Context, code string) (*AuthorizeData, error)

	// RemoveAuthorize revokes or deletes the authorization code.
	RemoveAuthorize(ctx context.Context, code string) error

	// SaveAccess writes AccessData.
	// If RefreshToken is not blank, it must save in a way that can be loaded using LoadRefresh.
	SaveAccess(context.Context, *AccessData) error

	// LoadAccess retrieves access data by token. Client information MUST be loaded together.
	// AuthorizeData and AccessData DON'T NEED to be loaded if not easily available.
	// Optionally can return error if expired.
	LoadAccess(ctx context.Context, token string) (*AccessData, error)

	// RemoveAccess revokes or deletes an AccessData.
	RemoveAccess(ctx context.Context, token string) error

	// LoadRefresh retrieves refresh AccessData. Client information MUST be loaded together.
	// AuthorizeData and AccessData DON'T NEED to be loaded if not easily available.
	// Optionally can return error if expired.
	LoadRefresh(ctx context.Context, token string) (*AccessData, error)

	// RemoveRefresh revokes or deletes refresh AccessData.
	RemoveRefresh(ctx context.Context, token string) error
}
