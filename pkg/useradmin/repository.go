package useradmin

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

type UserRepository interface {
	StoreUser(context.Context, string) error
	FindUser(context.Context, string) (bool, error)
	StoreGroup(context.Context, string, []string) error
	FindGroup(context.Context, string) (bool, error)
	FetchGroupUsers(context.Context, string) ([]string, error)
	Purge(context.Context) error
}

func NewUserRepository(urn string) (UserRepository, error) {
	db, err := sql.Open("mysql", urn)
	if err != nil {
		return nil, errors.Wrap(err, "error opening DB")
	}
	return &userRepository{db}, nil
}

type userRepository struct {
	db *sql.DB
}

func (r *userRepository) StoreUser(ctx context.Context, username string) (err error) {
	tx, txerr := r.startTransaction(ctx)
	if txerr != nil {
		return errors.Wrap(txerr, "error starting store user transaction")
	}
	defer func(tx *sql.Tx) {
		r.completeTransaction(tx, err)
	}(tx)
	var result sql.Result
	result, err = r.db.ExecContext(ctx, `INSERT INTO users(name) VALUES ( ? )`, username)
	if err != nil {
		err = errors.Wrap(err, "error inserting user")
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		err = errors.Wrap(err, "error inserting user")
	}
	return err
}

func (r *userRepository) FindUser(ctx context.Context, username string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT count(name) FROM users where name = ?", username).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "error selecting user count")
	}
	return count > 0, nil
}

func (r *userRepository) StoreGroup(ctx context.Context, groupname string, usernames []string) (err error) {
	tx, txerr := r.startTransaction(ctx)
	if txerr != nil {
		return errors.Wrap(txerr, "error starting store group transaction")
	}
	defer func(tx *sql.Tx) {
		r.completeTransaction(tx, err)
	}(tx)
	var result sql.Result
	result, err = tx.ExecContext(ctx, `INSERT INTO usergroups(name) VALUES ( ? )`, groupname)
	if err != nil {
		err = errors.Wrap(err, "error inserting group")
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		err = errors.Wrap(err, "error inserting group")
		return err
	}
	for _, username := range usernames {
		result, err = tx.ExecContext(ctx, `INSERT INTO groupusers(groupname,username) VALUES ( ?, ? )`, groupname, username)
		if err != nil {
			err = errors.Wrap(err, "error inserting group user")
			return err
		}
		_, err = result.LastInsertId()
		if err != nil {
			err = errors.Wrap(err, "error inserting group user")
			return err
		}
	}
	return err
}

func (r *userRepository) FindGroup(ctx context.Context, groupname string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT count(name) FROM usergroups where name = ?", groupname).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "error selecting group count")
	}
	return count > 0, nil
}

func (r *userRepository) FetchGroupUsers(ctx context.Context, groupname string) ([]string, error) {
	results, err := r.db.Query("SELECT username FROM groupusers where groupname = ?", groupname)
	if err != nil {
		return nil, errors.Wrap(err, "error selecting group users")
	}
	users := make([]string, 0)
	for results.Next() {
		var user string
		err = results.Scan(&user)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning group users")
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *userRepository) Purge(ctx context.Context) (err error) {
	tx, txerr := r.startTransaction(ctx)
	if txerr != nil {
		return errors.Wrap(txerr, "error starting purge transaction")
	}
	defer func(tx *sql.Tx) {
		r.completeTransaction(tx, err)
	}(tx)
	_, err = tx.ExecContext(ctx, `DELETE FROM groupusers`)
	if err != nil {
		return errors.Wrap(err, "error deleting groupusers")
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM usergroups`)
	if err != nil {
		return errors.Wrap(err, "error deleting usergroups")
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM users`)
	if err != nil {
		return errors.Wrap(err, "error deleting users")
	}
	return err
}

func (r *userRepository) startTransaction(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *userRepository) completeTransaction(tx *sql.Tx, err error) error {
	if err == nil {
		if commitErr := tx.Commit(); commitErr != nil {
			err = errors.Wrap(commitErr, "error committing transaction")
		}
	} else {
		tx.Rollback()
	}
	return err
}
