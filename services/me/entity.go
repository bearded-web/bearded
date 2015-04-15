package me

type ChangePasswordEntity struct {
	Old string `json:"old"`
	New string `json:"new"`
}
