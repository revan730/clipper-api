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
	FindAllUsers(q url.Values) ([]types.User, error)
	FindAllUsersCount() (int64, error)
	ChangeUserAdminStatus(userID int64, isAdmin bool) error
	CreateRepo(fullName string, userID int64) error
	SaveRepo(repo *types.GithubRepo) error
	FindRepoByName(fullName string) (*types.GithubRepo, error)
	FindRepoByID(repoID int64) (*types.GithubRepo, error)
	DeleteRepoByID(repoID int64) error
	FindAllUserRepos(userID int64, q url.Values) ([]types.GithubRepo, error)
	FindAllUserReposCount(userID int64) (int64, error)
	FindAllRepos(q url.Values) ([]types.GithubRepo, error)
	FindAllReposCount() (int64, error)
	CreateBranchConfig(c *types.BranchConfig) error
	FindBranchConfig(repoID int64, branch string) (*types.BranchConfig, error)
	DeleteBranchConfig(repoID int64, branch string) error
	DeleteBranchConfigByID(configID int64) error
	FindAllBranchConfigs(repoID int64, q url.Values) ([]types.BranchConfig, error)
	FindAllBranchConfigsCount(repoID int64) (int64, error)
}
