package postgres

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/bsc-bridge-svc/internal/config"
	"github.com/bsc-bridge-svc/internal/data"
	"github.com/pkg/errors"
)

// Users interface, which defines the main functions to query the underlying postgres database
type Users interface {
	New() Users
	Get() (*data.User, error)
	GetUser(address, denom string) (*data.User, error)
	GetUserById(id int64) (*data.User, error)
	CreateUser(user data.User) error
	UpdateUser(user data.User) error
	DeleteUser(address string) error
}

type users struct {
	db  *sql.DB
	sql sq.SelectBuilder
}

const (
	all        = "*"
	usersTable = "users"
)

var usersSelect = sq.Select(all).From(usersTable).PlaceholderFormat(sq.Dollar)

func NewUsers(cfg config.Config) Users {
	return &users{
		db:  cfg.DB(),
		sql: usersSelect.RunWith(cfg.DB()),
	}
}

func (us *users) New() Users {
	return &users{
		db:  us.db,
		sql: usersSelect.RunWith(us.db),
	}
}

func (us *users) Get() (*data.User, error) {
	rowScanner := us.sql.QueryRow()
	user := data.User{}
	err := rowScanner.Scan(
		&user.ID,
		&user.Address,
		&user.Amount,
		&user.Denom,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "failed to query user")
	} else if err == sql.ErrNoRows {
		return nil, nil
	}
	return &user, nil
}

func (us *users) GetUser(address, denom string) (*data.User, error) {
	us.sql = us.sql.Where(sq.Eq{"address": address, "denom": denom})
	return us.Get()
}

func (us *users) GetUserById(id int64) (*data.User, error) {
	us.sql = us.sql.Where(sq.Eq{"id": id})
	return us.Get()
}

func (us *users) newInsert() sq.InsertBuilder {
	return sq.Insert(usersTable).RunWith(us.db).PlaceholderFormat(sq.Dollar)
}

func (us *users) CreateUser(user data.User) error {
	_, err := us.newInsert().SetMap(user.ToMap()).Exec()
	if err != nil {
		return errors.Wrap(err, "failed to insert user")
	}
	return nil
}

func (us *users) newUpdate() sq.UpdateBuilder {
	return sq.Update(usersTable).RunWith(us.db).PlaceholderFormat(sq.Dollar)
}

func (us *users) UpdateUser(user data.User) error {
	_, err := us.newUpdate().SetMap(user.ToMap()).Where(sq.Eq{"address": user.Address}).Exec()
	if err != nil {
		return errors.Wrap(err, "failed to update user data")
	}
	return nil
}

func (us *users) newDelete() sq.DeleteBuilder {
	return sq.Delete(usersTable).RunWith(us.db).PlaceholderFormat(sq.Dollar)
}

func (us *users) DeleteUser(address string) error {
	_, err := us.newDelete().Where(sq.Eq{"address": address}).Exec()
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}
	return nil
}
