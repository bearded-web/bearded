package me

import "github.com/bearded-web/bearded/models/user"


type Info struct {
	User	*user.User `json:"user"`
}
