package server

import (
	stderrors "errors"
	"io"

	"github.com/emersion/go-sasl"
	smtp "github.com/emersion/go-smtp"
)

func (a *App) startSMTP() error {
	server := smtp.NewServer(&smtpBackend{app: a})
	server.Addr = a.settings.SMTPAddr
	server.Domain = "nozomi-relay.local"
	server.AllowInsecureAuth = true
	server.MaxMessageBytes = 10 * 1024 * 1024
	server.MaxRecipients = 50
	return server.ListenAndServe()
}

type smtpBackend struct{ app *App }

func (b *smtpBackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &smtpSession{app: b.app}, nil
}

type smtpSession struct {
	app       *App
	auth      bool
	accountID int64
	from      string
	to        []string
}

func (s *smtpSession) AuthMechanisms() []string { return []string{sasl.Plain} }
func (s *smtpSession) Auth(mech string) (sasl.Server, error) {
	return sasl.NewPlainServer(func(identity, username, password string) error {
		var stored string
		var active int
		err := s.app.db.QueryRow(`select id, password, active from smtp_accounts where username=?`, username).Scan(&s.accountID, &stored, &active)
		if err != nil || stored != password || active != 1 {
			return stderrors.New("invalid username or password")
		}
		s.auth = true
		return nil
	}), nil
}
func (s *smtpSession) Mail(from string, opts *smtp.MailOptions) error {
	if !s.auth {
		return smtp.ErrAuthRequired
	}
	s.from = from
	return nil
}
func (s *smtpSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	if !s.auth {
		return smtp.ErrAuthRequired
	}
	s.to = append(s.to, to)
	return nil
}
func (s *smtpSession) Data(r io.Reader) error {
	if !s.auth {
		return smtp.ErrAuthRequired
	}
	rawBytes, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	input := parseMail(s.from, s.to, rawBytes)
	return s.app.processMail(input, s.accountID)
}
func (s *smtpSession) Reset()        { s.from = ""; s.to = nil }
func (s *smtpSession) Logout() error { return nil }
