package passlib

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcrypt(t *testing.T) {
	p := "password"

	b := NewBcrypt(0)
	if b == nil {
		t.Fatal("NewBcrypt returns nil")
	}
	if actual := b.Name(); actual != bcryptName {
		t.Errorf("bcrypt Name() %s != %s ", bcryptName, actual)
	}
	if b.cost != bcrypt.DefaultCost {
		t.Fatalf("bcrypt cost expected %d but actual is %d", bcrypt.DefaultCost, b.cost)
	}

	hash, err := b.Encrypt(p)
	if err != nil {
		t.Fatalf("Encrypt should return hash, but return error instead %v", err)
	}
	if len(hash) == 0 {
		t.Fatalf("Hash shouldn't be an empty string")
	}
	ok, err := b.Verify(p, hash)
	if err != nil {
		t.Fatalf("Verify returns error %v", err)
	}
	if !ok {
		t.Fatalf("Verify returns false, but true is expected")
	}

}

func TestBcryptEncrypt(t *testing.T) {
	b := NewBcrypt(0)
	if b == nil {
		t.Fatal("NewBcrypt returns nil")
	}
	//	hash, err := b.Encrypt("password")

}
