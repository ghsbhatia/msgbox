package msgstore

import (
	"context"
	"fmt"
	"time"

	"github.com/ghsbhatia/msgbox/pkg/svcclient"
)

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

	if len(msg.Re) > 0 {
		return s.storeReply(ctx, msg)
	}

	// If message recipient is a group, get users for the group and store them as
	// recipients

	recipients := []string{}
	if len(msg.Recipient.Groupname) > 0 {
		grpusers, err := s.getGroupUsers(ctx, msg.Recipient.Groupname)
		if err != nil {
			return "", err
		}
		recipients = grpusers
	} else if len(msg.Recipient.Username) > 0 {
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
	return msg, err
}

// Get messages for a given user
func (s *service) GetMessages(ctx context.Context, userid string) ([]message, error) {
	records, err := s.repository.GetUserMessages(ctx, userid)
	msgs := mapRecords(records)
	return msgs, err
}

// Get reply messages for message identified by given message id
func (s *service) GetReplies(ctx context.Context, msgid string) ([]message, error) {
	records, err := s.repository.GetReplyMessages(ctx, msgid)
	msgs := mapRecords(records)
	return msgs, err
}

// Get users for group identified by given group id
func (s *service) getGroupUsers(ctx context.Context, groupid string) ([]string, error) {
	var group struct {
		Groupname string   `json:"groupname"`
		Usernames []string `json:"usernames"`
	}
	requesturl := fmt.Sprintf("%s/groups/%s", s.usersvcurl, groupid)
	err := s.httpsvclient.Get(requesturl, &group)
	return group.Usernames, err
}

// Get recipients for the reply to message identified by given message id
func (s *service) getReplyRecipient(ctx context.Context, msgid string) (*replyMessageRecipient, error) {
	record, err := s.repository.GetMessage(ctx, msgid)
	if err != nil {
		return nil, err
	}
	return &replyMessageRecipient{record.Sender, record.GroupId}, err
}

// Store reply message by deriving recipient from original message
func (s *service) storeReply(ctx context.Context, msg message) (string, error) {

	recipient, err := s.getReplyRecipient(ctx, msg.Re)
	if err != nil {
		return "", err
	}

	var recipients []string
	if len(recipient.group) > 0 {
		var err error
		recipients, err = s.getGroupUsers(ctx, recipient.group)
		if err != nil {
			return "", err
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
