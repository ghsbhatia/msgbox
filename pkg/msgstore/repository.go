package msgstore

import (
	"context"
	"fmt"
	"time"

	"github.com/ghsbhatia/msgbox/pkg/ctxlog"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Record represents message stored in the repository.
type Record struct {
	Id           string    `json:"id,omitempty" bson:"_id,omitempty"`
	ReplyToMsgId string    // Optional: Id of message to which this message is a reply
	Sender       string    // Sender userid
	GroupId      string    // Optional: Id of group if this message recipient is a group
	Recipients   []string  // Recipient userids
	Subject      string    // Subject
	Body         string    // Message Content
	Timestamp    time.Time // Auto Generated: System time when this message is stored
}

// Repository for message persistence
type MessageRepository interface {
	// Store Message: args: message, return: message id
	StoreMessage(context.Context, *Record) (string, error)
	// Get Messages: args: user id, return: messages
	GetUserMessages(context.Context, string) ([]Record, error)
	// Get Message: args: message id, return: message
	GetMessage(context.Context, string) (Record, error)
	// Get Message Replies: args: message id, return: messages
	GetReplyMessages(context.Context, string) ([]Record, error)
	// Delete all Content
	Purge(context.Context) error
}

// Get a new instance of message repository. Args: database url, database name
func NewMessageRepository(dburl string, db string) (MessageRepository, error) {
	connection, err := getDBConnection(dburl)
	if err != nil {
		return nil, err
	}
	return &messageRepository{connection: connection, database: db}, nil
}

const (
	// collection to store messages
	MSGCOLLECTION = "messages"
	// collection to store message relationships so that replies can be correlated
	// to original message
	REPCOLLECTION = "replies"
)

// message repository implementation
type messageRepository struct {
	connection interface{}
	database   string
}

// relationship between original message and reply messages
type relationship struct {
	MessageId primitive.ObjectID
	ReplyIds  []primitive.ObjectID
}

func (r *messageRepository) StoreMessage(ctx context.Context, message *Record) (msgid string, err error) {

	defer func(begin time.Time) {
		logger := log.With(ctxlog.Logger(ctx), "component", "repository")
		logger.Log(
			"method", "store message",
			"subject", message.Subject,
			"id", msgid,
			"re", message.ReplyToMsgId,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())

	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(MSGCOLLECTION)

	// If this message is a reply, original message must exist

	if len(message.ReplyToMsgId) > 0 {
		_, err = r.GetMessage(ctx, message.ReplyToMsgId)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("error finding orginal msg with id %s", message.ReplyToMsgId))
			return "", err
		}
	}

	message.Timestamp = time.Now()
	result, dberr := collection.InsertOne(ctx, message)
	if dberr != nil {
		err = dberr
		return "", err
	}

	msgid = fmt.Sprintf("%s", result.InsertedID.(primitive.ObjectID).Hex())

	// If this message is a reply, create relationship to original message

	if len(message.ReplyToMsgId) > 0 {
		originalMsgId, _ := primitive.ObjectIDFromHex(message.ReplyToMsgId)
		replyIds, _ := r.getReplyRelationship(ctx, originalMsgId)
		if len(replyIds) == 0 {
			r.storeReplyRelationship(ctx, originalMsgId, result.InsertedID.(primitive.ObjectID))
		} else {
			replyIds = append(replyIds, result.InsertedID.(primitive.ObjectID))
			r.updateReplyRelationship(ctx, originalMsgId, replyIds)
		}
	}

	return msgid, nil
}

func (r *messageRepository) GetUserMessages(ctx context.Context, user string) (results []Record, err error) {

	logger := log.With(ctxlog.Logger(ctx), "component", "repository")

	defer func(begin time.Time) {
		logger.Log(
			"method", "get user messages",
			"user", user,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())

	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(MSGCOLLECTION)

	filter := bson.D{{"recipients", user}}
	cursor, dberr := collection.Find(ctx, filter, options.Find())
	if dberr != nil {
		err = dberr
		return nil, err
	}

	for cursor.Next(ctx) {
		var msg Record
		dberr := cursor.Decode(&msg)
		if dberr != nil {
			logger.Log("method", "get user messages", "user", user, "cursor error", dberr)
		} else {
			results = append(results, msg)
		}
	}

	return results, nil

}

func (r *messageRepository) GetMessage(ctx context.Context, msgid string) (result Record, err error) {

	defer func(begin time.Time) {
		logger := log.With(ctxlog.Logger(ctx), "component", "repository")
		logger.Log(
			"method", "get message",
			"id", msgid,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())

	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(MSGCOLLECTION)

	result = Record{}

	docId, oiderr := primitive.ObjectIDFromHex(msgid)
	if oiderr == nil {
		err = collection.FindOne(ctx, bson.M{"_id": docId}).Decode(&result)
		result.Id = msgid
	} else {
		err = oiderr
	}

	return result, err
}

func (r *messageRepository) GetReplyMessages(ctx context.Context, msgid string) (results []Record, err error) {

	logger := log.With(ctxlog.Logger(ctx), "component", "repository")

	defer func(begin time.Time) {
		logger.Log(
			"method", "get reply messages",
			"msgid", msgid,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())

	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(MSGCOLLECTION)

	originalMsgId, oiderr := primitive.ObjectIDFromHex(msgid)
	if oiderr != nil {
		err = oiderr
		return nil, err
	}

	replyIds, _ := r.getReplyRelationship(ctx, originalMsgId)
	if len(replyIds) > 0 {
		filter := bson.D{{"_id", bson.D{{"$in", replyIds}}}}
		cursor, dberr := collection.Find(context.TODO(), filter, options.Find())
		if dberr != nil {
			err = dberr
			return nil, err
		}
		for cursor.Next(context.TODO()) {
			var msg Record
			dberr := cursor.Decode(&msg)
			if dberr != nil {
				logger.Log("method", "get reply messages", "msgid", msgid, "cursor error", dberr)
			} else {
				results = append(results, msg)
			}
		}

	}

	return results, nil
}

func (r *messageRepository) Purge(ctx context.Context) (err error) {

	logger := log.With(ctxlog.Logger(ctx), "component", "repository")

	defer func(begin time.Time) {
		logger.Log(
			"method", "purge",
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())

	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(REPCOLLECTION)

	deleteResult, dberr := collection.DeleteMany(ctx, bson.D{{}})
	if dberr == nil {
		msg := fmt.Sprintf("Deleted %v documents in the %s collection\n", deleteResult.DeletedCount, REPCOLLECTION)
		logger.Log("method", "purge", "info", msg)
	} else {
		err = dberr
	}

	collection = client.Database(r.database).Collection(MSGCOLLECTION)
	deleteResult, dberr = collection.DeleteMany(context.TODO(), bson.D{{}})
	if dberr == nil {
		msg := fmt.Sprintf("Deleted %v documents in the %s collection\n", deleteResult.DeletedCount, MSGCOLLECTION)
		logger.Log("method", "purge", "info", msg)
	} else {
		err = dberr
	}

	return err
}

func (r *messageRepository) getReplyRelationship(ctx context.Context, msgid primitive.ObjectID) (results []primitive.ObjectID, err error) {

	defer func(begin time.Time) {
		logger := log.With(ctxlog.Logger(ctx), "component", "repository")
		logger.Log(
			"method", "get reply relationship",
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())

	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(REPCOLLECTION)

	relation := &relationship{}
	err = collection.FindOne(ctx, bson.M{"messageid": msgid}).Decode(&relation)
	results = relation.ReplyIds

	return results, err
}

func (r *messageRepository) storeReplyRelationship(ctx context.Context, msgid primitive.ObjectID, replyid primitive.ObjectID) (relid string, err error) {

	logger := log.With(ctxlog.Logger(ctx), "component", "repository")

	defer func(begin time.Time) {
		logger.Log(
			"method", "store reply relationship",
			"took", time.Since(begin),
			"relationship id", relid,
			"err", err,
		)
	}(time.Now())

	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(REPCOLLECTION)

	relation := &relationship{ReplyIds: make([]primitive.ObjectID, 1)}
	relation.MessageId = msgid
	relation.ReplyIds[0] = replyid

	result, dberr := collection.InsertOne(ctx, relation)
	if dberr != nil {
		err = dberr
		return "", err
	}

	relid = fmt.Sprintf("%s", result.InsertedID.(primitive.ObjectID).Hex())
	return relid, nil
}

func (r *messageRepository) updateReplyRelationship(ctx context.Context, msgId primitive.ObjectID, replyIds []primitive.ObjectID) (err error) {

	logger := log.With(ctxlog.Logger(ctx), "component", "repository")

	defer func(begin time.Time) {
		logger.Log(
			"method", "update reply relationship",
			"msgid", msgId,
			"took", time.Since(begin),
			"relationship id", msgId,
			"err", err,
		)
	}(time.Now())

	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(REPCOLLECTION)

	filter := bson.M{"messageid": msgId}
	update := bson.D{{"$set", bson.D{{"replyids", replyIds}}}}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		msg := fmt.Sprintf("Matched %v documents and updated %v documents.\n", result.MatchedCount, result.ModifiedCount)
		logger.Log("method", "update reply relationship", "info", msg)
	}

	return err
}

func getDBConnection(dburl string) (interface{}, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(dburl)
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}
