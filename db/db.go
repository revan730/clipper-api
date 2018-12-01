package db

import (
	"net/url"

	"github.com/revan730/clipper-api/types"
)

// DatabaseClient provides interface for data access layer operations
type DatabaseClient interface {
	Close()
	CreateSchema() error
	CreateUser(login, pass string, isAdmin bool) error
	SaveUser(user *types.User) error
	FindUser(login string) (*types.User, error)
	FindUserByID(userID int64) (*types.User, error)
	CreateRepo(fullName string, userID int64) error
	SaveRepo(repo *types.GithubRepo) error
	FindRepoByName(fullName string) (*types.GithubRepo, error)
	FindRepoByID(repoID int64) (*types.GithubRepo, error)
	DeleteRepoByID(repoID int64) error
	FindAllUserRepos(userID int64, q url.Values) ([]types.GithubRepo, error)
	CreateBranchConfig(c *types.BranchConfig) error
	FindBranchConfig(repoID int64, branch string) (*types.BranchConfig, error)
	DeleteBranchConfig(repoID int64, branch string) error
	DeleteBranchConfigByID(configID int64) error
	FindAllBranchConfigs(repoID int64, q url.Values) ([]types.BranchConfig, error)
}
