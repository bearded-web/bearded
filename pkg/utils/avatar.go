package utils

import (
	"crypto/md5"
	"fmt"
)

type AvatarType string

const (
	//	a simple, cartoon-style silhouetted outline of a person (does not vary by email hash)
	AvatarMysteryMan AvatarType = "mm"
	//	a geometric pattern based on an email hash
	AvatarIdenticon AvatarType = "identicon"
	//	a generated 'monster' with different colors, faces, etc
	AvatarMonster AvatarType = "monsterid"
	//	generated faces with differing features and backgrounds
	AvatarWavatar AvatarType = "wavatar"
	//	awesome generated, 8-bit arcade-style pixelated faces
	AvatarRetro AvatarType = "retro"
	//	do not load any image if none is associated with the email hash, instead return an HTTP 404 (File Not Found) response
	Avatar404 AvatarType = "404"
	//	a transparent PNG image (border added to HTML below for demonstration purposes)
	Avatar AvatarType = "blank"
)

// Generate gravatar url
// Read more about gravatar: http://en.gravatar.com/site/implement/images/
func GetGravatar(id string, size int, dtype AvatarType) string {
	h := md5.New()
	h.Write([]byte(id))
	sum := h.Sum(nil)
	return fmt.Sprintf("https://secure.gravatar.com/avatar/%x?s=%d&d=%s", sum, size, dtype)
}
