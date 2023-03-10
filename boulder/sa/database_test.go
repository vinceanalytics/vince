package sa

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gernest/vince/boulder/test"
	"github.com/gernest/vince/boulder/test/vars"
)

func TestInvalidDSN(t *testing.T) {
	_, err := NewDbMap("invalid", DbSettings{})
	test.AssertError(t, err, "DB connect string missing the slash separating the database name")
}

var errExpected = errors.New("expected")

func TestDbSettings(t *testing.T) {
	// TODO(#5248): Add a full db.mockWrappedMap to sa/database tests
	oldSetMaxOpenConns := setMaxOpenConns
	oldSetMaxIdleConns := setMaxIdleConns
	oldSetConnMaxLifetime := setConnMaxLifetime
	oldSetConnMaxIdleTime := setConnMaxIdleTime
	defer func() {
		setMaxOpenConns = oldSetMaxOpenConns
		setMaxIdleConns = oldSetMaxIdleConns
		setConnMaxLifetime = oldSetConnMaxLifetime
		setConnMaxIdleTime = oldSetConnMaxIdleTime
	}()

	maxOpenConns := -1
	maxIdleConns := -1
	connMaxLifetime := time.Second * 1
	connMaxIdleTime := time.Second * 1

	setMaxOpenConns = func(db *sql.DB, m int) {
		maxOpenConns = m
		oldSetMaxOpenConns(db, maxOpenConns)
	}
	setMaxIdleConns = func(db *sql.DB, m int) {
		maxIdleConns = m
		oldSetMaxIdleConns(db, maxIdleConns)
	}
	setConnMaxLifetime = func(db *sql.DB, c time.Duration) {
		connMaxLifetime = c
		oldSetConnMaxLifetime(db, connMaxLifetime)
	}
	setConnMaxIdleTime = func(db *sql.DB, c time.Duration) {
		connMaxIdleTime = c
		oldSetConnMaxIdleTime(db, connMaxIdleTime)
	}
	dbSettings := DbSettings{
		MaxOpenConns:    100,
		MaxIdleConns:    100,
		ConnMaxLifetime: 100,
		ConnMaxIdleTime: 100,
	}
	_, err := NewDbMap("sa@tcp(boulder-mysql:3306)/boulder_sa_integration", dbSettings)
	if err != nil {
		t.Errorf("connecting to DB: %s", err)
	}
	if maxOpenConns != 100 {
		t.Errorf("maxOpenConns was not set: expected %d, got %d", 100, maxOpenConns)
	}
	if maxIdleConns != 100 {
		t.Errorf("maxIdleConns was not set: expected %d, got %d", 100, maxIdleConns)
	}
	if connMaxLifetime != 100 {
		t.Errorf("connMaxLifetime was not set: expected %d, got %d", 100, connMaxLifetime)
	}
	if connMaxIdleTime != 100 {
		t.Errorf("connMaxIdleTime was not set: expected %d, got %d", 100, connMaxIdleTime)
	}
}

func TestNewDbMap(t *testing.T) {
	const mysqlConnectURL = "policy:password@tcp(boulder-mysql:3306)/boulder_policy_integration?readTimeout=800ms&writeTimeout=800ms"
	const expected = "policy:password@tcp(boulder-mysql:3306)/boulder_policy_integration?clientFoundRows=true&parseTime=true&readTimeout=800ms&writeTimeout=800ms&long_query_time=0.6400000000000001&max_statement_time=0.76&sql_mode=%27STRICT_ALL_TABLES%27"
	oldSQLOpen := sqlOpen
	defer func() {
		sqlOpen = oldSQLOpen
	}()
	sqlOpen = func(dbType, connectString string) (*sql.DB, error) {
		if connectString != expected {
			t.Errorf("incorrect connection string mangling, want %#v, got %#v", expected, connectString)
		}
		return nil, errExpected
	}

	dbMap, err := NewDbMap(mysqlConnectURL, DbSettings{})
	if err != errExpected {
		t.Errorf("got incorrect error. Got %v, expected %v", err, errExpected)
	}
	if dbMap != nil {
		t.Errorf("expected nil, got %v", dbMap)
	}

}

func TestStrictness(t *testing.T) {
	dbMap, err := NewDbMap(vars.DBConnSA, DbSettings{1, 0, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	_, err = dbMap.Exec(`insert into orderToAuthz2 set
		orderID=999999999999999999999999999,
		authzID=999999999999999999999999999;`)
	if err == nil {
		t.Fatal("Expected error when providing out of range value, got none.")
	}
	if !strings.Contains(err.Error(), "Out of range value for column") {
		t.Fatalf("Got wrong type of error: %s", err)
	}
}
