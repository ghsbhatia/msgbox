package msgstore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedSvcClient struct {
	Groupname string
	Usernames []string
}

func (m *MockedSvcClient) Get(url string, v interface{}) error {
	bytearray, _ := json.Marshal(m)
	err := json.Unmarshal(bytearray, v)
	return err
}

type MockedRepository struct {
	mock.Mock
}

func (m *MockedRepository) StoreMessage(ctx context.Context, rec *Record) (msgid string, err error) {
	args := m.Called(ctx, rec)
	return args.String(0), args.Error(1)
}

func (m *MockedRepository) GetUserMessages(ctx context.Context, user string) (results []Record, err error) {
	args := m.Called(ctx, user)
	return args.Get(0).([]Record), args.Error(1)
}

func (m *MockedRepository) GetMessage(ctx context.Context, msgid string) (result Record, err error) {
	args := m.Called(ctx, msgid)
	return args.Get(0).(Record), args.Error(1)
}

func (m *MockedRepository) GetReplyMessages(ctx context.Context, msgid string) (results []Record, err error) {
	args := m.Called(ctx, msgid)
	return args.Get(0).([]Record), args.Error(1)
}

func (m *MockedRepository) Purge(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test executor for message store service
func TestMessageStoreService(t *testing.T) {
	s := &serviceTestSuite{}
	t.Run("StoreMessageForUser", func(t *testing.T) { s.testStoreMessageForUser(t) })
	t.Run("StoreMessageForGroup", func(t *testing.T) { s.testStoreMessageForGroup(t) })
	t.Run("StoreReply", func(t *testing.T) { s.testStoreReply(t) })
	t.Run("GetMessageForId", func(t *testing.T) { s.testGetMessageForId(t) })
	t.Run("GetMessagesForUser", func(t *testing.T) { s.testGetMessagesForUser(t) })
	t.Run("GetReplyMessages", func(t *testing.T) { s.testGetReplyMessages(t) })
}

// Test suite for message store service
type serviceTestSuite struct{}

// Test scenario - Store a message for recipient user
func (s *serviceTestSuite) testStoreMessageForUser(t *testing.T) {
	ctx := context.TODO()

	rec := &Record{
		Sender:     "tester",
		Subject:    "test",
		Body:       "body",
		Recipients: []string{"user1"},
	}

	repository := new(MockedRepository)
	repository.On("StoreMessage", ctx, rec).Return("id:01", nil)

	service := NewService(repository, nil, "/foo")

	rcv := receiver{Username: "user1"}
	msg := message{Sender: "tester", Subject: "test", Body: "body", Recipient: rcv}

	msgid, _ := service.StoreMessage(ctx, msg)

	repository.AssertExpectations(t)
	assert.Equal(t, "id:01", msgid)
}

// Test scenario - Store a message for recipient group
func (s *serviceTestSuite) testStoreMessageForGroup(t *testing.T) {
	ctx := context.TODO()

	rec := &Record{
		Sender:     "tester",
		Subject:    "test",
		Body:       "body",
		GroupId:    "tstgroup",
		Recipients: []string{"user1", "user2"},
	}

	repository := new(MockedRepository)
	repository.On("StoreMessage", ctx, rec).Return("id:01", nil)

	svcclient := &MockedSvcClient{"tstgroup", []string{"user1", "user2"}}

	service := NewService(repository, svcclient, "/foo")

	rcv := receiver{Groupname: "tstgroup"}
	msg := message{Sender: "tester", Subject: "test", Body: "body", Recipient: rcv}

	msgid, _ := service.StoreMessage(ctx, msg)

	repository.AssertExpectations(t)
	assert.Equal(t, "id:01", msgid)
}

// Test scenario - Store a reply message
func (s *serviceTestSuite) testStoreReply(t *testing.T) {
	ctx := context.TODO()

	rec1 := Record{
		Id:         "id:01",
		Sender:     "tester",
		Subject:    "test",
		Body:       "body",
		Recipients: []string{"user1"},
	}

	rec2 := &Record{
		ReplyToMsgId: "id:01",
		Sender:       "user1",
		Subject:      "re:test",
		Body:         "body",
		Recipients:   []string{"tester"},
	}

	repository := new(MockedRepository)
	repository.On("GetMessage", ctx, "id:01").Return(rec1, nil)
	repository.On("StoreMessage", ctx, rec2).Return("id:02", nil)

	service := NewService(repository, nil, "/foo")

	msg := message{Re: "id:01", Sender: "user1", Subject: "re:test", Body: "body"}

	msgid, _ := service.StoreMessage(ctx, msg)

	repository.AssertExpectations(t)
	assert.Equal(t, "id:02", msgid)
}

// Test scenario - Get a message for given id
func (s *serviceTestSuite) testGetMessageForId(t *testing.T) {
	ctx, ts := context.TODO(), time.Now()

	rec := Record{
		Id:         "id:01",
		Sender:     "tester",
		Subject:    "test",
		Body:       "body",
		Recipients: []string{"user1"},
		Timestamp:  ts,
	}

	repository := new(MockedRepository)
	repository.On("GetMessage", ctx, "id:01").Return(rec, nil)

	service := NewService(repository, nil, "/foo")

	msg, _ := service.GetMessage(ctx, "id:01")

	repository.AssertExpectations(t)

	recipient := receiver{Username: "user1"}
	exp := message{
		Id:        "id:01",
		Sender:    "tester",
		Subject:   "test",
		Body:      "body",
		Recipient: recipient,
		Timestamp: ts.Format(time.RFC3339),
	}

	assert.Equal(t, msg, exp)
}

// Test scenario - Get messages for user
func (s *serviceTestSuite) testGetMessagesForUser(t *testing.T) {
	ctx, ts1, ts2, ts3, ts4 := context.TODO(), time.Now(), time.Now(), time.Now(), time.Now()

	rec1 := Record{
		Id:         "id:01",
		Sender:     "tester1",
		Subject:    "test1",
		Body:       "body1",
		Recipients: []string{"user1"},
		Timestamp:  ts1,
	}
	rec2 := Record{
		Id:         "id:02",
		Sender:     "tester2",
		Subject:    "test2",
		Body:       "body2",
		Recipients: []string{"user1"},
		Timestamp:  ts2,
	}
	rec3 := Record{
		Id:         "id:03",
		Sender:     "tester3",
		Subject:    "test3",
		Body:       "body3",
		GroupId:    "tstgroup",
		Recipients: []string{"user1", "user2"},
		Timestamp:  ts3,
	}
	rec4 := Record{
		Id:           "id:04",
		ReplyToMsgId: "id:03",
		Sender:       "user2",
		Subject:      "Re:test3",
		Body:         "body3",
		GroupId:      "tstgroup",
		Recipients:   []string{"user1", "user2", "tester3"},
		Timestamp:    ts4,
	}

	repository := new(MockedRepository)
	repository.On("GetUserMessages", ctx, "user1").Return([]Record{rec1, rec2, rec3, rec4}, nil)

	service := NewService(repository, nil, "/foo")

	msgs, _ := service.GetMessages(ctx, "user1")

	repository.AssertExpectations(t)

	exp := message{
		Id:        "id:01",
		Sender:    "tester1",
		Subject:   "test1",
		Body:      "body1",
		Recipient: receiver{Username: "user1"},
		Timestamp: ts1.Format(time.RFC3339),
	}

	assert.Equal(t, msgs[0], exp)

	exp = message{
		Id:        "id:02",
		Sender:    "tester2",
		Subject:   "test2",
		Body:      "body2",
		Recipient: receiver{Username: "user1"},
		Timestamp: ts2.Format(time.RFC3339),
	}

	assert.Equal(t, msgs[1], exp)

	exp = message{
		Id:        "id:03",
		Sender:    "tester3",
		Subject:   "test3",
		Body:      "body3",
		Recipient: receiver{Groupname: "tstgroup"},
		Timestamp: ts3.Format(time.RFC3339),
	}

	assert.Equal(t, msgs[2], exp)
	exp = message{
		Id:        "id:04",
		Re:        "id:03",
		Sender:    "user2",
		Subject:   "Re:test3",
		Body:      "body3",
		Recipient: receiver{Groupname: "tstgroup"},
		Timestamp: ts4.Format(time.RFC3339),
	}

	assert.Equal(t, msgs[3], exp)
}

// Test scenario - Get reply messages
func (s *serviceTestSuite) testGetReplyMessages(t *testing.T) {
	ctx, ts1, ts2 := context.TODO(), time.Now(), time.Now()
	rec1 := Record{
		Id:           "id:01",
		ReplyToMsgId: "idtest",
		Sender:       "user1",
		Subject:      "re:test",
		Body:         "body1",
		Recipients:   []string{"user2"},
		Timestamp:    ts1,
	}
	rec2 := Record{
		Id:           "id:02",
		ReplyToMsgId: "idtest",
		Sender:       "user1",
		Subject:      "re:re:test",
		Body:         "body2",
		Recipients:   []string{"user2"},
		Timestamp:    ts2,
	}

	repository := new(MockedRepository)
	repository.On("GetReplyMessages", ctx, "idtest").Return([]Record{rec1, rec2}, nil)

	service := NewService(repository, nil, "/foo")

	msgs, _ := service.GetReplies(ctx, "idtest")

	repository.AssertExpectations(t)

	exp := message{
		Id:        "id:01",
		Re:        "idtest",
		Sender:    "user1",
		Subject:   "re:test",
		Body:      "body1",
		Recipient: receiver{Username: "user2"},
		Timestamp: ts1.Format(time.RFC3339),
	}

	assert.Equal(t, msgs[0], exp)

	exp = message{
		Id:        "id:02",
		Re:        "idtest",
		Sender:    "user1",
		Subject:   "re:re:test",
		Body:      "body2",
		Recipient: receiver{Username: "user2"},
		Timestamp: ts1.Format(time.RFC3339),
	}

	assert.Equal(t, msgs[1], exp)

}
