package types

import (
	"golang.org/x/crypto/bcrypt"
)

// User represents system's user
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
	UserID   int64  `json:"-" pg:",fk" sql:"on_delete:CASCADE"`
	User     *User  `json:"-"`
}

// BranchConfig sets CI configuration for specific branch of repo
type BranchConfig struct {
	ID           int64       `json:"-"`
	GithubRepoID int64       `json:"-" pg:",fk" sql:"on_delete:CASCADE"`
	GithubRepo   *GithubRepo `json:"-"`
	Branch       string      `json:"branch"`
	IsCiEnabled  bool        `json:"ci_enabled"`
}

// Authenticate checks if provided password matches
// for this user
func (u User) Authenticate(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// CredentialsMessage is used for
// json binding
type CredentialsMessage struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// WebhookSecretMessage is used for
// json binding
type WebhookSecretMessage struct {
	Secret string `json:"secret"`
}

// RepoMessage is used for
// json binding
type RepoMessage struct {
	FullName string `json:"fullName"`
	ID       int64  `json:"repoID"`
}

// BranchMessage is used for
// json binding
type BranchMessage struct {
	BranchName string `json:"branch"`
}

type RepositoryMessage struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
}

type HeadMessage struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

// WebhookPayload is used for
// json binding of webhook payload
type WebhookMessage struct {
	Action     string            `json:"action"`
	Repository RepositoryMessage `json:"repository"`
	Ref        string            `json:"ref"`
	Head       HeadMessage       `json:"head"`
	GitUrl     string            `json:"clone_url"`
}
