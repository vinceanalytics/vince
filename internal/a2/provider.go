package a2

import (
	"errors"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/auth/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type provider struct {
	db db.Provider
}

func New(db db.Provider) Storage {
	return &provider{}
}

var _ Storage = (*provider)(nil)

func (provider) Close()            {}
func (p *provider) Clone() Storage { return p }

func (p provider) GetClient(id string) (Client, error) {
	var o v1.AuthorizedClient
	err := p.get(keys.AClient(id), &o)
	if err != nil {
		return nil, err
	}
	return xclient{b: &o}, nil
}

func (p provider) SaveAuthorize(a *AuthorizeData) error {
	return p.save(
		keys.AAuthorize(a.Code), AuthorizeDataTo(a),
	)
}

func (p provider) LoadAuthorize(code string) (*AuthorizeData, error) {
	var o v1.AuthorizeData
	err := p.get(keys.AAuthorize(code), &o)
	if err != nil {
		return nil, err
	}
	return AuthorizeDataFrom(&o), nil
}

func (p provider) RemoveAuthorize(code string) error {
	return p.remove(keys.AAuthorize(code))
}

func (p provider) SaveAccess(a *AccessData) error {
	return errors.Join(
		p.save(
			keys.AAccess(a.AccessToken), AccessDataTo(a),
		),
		p.save(
			keys.ARefresh(a.RefreshToken), AccessDataTo(a),
		),
	)
}

func (p provider) LoadAccess(token string) (*AccessData, error) {
	var o v1.AccessData
	err := p.get(keys.AAccess(token), &o)
	if err != nil {
		return nil, err
	}
	return AccessDataFrom(&o), nil
}

func (p provider) RemoveAccess(token string) error {
	return p.remove(keys.AAccess(token))
}

func (p provider) LoadRefresh(token string) (*AccessData, error) {
	var o v1.AccessData
	err := p.get(keys.ARefresh(token), &o)
	if err != nil {
		return nil, err
	}
	return AccessDataFrom(&o), nil
}

func (p provider) RemoveRefresh(token string) error {
	return p.remove(keys.ARefresh(token))
}

func (p provider) save(key *keys.Key, m proto.Message) error {
	b := must.Must(proto.Marshal(m))("failed serializing data")
	return p.db.Txn(true, func(txn db.Txn) error {
		defer key.Release()
		return txn.Set(key.Bytes(), b)
	})
}

func (p provider) remove(key *keys.Key) error {
	return p.db.Txn(true, func(txn db.Txn) error {
		defer key.Release()
		return txn.Delete(key.Bytes())
	})
}

func (p provider) get(key *keys.Key, o proto.Message) error {
	return p.db.Txn(false, func(txn db.Txn) error {
		defer key.Release()
		return txn.Get(key.Bytes(), func(val []byte) error {
			return proto.Unmarshal(val, o)
		}, func() error {
			return ErrNotFound
		})
	})

}

func AuthorizeDataFrom(a *v1.AuthorizeData) *AuthorizeData {
	if a == nil {
		return nil
	}
	return &AuthorizeData{
		Client:              xclient{b: a.Client},
		Code:                a.Code,
		ExpiresIn:           a.ExpiresIn,
		Scope:               a.Scope,
		RedirectUri:         a.RedirectUri,
		State:               a.State,
		CreatedAt:           a.CreatedAt.AsTime(),
		CodeChallenge:       a.CodeChallenge,
		CodeChallengeMethod: a.CodeChallengeMethod,
	}
}

func AuthorizeDataTo(a *AuthorizeData) *v1.AuthorizeData {
	return &v1.AuthorizeData{
		Client: &v1.AuthorizedClient{
			Id:          a.Client.GetId(),
			Secret:      a.Client.GetSecret(),
			RedirectUrl: a.Client.GetRedirectUri(),
		},
		Code:                a.Code,
		ExpiresIn:           a.ExpiresIn,
		Scope:               a.Scope,
		RedirectUri:         a.RedirectUri,
		State:               a.State,
		CreatedAt:           timestamppb.New(a.CreatedAt),
		CodeChallenge:       a.CodeChallenge,
		CodeChallengeMethod: a.CodeChallengeMethod,
	}
}

func AccessDataFrom(a *v1.AccessData) *AccessData {
	if a == nil {
		return nil
	}
	return &AccessData{
		Client:        xclient{b: a.Client},
		AuthorizeData: AuthorizeDataFrom(a.AuthorizeData),
		AccessData:    AccessDataFrom(a.AccessData),
		AccessToken:   a.AccessToken,
		RefreshToken:  a.RefreshToken,
		ExpiresIn:     a.ExpiresIn,
		Scope:         a.Scope,
		RedirectUri:   a.RedirectUri,
		CreatedAt:     a.CreatedAt.AsTime(),
	}
}

func AccessDataTo(a *AccessData) *v1.AccessData {
	if a == nil {
		return nil
	}
	return &v1.AccessData{
		Client: &v1.AuthorizedClient{
			Id:          a.Client.GetId(),
			Secret:      a.Client.GetSecret(),
			RedirectUrl: a.Client.GetRedirectUri(),
		},
		AuthorizeData: AuthorizeDataTo(a.AuthorizeData),
		AccessData:    AccessDataTo(a.AccessData),
		AccessToken:   a.AccessToken,
		RefreshToken:  a.RefreshToken,
		ExpiresIn:     a.ExpiresIn,
		Scope:         a.Scope,
		RedirectUri:   a.RedirectUri,
		CreatedAt:     timestamppb.New(a.CreatedAt),
	}
}

type xclient struct {
	b *v1.AuthorizedClient
}

var _ Client = (*xclient)(nil)

func (x xclient) GetId() string {
	return x.b.Id
}

func (x xclient) GetSecret() string {
	return x.b.Secret
}

func (x xclient) GetRedirectUri() string {
	return x.b.Secret
}

func (x xclient) GetUserData() any {
	return nil
}
