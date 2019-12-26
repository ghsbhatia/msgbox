package msgstore

import (
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
)

// Test executor for message repository
func TestMessageRepository(t *testing.T) {
	r := NewMessageRepository("mongodb://root:secret@127.0.0.1:27017/msgbox-mongo?authSource=admin&gssapiServiceName=mongodb", "test")
	s := &repositoryTestSuite{}
	t.Run("StoreMessage", func(t *testing.T) { s.testStoreMessage(t, r) })
	t.Run("StoreReply", func(t *testing.T) { s.testStoreReply(t, r) })
	t.Run("GetMessage", func(t *testing.T) { s.testGetMessage(t, r) })
	t.Run("GetUserMessages", func(t *testing.T) { s.testGetUserMessages(t, r) })
}

// Test suite for message respository
type repositoryTestSuite struct{}

// Test scenario - store a message
func (s *repositoryTestSuite) testStoreMessage(t *testing.T, r MessageRepository) {

	r.Purge()

	is := is.New(t)

	msg := &Message{Sender: "alice", Recipients: []string{"bob"}, Subject: "test", Body: "this is a test message"}
	msgid, err := r.StoreMessage(msg)

	is.NoErr(err)
	is.True(msgid != "")

}

// Test scenario - store reply to a message
func (s *repositoryTestSuite) testStoreReply(t *testing.T, r MessageRepository) {

	r.Purge()

	is := is.New(t)

	msg := &Message{Sender: "alice", Recipients: []string{"bob", "peter"}, Subject: "test", Body: "this is a test message"}
	msgid, err := r.StoreMessage(msg)

	is.NoErr(err)
	is.True(msgid != "")

	{
		msg := &Message{ReplyToMsgId: msgid, Sender: "bob", Recipients: []string{"alice"}, Subject: "re:test", Body: "send another message"}
		mid, err := r.StoreMessage(msg)

		is.NoErr(err)
		is.True(mid != "")
	}

	{
		msg := &Message{ReplyToMsgId: msgid, Sender: "peter", Recipients: []string{"alice"}, Subject: "re:test", Body: "message acknowledged"}
		mid, err := r.StoreMessage(msg)

		is.NoErr(err)
		is.True(mid != "")
	}

	{
		messages, err := r.GetReplyMessages(msgid)

		is.NoErr(err)
		is.True(len(messages) == 2)
	}

}

// Test scenario - Get message
func (s *repositoryTestSuite) testGetMessage(t *testing.T, r MessageRepository) {

	r.Purge()

	is := is.New(t)

	msg := &Message{Sender: "alice", Recipients: []string{"bob"}, Subject: "test", Body: "this is a test message"}
	msgid, err := r.StoreMessage(msg)

	is.NoErr(err)

	{
		msg, err := r.GetMessage(msgid)

		is.NoErr(err)
		is.Equal(msg.Sender, "alice")
		is.Equal(msg.Recipients, []string{"bob"})
		is.Equal(msg.Subject, "test")
		is.Equal(msg.Body, "this is a test message")
		delta := time.Now().Unix() - msg.Timestamp.Unix()
		is.True(delta*delta < 4)
		fmt.Printf("timestamp = %s\n", msg.Timestamp.Format(time.RFC3339))
	}
}

// Test scenario - get messages sent to user
func (s *repositoryTestSuite) testGetUserMessages(t *testing.T, r MessageRepository) {

	r.Purge()

	is := is.New(t)

	msg := &Message{Sender: "alice", Recipients: []string{"bob", "peter"}, Subject: "test", Body: "this is a test message"}
	msgid, err := r.StoreMessage(msg)

	is.NoErr(err)
	is.True(msgid != "")

	{
		msg := &Message{ReplyToMsgId: msgid, Sender: "bob", Recipients: []string{"alice", "peter"}, Subject: "re:test", Body: "send another message"}
		mid, err := r.StoreMessage(msg)

		is.NoErr(err)
		is.True(mid != "")
	}

	{
		msg := &Message{ReplyToMsgId: msgid, Sender: "bob", Recipients: []string{"alice", "peter"}, Subject: "re:test", Body: "message acknowledged"}
		mid, err := r.StoreMessage(msg)

		is.NoErr(err)
		is.True(mid != "")
	}

	{
		messages, err := r.GetUserMessages("alice")

		is.NoErr(err)
		is.True(len(messages) == 2)
	}

	{
		messages, err := r.GetUserMessages("bob")

		is.NoErr(err)
		is.True(len(messages) == 1)
	}

}
