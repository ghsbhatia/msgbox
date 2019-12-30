package msgstore

import (
	"encoding/json"
	"fmt"
	_ "strings"
	"testing"

	"github.com/matryer/is"
)

// Test executor for message transport
func TestMessageTransport(t *testing.T) {
	s := &messageTestSuite{}
	t.Run("MessageMarshal", func(t *testing.T) { s.testMessageMarshal(t) })
	t.Run("MessageUnmarshal", func(t *testing.T) { s.testMessageUnmarshal(t) })
}

// Test suite for message creation
type messageTestSuite struct{}

// Test scenrio - Marshal message instance into json string
func (s *messageTestSuite) testMessageMarshal(t *testing.T) {

	is := is.New(t)

	msg := &message{
		Id:        "id1",
		Re:        "id2",
		Sender:    "bob",
		Recipient: receiver{Groupname: "testgroup"},
		Subject:   "test",
		Body:      "test message",
		Timestamp: "2019-09-03T18:32:01Z",
	}

	data, err := json.Marshal(msg)

	fmt.Printf("%s\n", data)

	is.NoErr(err)
	is.Equal(data, []byte(`{"id":"id1","re":"id2","sender":"bob","recipient":{"groupname":"testgroup"},"subject":"test","body":"test message","sentAt":"2019-09-03T18:32:01Z"}`))

}

// Test scenrio - Unmarshal json string into messageQueryResponse instance
func (s *messageTestSuite) testMessageUnmarshal(t *testing.T) {

	is := is.New(t)

	msg1 := message{
		Id:        "id1",
		Re:        "id2",
		Sender:    "bob",
		Recipient: receiver{Groupname: "testgroup"},
		Subject:   "test",
		Body:      "test message",
		Timestamp: "2019-09-03T18:32:01Z",
	}

	msg2 := message{
		Id:        "id2",
		Sender:    "bob",
		Recipient: receiver{Groupname: "testgroup"},
		Subject:   "test",
		Body:      "test message",
		Timestamp: "2019-09-03T16:32:01Z",
	}

	messages := []message{msg1, msg2}

	response := replyQueryResponse{}
	response.Content = messages

	data, err := json.Marshal(response.Content)

	fmt.Printf("%s\n", data)

	is.NoErr(err)
}
