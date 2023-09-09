package a2

import (
	"log/slog"
	"time"
)

// Server is an OAuth2 implementation
type Server struct {
	Config            *ServerConfig
	Storage           Storage
	AuthorizeTokenGen AuthorizeTokenGen
	AccessTokenGen    AccessTokenGen
	Now               func() time.Time
	Logger            *slog.Logger
}

func New(log *slog.Logger) *Server {
	return &Server{
		Config: &ServerConfig{
			AuthorizationExpiration: int32(time.Duration(5 * time.Minute).Seconds()),
			AccessExpiration:        int32(time.Duration(5 * 24 * time.Hour).Seconds()),
			TokenType:               "Bearer",
			AllowedAuthorizeTypes:   AllowedAuthorizeType{TOKEN},
			AllowedAccessTypes: AllowedAccessType{
				CLIENT_CREDENTIALS, PASSWORD, REFRESH_TOKEN,
			},
			ErrorStatusCode: 200,
		},
		Storage:           Provider{},
		AuthorizeTokenGen: JWT{},
		AccessTokenGen:    JWT{},
		Now:               time.Now,
		Logger:            log.With("component", "a2"),
	}
}

// NewServer creates a new server instance
func NewServer(config *ServerConfig, storage Storage) *Server {
	return &Server{
		Config:            config,
		Storage:           storage,
		AuthorizeTokenGen: &AuthorizeTokenGenDefault{},
		AccessTokenGen:    &AccessTokenGenDefault{},
		Now:               time.Now,
	}
}

// NewResponse creates a new response for the server
func (s *Server) NewResponse() *Response {
	r := NewResponse(s.Storage)
	r.ErrorStatusCode = s.Config.ErrorStatusCode
	return r
}
