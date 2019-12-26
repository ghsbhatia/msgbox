package useradmin

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/matryer/is"
)

// Test executor for user creation
func TestUser(t *testing.T) {
	s := &userTestSuite{}
	t.Run("UserMarshal", func(t *testing.T) { s.testUserMarshal(t) })
	t.Run("UserUnmarshal", func(t *testing.T) { s.testUserUnMarshal(t) })
}

// Test suite for user creation
type userTestSuite struct{}

// Test scenrio - Marshal userRegistrationRequest instance into json string
func (s *userTestSuite) testUserMarshal(t *testing.T) {

	is := is.New(t)

	alice := &userRegistrationRequest{"alice"}

	data, err := json.Marshal(alice)

	is.NoErr(err)
	is.Equal(data, []byte(`{"username":"alice"}`))

}

// Test scenrio - Unmarshal json string into userRegistrationRequest instance
func (s *userTestSuite) testUserUnMarshal(t *testing.T) {

	is := is.New(t)

	data := []byte(`{"username": "alice"}`)

	var userReg userRegistrationRequest
	err := json.Unmarshal(data, &userReg)

	is.NoErr(err)
	is.Equal(userReg.Username, "alice")

}

// Test executor for group creation
func TestGroup(t *testing.T) {
	s := &groupTestSuite{}
	t.Run("GroupMarshal", func(t *testing.T) { s.testGroupMarshal(t) })
	t.Run("GroupUnmarshal", func(t *testing.T) { s.testGroupUnMarshal(t) })
}

// Test suite for group creation
type groupTestSuite struct{}

// Test scenrio - Marshal GroupCreation instance into json string
func (s *groupTestSuite) testGroupMarshal(t *testing.T) {

	is := is.New(t)

	groupname := "quantummetric"
	usernames := []string{"alice", "bob", "carole"}
	qmetricgc := &groupRegistrationRequest{groupname, usernames}

	data, err := json.Marshal(qmetricgc)

	is.NoErr(err)
	str := fmt.Sprintf("%s", data)
	is.True(strings.Contains(str, "quantummetric"))
	is.True(strings.Contains(str, "bob"))
	fmt.Printf("%s", data)
}

// Test scenrio - Unmarshal json string into GroupCreation instance
func (s *groupTestSuite) testGroupUnMarshal(t *testing.T) {

	is := is.New(t)

	data := []byte(`
    {
      "groupname": "quantummetric",
      "usernames": ["alice", "bob", "carole"]
    }
  `)

	var grpCreation groupRegistrationRequest
	err := json.Unmarshal(data, &grpCreation)

	is.NoErr(err)
	is.Equal(grpCreation.Groupname, "quantummetric")
	is.Equal(grpCreation.Usernames[0], "alice")
	is.Equal(len(grpCreation.Usernames), 3)

}
