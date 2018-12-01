package db

import (
	"net/url"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/revan730/clipper-api/types"
	"golang.org/x/crypto/bcrypt"
)

// PostgresClient provides data access layer to objects in Postgres.
// implements DatabaseClient interface
type PostgresClient struct {
	pg         *pg.DB
	adminLogin string
	adminPass  string
}

// NewPGClient creates new copy of PostgresClient
func NewPGClient(config types.PGClientConfig) *PostgresClient {
	DBClient := &PostgresClient{
		adminLogin: config.AdminLogin,
		adminPass:  config.AdminPassword,
	}
	pgdb := pg.Connect(&pg.Options{
		User:         config.DBUser,
		Addr:         config.DBAddr,
		Password:     config.DBPassword,
		Database:     config.DB,
		MinIdleConns: 2,
	})
	DBClient.pg = pgdb
	return DBClient
}

// Close gracefully closes db connection
func (d *PostgresClient) Close() {
	d.pg.Close()
}

func (d *PostgresClient) createFirstAdmin() error {
	// Check if first admin exists
	user, err := d.FindUser(d.adminLogin)
	if err != nil {
		return err
	}
	if user != nil {
		return nil
	}
	return d.CreateUser(d.adminLogin, d.adminPass, true)
}

// CreateSchema creates database tables if they not exist
func (d *PostgresClient) CreateSchema() error {
	for _, model := range []interface{}{(*types.User)(nil),
		(*types.GithubRepo)(nil),
		(*types.BranchConfig)(nil),
	} {
		err := d.pg.CreateTable(model, &orm.CreateTableOptions{
			IfNotExists:   true,
			FKConstraints: true,
		})
		if err != nil {
			return err
		}
	}
	// Create default admin user
	return d.createFirstAdmin()
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// CreateUser creates new user with provided credentials and admin status
func (d *PostgresClient) CreateUser(login, pass string, isAdmin bool) error {
	hash, err := hashPassword(pass)
	if err != nil {
		return err
	}
	user := &types.User{
		Login:    login,
		Password: hash,
		IsAdmin:  isAdmin,
	}

	return d.pg.Insert(user)
}

// SaveUser writes provided user to db
func (d *PostgresClient) SaveUser(user *types.User) error {
	return d.pg.Update(user)
}

// FindUser returns user struct with provided login if it exists
func (d *PostgresClient) FindUser(login string) (*types.User, error) {
	user := &types.User{
		Login: login,
	}

	err := d.pg.Model(user).
		Where("login = ?", login).
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// FindUserByID returns user struct with provided id if it exists
func (d *PostgresClient) FindUserByID(userID int64) (*types.User, error) {
	user := &types.User{
		ID: userID,
	}

	err := d.pg.Select(user)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// CreateRepo creates github repo record with provided full name and owner id
func (d *PostgresClient) CreateRepo(fullName string, userID int64) error {
	repo := &types.GithubRepo{
		FullName: fullName,
		UserID:   userID,
	}

	return d.pg.Insert(repo)
}

// SaveRepo writes provided github repo to db
func (d *PostgresClient) SaveRepo(repo *types.GithubRepo) error {
	return d.pg.Update(repo)
}

// FindRepoByName returns repo struct with provided full name if it exists
func (d *PostgresClient) FindRepoByName(fullName string) (*types.GithubRepo, error) {
	repo := &types.GithubRepo{
		FullName: fullName,
	}

	err := d.pg.Model(repo).
		Where("full_name = ?", fullName).
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return repo, nil
}

// FindRepoByID returns repo struct with provided id if it exists
func (d *PostgresClient) FindRepoByID(repoID int64) (*types.GithubRepo, error) {
	repo := &types.GithubRepo{
		ID: repoID,
	}

	err := d.pg.Select(repo)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return repo, nil
}

// DeleteRepoByID deletes repo record with provided id if it exists
func (d *PostgresClient) DeleteRepoByID(repoID int64) error {
	repo := &types.GithubRepo{
		ID: repoID,
	}

	return d.pg.Delete(repo)
}

// FindAllUserRepos returns all repo records for provided user
// with pagination support (by passing query params of request)
func (d *PostgresClient) FindAllUserRepos(userID int64, q url.Values) ([]types.GithubRepo, error) {
	var repos []types.GithubRepo

	err := d.pg.Model(&repos).
		Apply(orm.Pagination(q)).
		Where("user_id = ?", userID).
		Select()

	return repos, err
}

// CreateBranchConfig creates repo branch config from provided struct
func (d *PostgresClient) CreateBranchConfig(c *types.BranchConfig) error {
	return d.pg.Insert(c)
}

// FindBranchConfig returns repo branch config with provided
// repo id and branch name if it exists
func (d *PostgresClient) FindBranchConfig(repoID int64, branch string) (*types.BranchConfig, error) {
	var c types.BranchConfig
	err := d.pg.Model(&c).
		Where("github_repo_id = ?", repoID).
		Where("branch = ?", branch).
		Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// DeleteBranchConfig deletes branch config record with provided repo id
// and branch name if it exists
func (d *PostgresClient) DeleteBranchConfig(repoID int64, branch string) error {
	config, err := d.FindBranchConfig(repoID, branch)
	if err != nil {
		return err
	}
	return d.pg.Delete(config)
}

// DeleteBranchConfigByID deletes branch config record with provided repo id
// if it exists
func (d *PostgresClient) DeleteBranchConfigByID(configID int64) error {
	c := &types.BranchConfig{
		ID: configID,
	}

	return d.pg.Delete(c)
}

// FindAllBranchConfigs returns all repo branch configs for provided repo id
// with pagination support (by passing query params of request)
func (d *PostgresClient) FindAllBranchConfigs(repoID int64, q url.Values) ([]types.BranchConfig, error) {
	var configs []types.BranchConfig

	err := d.pg.Model(&configs).
		Apply(orm.Pagination(q)).
		Where("github_repo_id = ?", repoID).
		Select()

	return configs, err
}
