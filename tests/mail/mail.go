package mail

import (
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/emersion/go-smtp"
)

var _ smtp.Backend = (*Backend)(nil)
var _ smtp.Session = (*Session)(nil)

type Backend struct {
	*smtp.Server
	mu        sync.Mutex
	Msg       map[string]Message
	OnMessage func(*Message)
}

func New(addr string, f ...func(*smtp.Server)) *Backend {
	be := &Backend{Msg: make(map[string]Message)}
	be.Server = smtp.NewServer(be)
	be.Addr = addr
	be.Domain = "localhost"
	be.ReadTimeout = 10 * time.Second
	be.WriteTimeout = 10 * time.Second
	be.MaxMessageBytes = 1024 * 1024
	be.MaxRecipients = 50
	be.AllowInsecureAuth = true
	for _, fn := range f {
		fn(be.Server)
	}
	return be
}

func (b *Backend) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	return &Session{b: b}, nil
}

type Session struct {
	b   *Backend
	msg Message
}

func (s *Session) AuthPlain(username, password string) error {
	return nil
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.msg.From = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	s.msg.To = append(s.msg.To, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	s.msg.Data = string(b)
	s.b.mu.Lock()
	s.b.Msg[s.msg.From] = s.msg
	if s.b.OnMessage != nil {
		s.b.OnMessage(&s.msg)
	}
	s.b.mu.Unlock()
	s.msg = Message{}
	return nil
}

func (s *Session) Reset() {
	s.msg = Message{}
}

func (s *Session) Logout() error {
	return nil
}

type Message struct {
	From string
	To   []string
	Data string
}
