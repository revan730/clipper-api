package db

import (
	"net/url"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/revan730/diploma-server/types"
	"golang.org/x/crypto/bcrypt"
)

type DatabaseClient struct {
	pg         *pg.DB
	adminLogin string
	adminPass  string
}

func NewDBClient(config types.Config) *DatabaseClient {
	DBClient := &DatabaseClient{
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

func (d *DatabaseClient) Close() {
	d.pg.Close()
}

func (d *DatabaseClient) createFirstAdmin() error {
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
func (d *DatabaseClient) CreateSchema() error {
	for _, model := range []interface{}{(*types.User)(nil),
		(*types.GithubRepo)(nil),
		(*types.BranchConfig)(nil)} {
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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func (d *DatabaseClient) CreateUser(login, pass string, isAdmin bool) error {
	hash, err := HashPassword(pass)
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

func (d *DatabaseClient) SaveUser(user *types.User) error {
	return d.pg.Update(user)
}

func (d *DatabaseClient) FindUser(login string) (*types.User, error) {
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
	} else {
		return user, nil
	}
}

func (d *DatabaseClient) FindUserById(userId int64) (*types.User, error) {
	user := &types.User{
		ID: userId,
	}

	err := d.pg.Select(user)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	} else {
		return user, nil
	}
}

func (d *DatabaseClient) CreateRepo(fullName string, userID int64) error {
	repo := &types.GithubRepo{
		FullName: fullName,
		UserID:   userID,
	}

	return d.pg.Insert(repo)
}

func (d *DatabaseClient) SaveRepo(repo *types.GithubRepo) error {
	return d.pg.Update(repo)
}

func (d *DatabaseClient) FindRepoByName(fullName string) (*types.GithubRepo, error) {
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
	} else {
		return repo, nil
	}
}

func (d *DatabaseClient) FindRepoByID(repoID int64) (*types.GithubRepo, error) {
	repo := &types.GithubRepo{
		ID: repoID,
	}

	err := d.pg.Select(repo)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, err
	} else {
		return repo, nil
	}
}

func (d *DatabaseClient) DeleteRepoByID(repoID int64) error {
	repo := &types.GithubRepo{
		ID: repoID,
	}

	return d.pg.Delete(repo)
}

func (d *DatabaseClient) FindAllUserRepos(userID int64, q url.Values) ([]types.GithubRepo, error) {
	var repos []types.GithubRepo

	err := d.pg.Model(&repos).
		Apply(orm.Pagination(q)).
		Where("user_id = ?", userID).
		Select()

	return repos, err
}

func (d *DatabaseClient) CreateBranchConfig(c *types.BranchConfig) error {
	return d.pg.Insert(c)
}

func (d *DatabaseClient) FindBranchConfig(repoID int64, branch string) (*types.BranchConfig, error) {
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
	} else {
		return &c, nil
	}
}

func (d *DatabaseClient) DeleteBranchConfig(repoID int64, branch string) error {
	config, err := d.FindBranchConfig(repoID, branch)
	if err != nil {
		return err
	}
	return d.pg.Delete(config)
}

func (d *DatabaseClient) DeleteBranchConfigByID(configID int64) error {
	c := &types.BranchConfig{
		ID: configID,
	}

	return d.pg.Delete(c)
}

func (d *DatabaseClient) FindAllBranchConfigs(repoID int64, q url.Values) ([]types.BranchConfig, error) {
	var configs []types.BranchConfig

	err := d.pg.Model(&configs).
		Apply(orm.Pagination(q)).
		Where("github_repo_id = ?", repoID).
		Select()

	return configs, err
}
