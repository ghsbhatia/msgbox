package msgstore

import (
	"context"
	"encoding/json"
	_ "fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kitlog "github.com/go-kit/kit/log"
	_ "github.com/matryer/is"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedService struct {
	mock.Mock
}

func (m *MockedService) StoreMessage(ctx context.Context, msg message) (string, error) {
	args := m.Called(msg)
	return args.String(0), args.Error(1)
}

func (m *MockedService) GetMessage(ctx context.Context, msgid string) (message, error) {
	args := m.Called(msgid)
	return args.Get(0).(message), args.Error(1)
}

func (m *MockedService) GetMessages(ctx context.Context, userid string) ([]message, error) {
	args := m.Called(userid)
	_, ok := args.Get(0).([]message)
	if !ok {
		return nil, args.Error(1)
	}
	return args.Get(0).([]message), args.Error(1)
}

func (m *MockedService) GetReplies(ctx context.Context, msgid string) ([]message, error) {
	args := m.Called(msgid)
	_, ok := args.Get(0).([]message)
	if !ok {
		return nil, args.Error(1)
	}
	return args.Get(0).([]message), args.Error(1)
}

// Test executor for message transport
func TestMessageTransport(t *testing.T) {
	s := &messageTestSuite{}
	t.Run("MessageMarshal", func(t *testing.T) { s.testMessageMarshal(t) })
	t.Run("MessagesMarshal", func(t *testing.T) { s.testMessagesMarshal(t) })
	t.Run("StoreMessage", func(t *testing.T) { s.testStoreMessage(t) })
	t.Run("StoreMessageInvalidUser", func(t *testing.T) { s.testStoreMessageInvalidUser(t) })
	t.Run("StoreMessageInvalidGroup", func(t *testing.T) { s.testStoreMessageInvalidGroup(t) })
	t.Run("StoreMessageNoSender", func(t *testing.T) { s.testStoreMessageNoSender(t) })
	t.Run("StoreMessageNoRecipient", func(t *testing.T) { s.testStoreMessageNoRecipient(t) })
	t.Run("StoreMessageSystemException", func(t *testing.T) { s.testStoreMessageSystemException(t) })
	t.Run("StoreReply", func(t *testing.T) { s.testStoreReply(t) })
	t.Run("StoreReplyInvalidId", func(t *testing.T) { s.testStoreReplyInvalidId(t) })
	t.Run("GetReplies", func(t *testing.T) { s.testGetReplies(t) })
	t.Run("GetRepliesInvalidId", func(t *testing.T) { s.testGetRepliesInvalidId(t) })
	t.Run("GetMessage", func(t *testing.T) { s.testGetMessage(t) })
	t.Run("GetMessageInvalidId", func(t *testing.T) { s.testGetMessageInvalidId(t) })
	t.Run("GetMessagesForUser", func(t *testing.T) { s.testGetMessagesForUser(t) })
	t.Run("GetMessagesInvalidUser", func(t *testing.T) { s.testGetMessagesInvalidUser(t) })
}

// Test suite for message creation
type messageTestSuite struct{}

// Test scenario - Store New Message
func (s *messageTestSuite) testStoreMessage(t *testing.T) {

	msg := message{
		Id:        "id1",
		Sender:    "tester",
		Recipient: receiver{Groupname: "testgroup"},
		Subject:   "test",
		Body:      "test message",
	}

	service := new(MockedService)
	service.On("StoreMessage", msg).Return("id:01", nil)

	data, _ := json.Marshal(msg)
	body := strings.NewReader(string(data))

	req := httptest.NewRequest("POST", "http://foo.com/messages", body)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"id":"id:01"}`, content)
	}

}

// Test scenario - Store New Message with Invalid User recipient
func (s *messageTestSuite) testStoreMessageInvalidUser(t *testing.T) {

	msg := message{
		Id:        "id1",
		Sender:    "tester",
		Recipient: receiver{Username: "unknown"},
		Subject:   "test",
		Body:      "test message",
	}

	service := new(MockedService)
	service.On("StoreMessage", msg).Return("", ErrUserNotFound)

	data, _ := json.Marshal(msg)
	body := strings.NewReader(string(data))

	req := httptest.NewRequest("POST", "http://foo.com/messages", body)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"error":"user not found"}`, content)
	}

}

// Test scenario - Store New Message with Invalid Group recipient
func (s *messageTestSuite) testStoreMessageInvalidGroup(t *testing.T) {

	msg := message{
		Id:        "id1",
		Sender:    "tester",
		Recipient: receiver{Groupname: "unknown"},
		Subject:   "test",
		Body:      "test message",
	}

	service := new(MockedService)
	service.On("StoreMessage", msg).Return("", ErrGroupNotFound)

	data, _ := json.Marshal(msg)
	body := strings.NewReader(string(data))

	req := httptest.NewRequest("POST", "http://foo.com/messages", body)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"error":"group not found"}`, content)
	}

}

// Test scenario - Store New Message System Exception
func (s *messageTestSuite) testStoreMessageSystemException(t *testing.T) {

	msg := message{
		Id:        "id1",
		Sender:    "tester",
		Recipient: receiver{Groupname: "testgroup"},
		Subject:   "test",
		Body:      "test message",
	}

	service := new(MockedService)
	service.On("StoreMessage", msg).Return("", ErrSystemError)

	data, _ := json.Marshal(msg)
	body := strings.NewReader(string(data))

	req := httptest.NewRequest("POST", "http://foo.com/messages", body)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"error":"system error"}`, content)
	}

}

// Test scenario - Store New Message with sender missing
func (s *messageTestSuite) testStoreMessageNoSender(t *testing.T) {

	msg := message{
		Id:        "id1",
		Recipient: receiver{Groupname: "testgroup"},
		Subject:   "test",
		Body:      "test message",
	}

	service := new(MockedService)

	data, _ := json.Marshal(msg)
	body := strings.NewReader(string(data))

	req := httptest.NewRequest("POST", "http://foo.com/messages", body)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"error":"invalid request"}`, content)
	}

}

// Test scenario - Store New Message with recipient missing
func (s *messageTestSuite) testStoreMessageNoRecipient(t *testing.T) {

	msg := message{
		Id:      "id1",
		Sender:  "tester",
		Subject: "test",
		Body:    "test message",
	}

	service := new(MockedService)

	data, _ := json.Marshal(msg)
	body := strings.NewReader(string(data))

	req := httptest.NewRequest("POST", "http://foo.com/messages", body)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"error":"invalid request"}`, content)
	}

}

// Test scenario - Get Message
func (s *messageTestSuite) testGetMessage(t *testing.T) {

	msg := message{
		Id:        "id1",
		Sender:    "tester",
		Recipient: receiver{Groupname: "testgroup"},
		Subject:   "test",
		Body:      "test message",
		Timestamp: "2019-09-03T18:32:01Z",
	}

	service := new(MockedService)
	service.On("GetMessage", "id1").Return(msg, nil)

	req := httptest.NewRequest("GET", "http://foo.com/messages/id1", nil)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		exp, _ := json.Marshal(msg)
		assert.Equal(t, exp, []byte(content))
	}

}

// Test scenario - Get Message with Invalid Id
func (s *messageTestSuite) testGetMessageInvalidId(t *testing.T) {

	service := new(MockedService)
	service.On("GetMessage", "id1").Return(message{}, ErrMsgNotFound)

	req := httptest.NewRequest("GET", "http://foo.com/messages/id1", nil)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"error":"message not found"}`, content)
	}

}

// Test scenario - Get Messages for User
func (s *messageTestSuite) testGetMessagesForUser(t *testing.T) {

	msg1 := message{
		Id:        "id1",
		Sender:    "user1",
		Recipient: receiver{Username: "tester"},
		Subject:   "test",
		Body:      "test message",
		Timestamp: "2019-09-03T18:32:01Z",
	}

	msg2 := message{
		Id:        "id2",
		Sender:    "user2",
		Recipient: receiver{Username: "tester"},
		Subject:   "test",
		Body:      "test message",
		Timestamp: "2019-09-03T18:40:32Z",
	}

	service := new(MockedService)
	service.On("GetMessages", "tester").Return([]message{msg1, msg2}, nil)

	req := httptest.NewRequest("GET", "http://foo.com/users/tester/mailbox", nil)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		exp, _ := json.Marshal([]message{msg1, msg2})
		assert.Equal(t, exp, []byte(content))
	}

}

// Test scenario - Get Messages for User Invalid User Id
func (s *messageTestSuite) testGetMessagesInvalidUser(t *testing.T) {

	service := new(MockedService)
	service.On("GetMessages", "unknown").Return(nil, ErrUserNotFound)

	req := httptest.NewRequest("GET", "http://foo.com/users/unknown/mailbox", nil)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"error":"user not found"}`, content)
	}

}

// Test scenario - Store Reply
func (s *messageTestSuite) testStoreReply(t *testing.T) {

	msg := message{
		Sender:  "tester",
		Subject: "test",
		Body:    "test message",
	}

	rep := message{
		Re:      "id:01",
		Sender:  "tester",
		Subject: "test",
		Body:    "test message",
	}

	service := new(MockedService)
	service.On("StoreMessage", rep).Return("id:02", nil)

	data, _ := json.Marshal(msg)
	body := strings.NewReader(string(data))

	req := httptest.NewRequest("POST", "http://foo.com/messages/id:01/replies", body)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"id":"id:02"}`, content)
	}

}

// Test scenario - Store Reply with Invalid Message Id
func (s *messageTestSuite) testStoreReplyInvalidId(t *testing.T) {

	msg := message{
		Sender:  "tester",
		Subject: "test",
		Body:    "test message",
	}

	rep := message{
		Re:      "unknown",
		Sender:  "tester",
		Subject: "test",
		Body:    "test message",
	}

	service := new(MockedService)
	service.On("StoreMessage", rep).Return("", ErrMsgNotFound)

	data, _ := json.Marshal(msg)
	body := strings.NewReader(string(data))

	req := httptest.NewRequest("POST", "http://foo.com/messages/unknown/replies", body)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"error":"message not found"}`, content)
	}

}

// Test scenario - Get Replies for Given Message Id
func (s *messageTestSuite) testGetReplies(t *testing.T) {

	msg1 := message{
		Id:        "id1",
		Re:        "id0",
		Sender:    "user1",
		Recipient: receiver{Username: "tester"},
		Subject:   "re:test",
		Body:      "test message",
		Timestamp: "2019-09-03T18:32:01Z",
	}

	msg2 := message{
		Id:        "id2",
		Re:        "id0",
		Sender:    "user2",
		Recipient: receiver{Username: "tester"},
		Subject:   "re:test",
		Body:      "test message",
		Timestamp: "2019-09-03T18:40:32Z",
	}

	service := new(MockedService)
	service.On("GetReplies", "id0").Return([]message{msg1, msg2}, nil)

	req := httptest.NewRequest("GET", "http://foo.com/messages/id0/replies", nil)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		exp, _ := json.Marshal([]message{msg1, msg2})
		assert.Equal(t, exp, []byte(content))
	}

}

// Test scenario - Get Replies Invalid Message Id
func (s *messageTestSuite) testGetRepliesInvalidId(t *testing.T) {

	service := new(MockedService)
	service.On("GetReplies", "unknown").Return(nil, ErrMsgNotFound)

	req := httptest.NewRequest("GET", "http://foo.com/messages/unknown/replies", nil)

	w := httptest.NewRecorder()

	MakeHandler(service, kitlog.NewNopLogger()).ServeHTTP(w, req)

	service.AssertExpectations(t)

	{
		resp := w.Result()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		content := strings.Trim(string(body), "\n")
		assert.Equal(t, `{"error":"message not found"}`, content)
	}

}

// Test scenrio - Marshal message instance into json string
func (s *messageTestSuite) testMessageMarshal(t *testing.T) {

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

	assert.NotNil(t, data)
	assert.NoError(t, err)
	assert.Equal(t, data, []byte(`{"id":"id1","re":"id2","sender":"bob","recipient":{"groupname":"testgroup"},"subject":"test","body":"test message","sentAt":"2019-09-03T18:32:01Z"}`))

}

// Test scenrio - Marshal messages into json string
func (s *messageTestSuite) testMessagesMarshal(t *testing.T) {

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

	assert.NotNil(t, data)
	assert.NoError(t, err)
}
