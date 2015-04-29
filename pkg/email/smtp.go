package email

import (
	"github.com/bearded-web/bearded/pkg/config"
	"gopkg.in/gomail.v1"
)

type SmtpBackend struct {
	cfg    config.Smtp
	mailer *gomail.Mailer
}

func NewSmtpBackend(cfg config.Smtp) *SmtpBackend {
	return &SmtpBackend{
		cfg:    cfg,
		mailer: gomail.NewMailer(cfg.Addr, cfg.User, cfg.Password, cfg.Port),
	}
}

func (b *SmtpBackend) Send(msg *gomail.Message) error {
	return b.mailer.Send(msg)
}

func (b *SmtpBackend) Close() {
}

func ValidateSmtpConfig(cfg config.Smtp) error {
	// TODO (m0sth8): validate smtp backend config
	return nil
}
