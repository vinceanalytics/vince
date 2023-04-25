package email

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/render"
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

func FromConfig(conf *config.Config) (Mailer, error) {
	s := &SMTP{
		address: fmt.Sprintf("%s:%d", conf.Mailer.Smtp.Host, conf.Mailer.Smtp.Port),
		from: &mail.Address{
			Name:    conf.Mailer.Address.Name,
			Address: conf.Mailer.Address.Email,
		},
	}
	c, err := smtp.Dial(s.address)
	if err != nil {
		return nil, err
	}
	c.Close()
	if conf.Mailer.Smtp.Auth != nil {
		switch a := conf.Mailer.Smtp.Auth.(type) {
		case *config.Config_Mailer_Smtp_Anonymous:
			s.auth = sasl.NewAnonymousClient(a.Anonymous.Trace)
		case *config.Config_Mailer_Smtp_OauthBearer:
			s.auth = sasl.NewOAuthBearerClient(&sasl.OAuthBearerOptions{
				Username: a.OauthBearer.Username,
				Token:    a.OauthBearer.Token,
				Host:     a.OauthBearer.Host,
				Port:     int(a.OauthBearer.Port),
			})
		case *config.Config_Mailer_Smtp_Plain:
			s.auth = sasl.NewPlainClient(
				a.Plain.Identity, a.Plain.Username, a.Plain.Password,
			)
		}
	}
	if conf.Env == config.Config_dev {
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
