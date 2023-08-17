package engine

import (
	"errors"
	"net"

	"github.com/dgraph-io/badger/v4"
	"github.com/dolthub/vitess/go/mysql"
	querypb "github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/vinceanalytics/vince/internal/keys"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
)

type Auth struct {
	DB *badger.DB
}

var _ mysql.AuthServer = (*Auth)(nil)

func (a Auth) Salt() ([]byte, error) {
	return mysql.NewSalt()
}

func (a Auth) ValidateHash(salt []byte, user string, authResponse []byte, remoteAddr net.Addr) (mysql.Getter, error) {
	return &mysql.StaticUserData{}, mysql.NewSQLError(mysql.ERAccessDeniedError, mysql.SSAccessDeniedError, "Access denied for user '%v'", user)
}

func (a Auth) AuthMethod(user, addr string) (string, error) {
	return mysql.MysqlClearPassword, nil
}

func (a Auth) Negotiate(c *mysql.Conn, user string, remoteAddr net.Addr) (mysql.Getter, error) {
	// Finish the negotiation.
	password, err := mysql.AuthServerNegotiateClearOrDialog(c, mysql.MysqlClearPassword)
	if err != nil {
		return nil, err
	}
	err = a.DB.View(func(txn *badger.Txn) error {
		key := keys.Account(user)
		it, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return it.Value(func(val []byte) error {
			var a v1.Account
			err := proto.Unmarshal(val, &a)
			if err != nil {
				return err
			}
			return bcrypt.CompareHashAndPassword(a.Password, []byte(password))
		})
	})
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("Authenticate mysql client", "conn", c.ID(), "err", err, "user", user)
		}
		return &mysql.StaticUserData{}, mysql.NewSQLError(mysql.ERAccessDeniedError, mysql.SSAccessDeniedError, "Access denied for user '%v'", user)
	}
	return StaticUserData(user), nil
}

type StaticUserData string

// Get returns the wrapped username and groups
func (sud StaticUserData) Get() *querypb.VTGateCallerID {
	return &querypb.VTGateCallerID{Username: string(sud)}
}
