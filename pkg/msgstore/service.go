package msgstore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ghsbhatia/msgbox/pkg/svcclient"
)

const no_docs_in_result = "no documents in result"
const user_not_found = "user:404"
const group_not_found = "group:404"

// Service interface for message store functions
type Service interface {
	// store a new message and return its id
	StoreMessage(context.Context, message) (string, error)
	// get message for a given id
	GetMessage(context.Context, string) (message, error)
	// get messages for a given user
	GetMessages(context.Context, string) ([]message, error)
	// get replies for a given message id
	GetReplies(context.Context, string) ([]message, error)
}

// Create a new service instance with a given message repository
func NewService(repository MessageRepository, client svcclient.HttpServiceClient, usersvcurl string) Service {
	return &service{repository, client, usersvcurl}
}

type service struct {
	repository   MessageRepository
	httpsvclient svcclient.HttpServiceClient
	usersvcurl   string
}

type replyMessageRecipient struct {
	sender string // Sender of original message
	group  string // Optional and present only if original message is for group
}

// Store the given message
func (s *service) StoreMessage(ctx context.Context, msg message) (string, error) {

	// Ensure sender is a registered user
	{
		_, err := s.getUser(ctx, msg.Sender)
		if err != nil {
			return "", s.mapError(err)
		}
	}

	if len(msg.Re) > 0 {
		return s.storeReply(ctx, msg)
	}

	// If message recipient is a group, get users for the group and store them as
	// recipients

	recipients := []string{}
	if len(msg.Recipient.Groupname) > 0 {
		grpusers, err := s.getGroupUsers(ctx, msg.Recipient.Groupname)
		if err != nil {
			return "", s.mapError(err)
		}
		recipients = grpusers
	} else if len(msg.Recipient.Username) > 0 {
		_, err := s.getUser(ctx, msg.Recipient.Username)
		if err != nil {
			return "", s.mapError(err)
		}
		recipients = append(recipients, msg.Recipient.Username)
	}

	record := &Record{
		ReplyToMsgId: msg.Re,
		Sender:       msg.Sender,
		Recipients:   recipients,
		GroupId:      msg.Recipient.Groupname,
		Subject:      msg.Subject,
		Body:         msg.Body,
	}

	return s.repository.StoreMessage(ctx, record)
}

// Get message corresponding to its id
func (s *service) GetMessage(ctx context.Context, msgid string) (message, error) {
	record, err := s.repository.GetMessage(ctx, msgid)
	var msg message
	if err == nil {
		msg = mapRecord(record)
	}
	return msg, s.mapError(err)
}

// Get messages for a given user
func (s *service) GetMessages(ctx context.Context, userid string) ([]message, error) {
	_, iderr := s.getUser(ctx, userid)
	if iderr != nil {
		return nil, s.mapError(iderr)
	}
	records, err := s.repository.GetUserMessages(ctx, userid)
	msgs := mapRecords(records)
	return msgs, s.mapError(err)
}

// Get reply messages for message identified by given message id
func (s *service) GetReplies(ctx context.Context, msgid string) ([]message, error) {
	_, iderr := s.repository.GetMessage(ctx, msgid)
	if iderr != nil {
		return nil, s.mapError(iderr)
	}
	records, err := s.repository.GetReplyMessages(ctx, msgid)
	msgs := mapRecords(records)
	return msgs, s.mapError(err)
}

// Get users for group identified by given group id
func (s *service) getGroupUsers(ctx context.Context, groupid string) ([]string, error) {
	var group struct {
		Groupname string   `json:"groupname"`
		Usernames []string `json:"usernames"`
	}
	requesturl := fmt.Sprintf("%s/groups/%s", s.usersvcurl, groupid)
	err := s.httpsvclient.Get(requesturl, &group)
	if err != nil && err.Error() == "404" {
		err = errors.New(group_not_found)
	}
	return group.Usernames, err
}

// Get user for given id
func (s *service) getUser(ctx context.Context, userid string) (string, error) {
	var user struct {
		Id string `json:"id"`
	}
	requesturl := fmt.Sprintf("%s/users/%s", s.usersvcurl, userid)
	err := s.httpsvclient.Get(requesturl, &user)
	if err != nil && err.Error() == "404" {
		err = errors.New(user_not_found)
	}
	return user.Id, err
}

// Get recipients for the reply to message identified by given message id
func (s *service) getReplyRecipient(ctx context.Context, msgid string) (*replyMessageRecipient, error) {
	record, err := s.repository.GetMessage(ctx, msgid)
	if err != nil {
		return nil, s.mapError(err)
	}
	return &replyMessageRecipient{record.Sender, record.GroupId}, err
}

// Store reply message by deriving recipient from original message
func (s *service) storeReply(ctx context.Context, msg message) (string, error) {

	_, iderr := s.repository.GetMessage(ctx, msg.Re)
	if iderr != nil {
		return "", s.mapError(iderr)
	}

	recipient, err := s.getReplyRecipient(ctx, msg.Re)
	if err != nil {
		return "", s.mapError(err)
	}

	var recipients []string
	if len(recipient.group) > 0 {
		var err error
		recipients, err = s.getGroupUsers(ctx, recipient.group)
		if err != nil {
			return "", s.mapError(err)
		}
	}

	recipients = appendunique(recipients, recipient.sender)

	record := &Record{
		ReplyToMsgId: msg.Re,
		Sender:       msg.Sender,
		Recipients:   recipients,
		GroupId:      recipient.group,
		Subject:      msg.Subject,
		Body:         msg.Body,
	}

	return s.repository.StoreMessage(ctx, record)

}

// Map DB Error to Service Error
func (s *service) mapError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, no_docs_in_result) {
		return ErrMsgNotFound
	}
	if strings.Contains(msg, user_not_found) {
		return ErrUserNotFound
	}
	if strings.Contains(msg, group_not_found) {
		return ErrGroupNotFound
	}
	return ErrSystemError
}

// Append an element to collection if it is not already present.
func appendunique(collection []string, element string) []string {
	for _, item := range collection {
		if item == element {
			return collection
		}
	}
	return append(collection, element)
}

// Map repository records to transport message structure.
func mapRecords(records []Record) []message {
	var messages []message = []message{}
	for _, record := range records {
		messages = append(messages, mapRecord(record))
	}
	return messages
}

// Map repository record to transport message structure.
func mapRecord(record Record) message {
	msg := message{
		Id:        record.Id,
		Re:        record.ReplyToMsgId,
		Sender:    record.Sender,
		Subject:   record.Subject,
		Body:      record.Body,
		Timestamp: record.Timestamp.Format(time.RFC3339),
	}
	if len(record.GroupId) > 0 {
		msg.Recipient.Groupname = record.GroupId
	} else {
		msg.Recipient.Username = record.Recipients[0]
	}
	return msg
}
