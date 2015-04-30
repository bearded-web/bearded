package email

import (
	"fmt"

	"github.com/bearded-web/bearded/pkg/config"
	"gopkg.in/gomail.v1"
)

type BackendType string

const (
	SmtpType    = BackendType("smtp")
	ConsoleType = BackendType("console")
)

var (
	ErrUnknownBackend = fmt.Errorf("unknown backend")
)

var Backends = map[BackendType]struct{}{
	SmtpType:    struct{}{},
	ConsoleType: struct{}{},
}

type Mailer interface {
	Send(*gomail.Message) error
	Close()
}

type Email struct {
	cfg     *config.Email
	backend Mailer
}

func New(cfg config.Email) (*Email, error) {
	e := &Email{}
	err := e.SetConfig(cfg)
	return e, err
}

func (e *Email) Send(msg *gomail.Message) error {
	return e.backend.Send(msg)
}

func (e *Email) Close() {
	e.backend.Close()
}

func (e *Email) SetConfig(cfg config.Email) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}
	e.cfg = &cfg
	if e.backend != nil {
		e.backend.Close()
	}
	switch BackendType(cfg.Backend) {
	case SmtpType:
		e.backend = NewSmtpBackend(cfg.Smtp)
	case ConsoleType:
		e.backend = NewConsoleBackend()
	default:
		return ErrUnknownBackend
	}
	return nil
}

func validateConfig(cfg config.Email) error {
	backend := BackendType(cfg.Backend)
	_, ok := Backends[backend]
	if !ok {
		return ErrUnknownBackend
	}
	switch backend {
	case SmtpType:
		return ValidateSmtpConfig(cfg.Smtp)
	}
	return nil
}

func NewMessage() *gomail.Message {
	return gomail.NewMessage()
}
