package me

type ChangePasswordEntity struct {
	Token string `json:"token,omitempty" description:"reset password token"`
	Old   string `json:"old,omitempty"`
	New   string `json:"new"`
}
