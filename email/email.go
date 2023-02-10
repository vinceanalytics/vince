package email

import (
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/config"
)

func compose(out io.Writer, tpl *template.Template, ctx *templates.Context, from *mail.Address, subject string) error {
	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", []*mail.Address{from})
	h.SetAddressList("To", []*mail.Address{ctx.CurrentUser.Address()})
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
	err = tpl.Execute(out, ctx)
	if err != nil {
		return err
	}
	w.Close()
	mw.Close()
	return nil
}

type Mailer interface {
	SendMail(from string, to []string, msg io.Reader) error
	io.Closer
}

var _ Mailer = (*SMTP)(nil)

type SMTP struct {
	auth    sasl.Client
	address string
}

func (s *SMTP) SendMail(from string, to []string, msg io.Reader) error {
	return smtp.SendMail(s.address, s.auth, from, to, msg)
}

func (s *SMTP) Close() error {
	return nil
}

func FromConfig(conf *config.Config) (*SMTP, error) {
	s := &SMTP{
		address: fmt.Sprintf("%s:%d", conf.Mailer.Smtp.Host, conf.Mailer.Smtp.Port),
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
	return s, nil
}
