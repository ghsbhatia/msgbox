package msgstore

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
	Id           string    // Auto Generated: Message Id
	ReplyToMsgId string    // Optional: Id of message to which this message is a reply
	Sender       string    // Sender userid
	Recipients   []string  // Recipient userids
	Subject      string    // Subject
	Body         string    // Message Content
	Timestamp    time.Time // Auto Generated: System time when this message is stored
}

type MessageRepository interface {
	// Store Message: Args: Message
	StoreMessage(*Message) (string, error)
	// Get Messages: Args: User
	GetUserMessages(string) ([]Message, error)
	// Get Message: Args: MessageId
	GetMessage(string) (Message, error)
	// Get Message Replies: Args: MessageId
	GetReplyMessages(string) ([]Message, error)
	// Delete all Content
	Purge() error
}

func NewMessageRepository(dburl string, db string) MessageRepository {
	connection, err := getDBConnction(dburl)
	if err != nil {
		return nil
	}
	return &messageRepository{connection: connection, database: db}
}

const (
	MSGCOLLECTION = "messages"
	REPCOLLECTION = "replies"
)

type messageRepository struct {
	connection interface{}
	database   string
}

type relationship struct {
	MessageId primitive.ObjectID
	ReplyIds  []primitive.ObjectID
}

func (r *messageRepository) StoreMessage(message *Message) (string, error) {

	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(MSGCOLLECTION)

	if len(message.ReplyToMsgId) > 0 {
		_, err := r.GetMessage(message.ReplyToMsgId)
		if err != nil {
			log.Fatal("Unable to get Original Message with id:", message.ReplyToMsgId, err)
			return "", err
		}
	}

	message.Timestamp = time.Now()
	result, err := collection.InsertOne(context.TODO(), message)
	if err != nil {
		log.Fatal(err)
		return "", err
	} else {
		fmt.Println("Inserted a single message document: ", result.InsertedID)
	}

	if len(message.ReplyToMsgId) > 0 {
		originalMsgId, _ := primitive.ObjectIDFromHex(message.ReplyToMsgId)
		replyIds, _ := r.getReplyRelationship(originalMsgId)
		if len(replyIds) == 0 {
			r.storeReplyRelationship(originalMsgId, result.InsertedID.(primitive.ObjectID))
		} else {
			replyIds = append(replyIds, result.InsertedID.(primitive.ObjectID))
			r.updateReplyRelationship(originalMsgId, replyIds)
		}
	}

	return fmt.Sprintf("%s", result.InsertedID.(primitive.ObjectID).Hex()), nil
}

func (r *messageRepository) GetUserMessages(user string) ([]Message, error) {
	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(MSGCOLLECTION)
	var results []Message
	filter := bson.D{{"recipients", user}}
	cur, err := collection.Find(context.TODO(), filter, options.Find())
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	for cur.Next(context.TODO()) {
		fmt.Println("Getting next message from cursor")
		var msg Message
		err := cur.Decode(&msg)
		if err != nil {
			log.Fatal(err)
		} else {
			results = append(results, msg)
		}
	}
	return results, nil
}

func (r *messageRepository) GetMessage(msgid string) (Message, error) {
	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(MSGCOLLECTION)
	result := Message{}
	docId, err := primitive.ObjectIDFromHex(msgid)
	if err == nil {
		err = collection.FindOne(context.TODO(), bson.M{"_id": docId}).Decode(&result)
	}
	if err != nil {
		log.Fatal("error:", err)
	} else {
		result.Id = msgid
	}
	return result, err
}

func (r *messageRepository) GetReplyMessages(msgid string) ([]Message, error) {
	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(MSGCOLLECTION)
	originalMsgId, err := primitive.ObjectIDFromHex(msgid)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var results []Message
	replyIds, _ := r.getReplyRelationship(originalMsgId)
	if len(replyIds) > 0 {
		filter := bson.D{{"_id", bson.D{{"$in", replyIds}}}}
		cur, err := collection.Find(context.TODO(), filter, options.Find())
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		for cur.Next(context.TODO()) {
			fmt.Println("Getting next message from cursor")
			var msg Message
			err := cur.Decode(&msg)
			if err != nil {
				log.Fatal(err)
			} else {
				results = append(results, msg)
			}
		}

	}

	return results, nil
}

func (r *messageRepository) Purge() error {
	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(REPCOLLECTION)
	deleteResult, err := collection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted %v documents in the %s collection\n", deleteResult.DeletedCount, REPCOLLECTION)
	collection = client.Database(r.database).Collection(MSGCOLLECTION)
	deleteResult, err = collection.DeleteMany(context.TODO(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deleted %v documents in the %s collection\n", deleteResult.DeletedCount, MSGCOLLECTION)
	return err
}

func (r *messageRepository) getReplyRelationship(msgid primitive.ObjectID) ([]primitive.ObjectID, error) {
	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(REPCOLLECTION)
	relation := &relationship{}
	err := collection.FindOne(context.TODO(), bson.M{"messageid": msgid}).Decode(&relation)
	return relation.ReplyIds, err
}

func (r *messageRepository) storeReplyRelationship(msgid primitive.ObjectID, replyid primitive.ObjectID) (string, error) {
	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(REPCOLLECTION)
	relation := &relationship{ReplyIds: make([]primitive.ObjectID, 1)}
	relation.MessageId = msgid
	relation.ReplyIds[0] = replyid
	result, err := collection.InsertOne(context.TODO(), relation)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single relationship document: ", result.InsertedID)
	return fmt.Sprintf("%s", result.InsertedID.(primitive.ObjectID).Hex()), nil
}

func (r *messageRepository) updateReplyRelationship(msgId primitive.ObjectID, replyIds []primitive.ObjectID) error {
	client := r.connection.(*mongo.Client)
	collection := client.Database(r.database).Collection(REPCOLLECTION)
	filter := bson.M{"messageid": msgId}
	update := bson.D{{"$set", bson.D{{"replyids", replyIds}}}}
	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
		return err
	}
	fmt.Printf("Matched %v documents and updated %v documents.\n", result.MatchedCount, result.ModifiedCount)
	return nil
}

func getDBConnction(dburl string) (interface{}, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(dburl)
	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	fmt.Println("Connected to MongoDB!")
	return client, nil
}
