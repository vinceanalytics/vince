package be

import (
	"context"
	"errors"

	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/session"
	"github.com/pingcap/tidb/util/sqlexec"
)

func NewSession(store kv.Storage) (session.Session, error) {
	se, err := session.CreateSession(store)
	if err != nil {
		return nil, err
	}
	return se, nil
}

func Execute(ctx context.Context, se session.Session, sql string) (sqlexec.RecordSet, error) {
	stmts, err := se.Parse(ctx, sql)
	if err != nil {
		return nil, err
	}
	if len(stmts) > 1 {
		return nil, errors.New("only single statement is supported")
	}
	if !ast.IsReadOnly(stmts[0]) {
		return nil, errors.New("only read statement is supported")
	}
	return se.ExecuteStmt(ctx, stmts[0])
}
