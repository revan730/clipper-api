package types

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            int64  `json:"-"`
	Login         string `sql:",unique" json:"-"`
	Password      string `json:"-"`
	IsAdmin       bool   `json:"-" sql:"default:false"`
	WebhookSecret string `json:"-" sql:"default:''"`
}

// GithubRepo represents GitHub repository
type GithubRepo struct {
	ID       int64  `json:"repoID"`
	FullName string `json:"fullName" sql:",unique"`
	OwnerID  int64  `json:"-"`
}

// BranchConfig sets CI configuration for specific branch of repo
type BranchConfig struct {
	ID          int64  `json:"-"`
	RepoID      int64  `json:"-"`
	Branch      string `json:"branch"`
	IsCiEnabled bool   `json:"ci_enabled"`
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

type RepoMessage struct {
	FullName string `json:"fullName"`
	ID       int64  `json:"repoID"`
}

type BranchMessage struct {
	BranchName string `json:"branch"`
}
