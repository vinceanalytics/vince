package a2

import (
	"context"
	"strconv"
	"time"
)

type TestingStorage struct {
	clients   map[string]Client
	authorize map[string]*AuthorizeData
	access    map[string]*AccessData
	refresh   map[string]string
}

func NewTestingStorage() *TestingStorage {
	r := &TestingStorage{
		clients:   make(map[string]Client),
		authorize: make(map[string]*AuthorizeData),
		access:    make(map[string]*AccessData),
		refresh:   make(map[string]string),
	}

	r.clients["1234"] = &DefaultClient{
		Id:          "1234",
		Secret:      "aabbccdd",
		RedirectUri: "http://localhost:14000/appauth",
	}

	r.clients["public-client"] = &DefaultClient{
		Id:          "public-client",
		RedirectUri: "http://localhost:14000/appauth",
	}

	r.authorize["9999"] = &AuthorizeData{
		Client:      r.clients["1234"],
		Code:        "9999",
		ExpiresIn:   3600,
		CreatedAt:   time.Now(),
		RedirectUri: "http://localhost:14000/appauth",
	}

	r.access["9999"] = &AccessData{
		Client:        r.clients["1234"],
		AuthorizeData: r.authorize["9999"],
		AccessToken:   "9999",
		ExpiresIn:     3600,
		CreatedAt:     time.Now(),
	}

	r.access["r9999"] = &AccessData{
		Client:        r.clients["1234"],
		AuthorizeData: r.authorize["9999"],
		AccessData:    r.access["9999"],
		AccessToken:   "9999",
		RefreshToken:  "r9999",
		ExpiresIn:     3600,
		CreatedAt:     time.Now(),
	}

	r.refresh["r9999"] = "9999"

	return r
}

func (s *TestingStorage) Clone(_ context.Context) Storage {
	return s
}

func (s *TestingStorage) Close(_ context.Context) {
}

func (s *TestingStorage) GetClient(_ context.Context, id string) (Client, error) {
	if c, ok := s.clients[id]; ok {
		return c, nil
	}
	return nil, ErrNotFound
}

func (s *TestingStorage) SetClient(_ context.Context, id string, client Client) error {
	s.clients[id] = client
	return nil
}

func (s *TestingStorage) SaveAuthorize(_ context.Context, data *AuthorizeData) error {
	s.authorize[data.Code] = data
	return nil
}

func (s *TestingStorage) LoadAuthorize(_ context.Context, code string) (*AuthorizeData, error) {
	if d, ok := s.authorize[code]; ok {
		return d, nil
	}
	return nil, ErrNotFound
}

func (s *TestingStorage) RemoveAuthorize(_ context.Context, code string) error {
	delete(s.authorize, code)
	return nil
}

func (s *TestingStorage) SaveAccess(_ context.Context, data *AccessData) error {
	s.access[data.AccessToken] = data
	if data.RefreshToken != "" {
		s.refresh[data.RefreshToken] = data.AccessToken
	}
	return nil
}

func (s *TestingStorage) LoadAccess(_ context.Context, code string) (*AccessData, error) {
	if d, ok := s.access[code]; ok {
		return d, nil
	}
	return nil, ErrNotFound
}

func (s *TestingStorage) RemoveAccess(_ context.Context, code string) error {
	delete(s.access, code)
	return nil
}

func (s *TestingStorage) LoadRefresh(ctx context.Context, code string) (*AccessData, error) {
	if d, ok := s.refresh[code]; ok {
		return s.LoadAccess(ctx, d)
	}
	return nil, ErrNotFound
}

func (s *TestingStorage) RemoveRefresh(_ context.Context, code string) error {
	delete(s.refresh, code)
	return nil
}

// Predictable testing token generation

type TestingAuthorizeTokenGen struct {
	counter int64
}

func (a *TestingAuthorizeTokenGen) GenerateAuthorizeToken(_ context.Context, data *AuthorizeData) (ret string, err error) {
	a.counter++
	return strconv.FormatInt(a.counter, 10), nil
}

type TestingAccessTokenGen struct {
	acounter int64
	rcounter int64
}

func (a *TestingAccessTokenGen) GenerateAccessToken(_ context.Context, data *AccessData, generaterefresh bool) (accesstoken string, refreshtoken string, err error) {
	a.acounter++
	accesstoken = strconv.FormatInt(a.acounter, 10)

	if generaterefresh {
		a.rcounter++
		refreshtoken = "r" + strconv.FormatInt(a.rcounter, 10)
	}
	return
}
