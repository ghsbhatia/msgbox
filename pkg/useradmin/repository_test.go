package useradmin

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/matryer/is"
)

// Test executor for user repository
func TestUserRepository(t *testing.T) {
	s := &repoTestSuite{}
	t.Run("FindUnknownUser", func(t *testing.T) { s.testFindUnknownUser(t) })
	t.Run("FindKnownUser", func(t *testing.T) { s.testFindKnownUser(t) })
	t.Run("StoreNewUser", func(t *testing.T) { s.testStoreNewUser(t) })
	t.Run("StoreDupUser", func(t *testing.T) { s.testStoreDupUser(t) })
	t.Run("FindUnknownGroup", func(t *testing.T) { s.testFindUnknownGroup(t) })
	t.Run("FindKnownGroup", func(t *testing.T) { s.testFindKnownGroup(t) })
	t.Run("StoreNewGroup", func(t *testing.T) { s.testStoreNewGroup(t) })
	t.Run("StoreDupGroup", func(t *testing.T) { s.testStoreDupGroup(t) })
	t.Run("FetchGroupusers", func(t *testing.T) { s.testFetchGroupUsers(t) })
}

// Test suite for user repository
type repoTestSuite struct{}

// Test scenario - Find a known user
func (s *repoTestSuite) testFindKnownUser(t *testing.T) {

	is := is.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := mock.NewRows([]string{"count(name)"}).AddRow(1)

	mock.ExpectQuery("SELECT").WithArgs("tstuser").WillReturnRows(rows)

	{
		repository := &userRepository{db}
		found, err := repository.FindUser(context.TODO(), "tstuser")
		is.NoErr(err)
		is.Equal(found, true)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

// Test scenario - Find an unknown user
func (s *repoTestSuite) testFindUnknownUser(t *testing.T) {

	is := is.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := mock.NewRows([]string{"count(name)"}).AddRow(0)

	mock.ExpectQuery("SELECT").WithArgs("tstuser").WillReturnRows(rows)

	{
		repository := &userRepository{db}
		found, err := repository.FindUser(context.TODO(), "tstuser")
		is.NoErr(err)
		is.Equal(found, false)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

// Test scenario - Store a new user
func (s *repoTestSuite) testStoreNewUser(t *testing.T) {
	is := is.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	result := sqlmock.NewResult(1, 1)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").WithArgs("tstuser").WillReturnResult(result)
	mock.ExpectCommit()

	{
		repository := &userRepository{db}
		err := repository.StoreUser(context.TODO(), "tstuser")
		is.NoErr(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// Test scenario - Store a duplicate user
func (s *repoTestSuite) testStoreDupUser(t *testing.T) {
	is := is.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	result := sqlmock.NewErrorResult(fmt.Errorf("duplicate user"))

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").WithArgs("tstuser").WillReturnResult(result)
	mock.ExpectRollback()

	{
		repository := &userRepository{db}
		err := repository.StoreUser(context.TODO(), "tstuser")
		is.True(err != nil)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

// Test scenario - Find an unknown group
func (s *repoTestSuite) testFindUnknownGroup(t *testing.T) {

	is := is.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := mock.NewRows([]string{"count(name)"}).AddRow(0)

	mock.ExpectQuery("SELECT").WithArgs("tstgroup").WillReturnRows(rows)

	{
		repository := &userRepository{db}
		found, err := repository.FindGroup(context.TODO(), "tstgroup")
		is.NoErr(err)
		is.Equal(found, false)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

// Test scenario - Find a known group
func (s *repoTestSuite) testFindKnownGroup(t *testing.T) {

	is := is.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := mock.NewRows([]string{"count(name)"}).AddRow(1)

	mock.ExpectQuery("SELECT").WithArgs("tstgroup").WillReturnRows(rows)

	{
		repository := &userRepository{db}
		found, err := repository.FindGroup(context.TODO(), "tstgroup")
		is.NoErr(err)
		is.Equal(found, true)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

// Test scenario - Store a new group
func (s *repoTestSuite) testStoreNewGroup(t *testing.T) {
	is := is.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	result := sqlmock.NewResult(1, 1)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO usergroups").WithArgs("tstgroup").WillReturnResult(result)
	mock.ExpectExec("INSERT INTO groupusers").WithArgs("tstgroup", "tstusr1").WillReturnResult(result)
	mock.ExpectExec("INSERT INTO groupusers").WithArgs("tstgroup", "tstusr2").WillReturnResult(result)
	mock.ExpectCommit()

	{
		repository := &userRepository{db}
		err := repository.StoreGroup(context.TODO(), "tstgroup", []string{"tstusr1", "tstusr2"})
		is.NoErr(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// Test scenario - Store a duplicate group
func (s *repoTestSuite) testStoreDupGroup(t *testing.T) {
	is := is.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	result := sqlmock.NewErrorResult(fmt.Errorf("duplicate group"))

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO usergroups").WithArgs("tstgroup").WillReturnResult(result)
	mock.ExpectRollback()

	{
		repository := &userRepository{db}
		err := repository.StoreGroup(context.TODO(), "tstgroup", []string{"tstusr1"})
		is.True(err != nil)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}

// Test scenario - Fetch group users
func (s *repoTestSuite) testFetchGroupUsers(t *testing.T) {

	is := is.New(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := mock.NewRows([]string{"username"}).AddRow("tstuser1").AddRow("tstuser2")

	mock.ExpectQuery("SELECT username FROM groupusers").WithArgs("tstgroup").WillReturnRows(rows)

	{
		repository := &userRepository{db}
		users, err := repository.FetchGroupUsers(context.TODO(), "tstgroup")
		is.NoErr(err)
		is.Equal(2, len(users))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}
