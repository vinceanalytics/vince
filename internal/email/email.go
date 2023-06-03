package email

import (
	"context"
	"html/template"
	"io"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

func Compose(ctx context.Context,
	out io.Writer, tpl *template.Template,
	from, to *mail.Address, subject string, f ...func(*templates.Context)) error {
	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", []*mail.Address{from})
	h.SetAddressList("To", []*mail.Address{to})
	h.SetSubject(subject)
	mw, err := mail.CreateWriter(out, h)
	if err != nil {
		return err
	}
	// Create a text part
	var th mail.InlineHeader
	th.Set("Content-Type", "text/html")
	w, err := mw.CreateSingleInline(th)
	if err != nil {
		return err
	}
	err = render.EXEC(ctx, w, tpl, f...)
	if err != nil {
		return err
	}
	w.Close()
	mw.Close()
	return nil
}

type Mailer interface {
	SendMail(from string, to []string, msg io.Reader) error
	From() *mail.Address
	io.Closer
}

var _ Mailer = (*SMTP)(nil)

type SMTP struct {
	auth    sasl.Client
	address string
	from    *mail.Address
}

func (s *SMTP) SendMail(from string, to []string, msg io.Reader) error {
	return smtp.SendMail(s.address, s.auth, from, to, msg)
}

func (s *SMTP) Close() error {
	return nil
}
func (s *SMTP) From() *mail.Address {
	return &mail.Address{
		Name:    s.from.Name,
		Address: s.from.Address,
	}
}

func FromConfig(o *config.Options) (Mailer, error) {
	s := &SMTP{
		address: o.Mailer.SMTP.Address,
		from: &mail.Address{
			Name:    o.Mailer.Name,
			Address: o.Mailer.Address,
		},
	}
	c, err := smtp.Dial(s.address)
	if err != nil {
		return nil, err
	}
	c.Close()
	if o.Mailer.SMTP.AuthAnonymous.Enabled {
		s.auth = sasl.NewAnonymousClient(o.Mailer.SMTP.AuthAnonymous.Trace)
	} else if o.Mailer.SMTP.AuthOAUTHBearer.Enabled {
		s.auth = sasl.NewOAuthBearerClient(&sasl.OAuthBearerOptions{
			Username: o.Mailer.SMTP.AuthOAUTHBearer.Username,
			Token:    o.Mailer.SMTP.AuthOAUTHBearer.Token,
			Host:     o.Mailer.SMTP.AuthOAUTHBearer.Host,
			Port:     o.Mailer.SMTP.AuthOAUTHBearer.Port,
		})
	} else if o.Mailer.SMTP.AuthPlain.Enabled {
		s.auth = sasl.NewPlainClient(
			o.Mailer.SMTP.AuthPlain.Identity,
			o.Mailer.SMTP.AuthPlain.Username,
			o.Mailer.SMTP.AuthPlain.Password,
		)
	}

	if o.Mailer.SMTP.EnableMailHog {
		return &MailHog{SMTP: s}, nil
	}
	return s, nil
}

type MailHog struct {
	*SMTP
}

func (m *MailHog) SendMail(from string, to []string, msg io.Reader) error {
	client, err := smtp.Dial(m.address)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.SendMail(from, to, msg)
}

type mailerKey struct{}

func Set(ctx context.Context, m Mailer) context.Context {
	return context.WithValue(ctx, mailerKey{}, m)
}

func Get(ctx context.Context) Mailer {
	return ctx.Value(mailerKey{}).(Mailer)
}
