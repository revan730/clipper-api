package types

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	commonTypes "github.com/revan730/clipper-common/types"
	"golang.org/x/crypto/bcrypt"
)

// User represents system's user
type User struct {
	ID            int64  `json:"-"`
	Login         string `sql:",unique" json:"-"`
	Password      string `json:"-"`
	IsAdmin       bool   `json:"-" sql:"default:false"`
	WebhookSecret string `json:"-" sql:"default:''"`
	AccessToken   string `json:"-" sql:"default:''"`
}

// Authenticate checks if provided password matches
// for this user
func (u User) Authenticate(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
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
	GitURL   string `json:"clone_url"`
}

type HeadMessage struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

type CommitMessage struct {
	SHA string `json:"id"`
}

type PullRequestMessage struct {
	Head HeadMessage `json:"head"`
}

// TODO: Refactor using nested struct declaration

// WebhookMessage is used for
// json binding of webhook payload
type WebhookMessage struct {
	Action      string             `json:"action"`
	Repository  RepositoryMessage  `json:"repository"`
	Ref         string             `json:"ref"`
	PullRequest PullRequestMessage `json:"pull_request"`
	HeadCommit  CommitMessage      `json:"head_commit"`
}

type BuildMessage struct {
	ID            int64     `json:"id"`
	GithubRepoID  int64     `json:"repoID"`
	IsSuccessfull bool      `json:"isSuccessfull"`
	Date          time.Time `json:"date"`
	Branch        string    `json:"branch"`
	Stdout        string    `json:"stdout"`
}

type BuildArrayMessage struct {
	Builds []*BuildMessage `json:"builds"`
}

func BuildMsgFromProto(b *commonTypes.Build) (*BuildMessage, error) {
	date, err := ptypes.Timestamp(b.Date)
	if err != nil {
		return nil, err
	}
	buildMsg := &BuildMessage{
		ID:            b.ID,
		GithubRepoID:  b.GithubRepoID,
		IsSuccessfull: b.IsSuccessfull,
		Date:          date,
		Branch:        b.Branch,
		Stdout:        b.Stdout,
	}
	return buildMsg, nil
}

func BuildArrayMsgFromProto(b *commonTypes.BuildsArray) (*BuildArrayMessage, error) {
	buildArray := &BuildArrayMessage{}

	for _, build := range b.Builds {
		buildMsg, err := BuildMsgFromProto(build)
		if err != nil {
			return nil, err
		}
		buildArray.Builds = append(buildArray.Builds, buildMsg)
	}

	return buildArray, nil
}

type PGClientConfig struct {
	DBAddr        string
	DB            string
	DBUser        string
	DBPassword    string
	AdminLogin    string
	AdminPassword string
}

type BuildsQueryParams struct {
	Branch string `form:"branch"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type DeploymentMessage struct {
	ID         int64  `json:"ID"`
	Branch     string `json:"branch"`
	RepoID     int64  `json:"repoID"`
	ArtifactID int64  `json:"artifactID"`
	K8SName    string `json:"k8sName"`
	Manifest   string `json:"manifest"`
	Replicas   int64  `json:"replicas"`
}

type ImageMessage struct {
	ImageID int64 `json:"imageID"`
}

func DeploymentMsgFromProto(kd *commonTypes.Deployment) *DeploymentMessage {
	return &DeploymentMessage{
		ID:         kd.ID,
		RepoID:     kd.RepoID,
		Branch:     kd.Branch,
		ArtifactID: kd.ArtifactID,
		K8SName:    kd.K8SName,
		Manifest:   kd.Manifest,
		Replicas:   kd.Replicas,
	}
}

func ProtoFromDeploymentMsg(d *DeploymentMessage) *commonTypes.Deployment {
	return &commonTypes.Deployment{
		ID:         d.ID,
		RepoID:     d.RepoID,
		Branch:     d.Branch,
		ArtifactID: d.ArtifactID,
		K8SName:    d.K8SName,
		Manifest:   d.Manifest,
		Replicas:   d.Replicas,
	}
}
