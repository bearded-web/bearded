package passlib

import "golang.org/x/crypto/bcrypt"

const bcryptName = "bcrypt"

type BcryptAlgo struct {
	cost int
}

func NewBcrypt(cost int) *BcryptAlgo {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptAlgo{cost}
}

func (a *BcryptAlgo) Name() string {
	return bcryptName
}

func (a *BcryptAlgo) Encrypt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), a.cost)
	if err != nil {
		return "", err
	} else {
		return string(hash), nil
	}
}

func (a *BcryptAlgo) Verify(password, hash string) (bool, error) {
	h := []byte(hash)
	err := bcrypt.CompareHashAndPassword(h, []byte(password))
	verified := err == nil
	if err == bcrypt.ErrMismatchedHashAndPassword {
		err = nil
	}
	return verified, err
}
