package passlib
// passlib is a password hashing library for golang, which provides cross-platform implementations
// of over 3 password hashing algorithms, as well as a framework for managing existing password hashes.
// It’s designed to be useful for a wide range of tasks, from verifying a hash found in /etc/shadow,
// to providing full-strength password hashing for multi-user application.
//
// Inspired by great python library passlib http://pythonhosted.org/passlib/

import (
	"fmt"
	"strings"
)


const (
	Divider  string = "$"
	DividerB byte   = 0x24
)

type Algo interface {
	Name() string // unique algorithm name
	Encrypt(password string) (string, error)
	Verify(password, hash string) (bool, error)
	//	Compatible(hash string) bool // is hash compatible with this algo?
}

type Context struct {
	Default Algo
}

func NewContext() *Context{
	return &Context{
		Default: NewBcrypt(0),
	}
}

type EncOpts struct {
}

type VerOpts struct {
}

func (c *Context) Encrypt(password string, _ ...EncOpts) (string, error) {
	algo := c.Default
	hash, err := algo.Encrypt(password)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s%s", Divider, algo.Name(), hash), nil
}

func (c *Context) Verify(password string, hash string, _ ...VerOpts) (bool, error) {
	algo := c.Default
	if strings.HasPrefix(hash, Divider) {
		hash = hash[1:]
	}
	parts := strings.SplitN(hash, Divider, 2)
	if len(parts) < 2 {
		return false, fmt.Errorf("Bad hash format")
	}
	if parts[0] != algo.Name() {
		return false, fmt.Errorf("Unsupported algorithm: %s", parts[0])
	}
	return c.Default.Verify(password, fmt.Sprintf("$%s", parts[1]))
}
