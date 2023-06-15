package email

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/vinceanalytics/vince/internal/core"
)

var _ Mailer = (*MockMailer)(nil)

type MockMailer struct {
	from string
	to   []string
	bytes.Buffer
}

func (m *MockMailer) SendMail(from string, to []string, msg io.Reader) error {
	m.from = from
	m.to = append(m.to, to...)
	m.ReadFrom(msg)
	return nil
}

func (m *MockMailer) From() *mail.Address {
	return &mail.Address{
		Name:    "jane",
		Address: "jane@example.com",
	}
}

func (m *MockMailer) Close() error {
	return nil
}

func TestSend_simple(t *testing.T) {
	now, err := time.Parse(time.RFC822Z, time.RFC822Z)
	if err != nil {
		t.Fatal(err)
	}
	m := &MockMailer{}
	ctx := Set(context.Background(), m)
	ctx = core.SetNow(ctx, func() time.Time {
		return now
	})
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	Register(ctx, vm)
	b, _ := os.ReadFile("testdata/simple_email.js")
	r, err := vm.RunString(string(b))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := os.ReadFile("testdata/simple_email.txt")
	want = removeBoundary(want)
	got := removeBoundary(m.Bytes())
	if !bytes.Equal(want, got) {
		t.Fatal("mismatch on generated email")
	}
	if got, want := r.ToInteger(), http.StatusOK; got != int64(want) {
		t.Errorf("expected %d got %d", want, got)
	}
}
func TestSend_missing_mailer(t *testing.T) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	Register(context.TODO(), vm)
	b, _ := os.ReadFile("testdata/missing_mailer.js")
	_, err := vm.RunString(string(b))
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "true") {
		t.Errorf("expected %v to contain true", err)
	}
}

const bound = "boundary="

func removeBoundary(b []byte) []byte {
	x := boundary(b)
	return bytes.ReplaceAll(b, x, []byte{})
}
func boundary(b []byte) []byte {
	i := bytes.Index(b, []byte(bound))
	line, _, _ := bufio.NewReader(bytes.NewReader(b[i:])).ReadLine()
	return bytes.TrimSpace(line[len(bound):])
}
