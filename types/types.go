package types

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

// AccessTokenMessage is used for
// json binding
type AccessTokenMessage struct {
	Token string `json:"token"`
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
	SHA string `json:"sha"`
}

type CommitMessage struct {
	SHA string `json:"id"`
}

// WebhookMessage is used for
// json binding of webhook payload
type WebhookMessage struct {
	Action     string            `json:"action"`
	Repository RepositoryMessage `json:"repository"`
	Ref        string            `json:"ref"`
	Head       HeadMessage       `json:"head"`
	GitURL     string            `json:"clone_url"`
	HeadCommit CommitMessage     `json:"head_commit"`
}
