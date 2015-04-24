// Package passlib.reset implements creation and verification of secure tokens
// useful for implementation of "reset forgotten password" feature in web
// applications.
//
// This package generates and verifies signed one-time tokens that can be
// embedded in a link sent to users when they initiate the password reset
// procedure. When a user changes their password, or when the expiry time
// passes, the token becomes invalid.
//
package reset

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"time"
)

const (
	decodedMinLength = 4 /*expiration*/ + 1 /*login*/ + 32 /*signature*/
	decodedMaxLength = 1024                                /* maximum decoded length, for safety */

	hashLen = 32
	tsLen   = 4
)

var (
	ErrMalformedToken = errors.New("malformed token")
	ErrExpiredToken   = errors.New("token expired")
	ErrWrongSignature = errors.New("wrong token signature")
)

// It is useful for avoiding DoS attacks with too long tokens: before passing
// a token to Verify function, check that it has length less than the
// [maximum login length allowed in your application] + MinLength.
var MinLength = base64.URLEncoding.EncodedLen(decodedMinLength)

func getUserSecretKey(pwdval, secret []byte) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write(pwdval)
	return m.Sum(nil)
}

func getSignature(b []byte, secret []byte) []byte {
	keym := hmac.New(sha256.New, secret)
	keym.Write(b)
	m := hmac.New(sha256.New, keym.Sum(nil))
	m.Write(b)
	return m.Sum(nil)
}

// NewToken returns a new password reset token for the given login, which
// expires after the given time duration since now, signed by the key generated
// from the given password value (which can be any value that will be changed
// once a user resets their password, such as password hash or salt used to
// generate it), and the given secret key.
func NewToken(user string, dur time.Duration, pwdHash, secret []byte) string {
	sk := getUserSecretKey(pwdHash, secret)
	return newTokenSinceNow(user, dur, sk)
}

// New returns a signed authentication cookie for the given login,
// expiration time, and secret key.
// If the login is empty, the function returns an empty string.
func newToken(user string, expires time.Time, secret []byte) string {
	if user == "" {
		return ""
	}
	llen := len(user)
	b := make([]byte, llen+tsLen+hashLen)
	// Put expiration time.
	binary.BigEndian.PutUint32(b, uint32(expires.Unix()))
	// Put login.
	copy(b[tsLen:], []byte(user))
	// Calculate and put signature.
	sig := getSignature([]byte(b[:tsLen+llen]), secret)
	copy(b[tsLen+llen:], sig)
	// Base64-encode.
	return base64.URLEncoding.EncodeToString(b)
}

// NewSinceNow returns a signed authentication token for the given login,
// duration since current time, and secret key.
func newTokenSinceNow(user string, dur time.Duration, secret []byte) string {
	return newToken(user, time.Now().Add(dur), secret)
}

// VerifyToken verifies the given token with the password value returned by the
// given function and the given secret key, and returns login extracted from
// the valid token. If the token is not valid, the function returns an error.
//
// Function pwdvalFn must return the current password value for the login it
// receives in arguments, or an error. If it returns an error, VerifyToken
// returns the same error.
func VerifyToken(token string, pwdvalFn func(string) ([]byte, error), secret []byte) (string, error) {
	blen := base64.URLEncoding.DecodedLen(len(token))
	// Avoid allocation if the token is too short
	if blen <= tsLen+hashLen {
		return "", ErrMalformedToken
	}
	b := make([]byte, blen)
	blen, err := base64.URLEncoding.Decode(b, []byte(token))
	if err != nil {
		return "", err
	}
	// Decoded length may be different from max length, which
	// we allocated, so check it, and set new length for b
	if blen <= tsLen+hashLen {
		return "", ErrMalformedToken
	}
	b = b[:blen]

	data := b[:blen-hashLen]
	exp := time.Unix(int64(binary.BigEndian.Uint32(data[:tsLen])), 0)
	if exp.Before(time.Now()) {
		return "", ErrExpiredToken
	}
	user := string(data[tsLen:])
	pwdval, err := pwdvalFn(user)
	if err != nil {
		return "", err
	}
	sig := b[blen-hashLen:]
	sk := getUserSecretKey(pwdval, secret)
	realSig := getSignature(data, sk)
	if subtle.ConstantTimeCompare(realSig, sig) != 1 {
		return "", ErrWrongSignature
	}
	return user, nil
}
