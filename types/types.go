package types

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id            int64  `json:"-"`
	Login         string `sql:",unique" json:"-"`
	Password      string `json:"-"`
	IsAdmin       bool   `json:"-" sql:"default:false"`
	WebhookSecret string `json:"-" sql:"default:''"`
}

func (u User) Authenticate(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

type CredentialsMessage struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type WebhookSecretMessage struct {
	Secret string `json:"secret"`
}
