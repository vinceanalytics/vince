package a2

import (
	"errors"
	"testing"
)

func TestServerErrorLogger(t *testing.T) {
	sconfig := NewServerConfig()
	server := NewServer(sconfig, NewTestingStorage())

	r := server.NewResponse()
	r.ErrorStatusCode = 404

	server.setErrorAndLog(r, E_INVALID_GRANT, errors.New("foo"), "", "")

	if r.ErrorId != E_INVALID_GRANT {
		t.Errorf("expected error to be set to %s", E_INVALID_GRANT)
	}
	if r.StatusText != deferror.Get(E_INVALID_GRANT) {
		t.Errorf("expected status text to be %s, got %s", deferror.Get(E_INVALID_GRANT), r.StatusText)
	}

}
