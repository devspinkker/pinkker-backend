package Chatsdomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	SenderID   primitive.ObjectID `bson:"sender_id"`
	ReceiverID primitive.ObjectID `bson:"receiver_id"`
	Content    string             `bson:"content"`
	Seen       bool               `bson:"seen"`
	Notified   bool               `bson:"notified"`
	CreatedAt  time.Time          `bson:"created_at"`
}

const (
	REQUESTS  = "request"
	PRIMARY   = "primary"
	SECONDARY = "secondary"
)

type Chat struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty"`
	User1ID     primitive.ObjectID   `bson:"user1_id"`
	User2ID     primitive.ObjectID   `bson:"user2_id"`
	MessageIDs  []primitive.ObjectID `bson:"message_ids"`
	CreatedAt   time.Time            `bson:"created_at"`
	LastMessage time.Time            `bson:"LastMessage"`
	NotifyA     primitive.ObjectID   `bson:"NotifyA"`
	StatusUser1 string               `bson:"status_user1"` // estado en el que esta el chat desde el usuario 1s
	StatusUser2 string               `bson:"status_user2"`
	Blocked     struct {
		BlockedByUser1 bool `bson:"blocked_by_user1"` // el usuario 1 bloqua
		BlockedByUser2 bool `bson:"blocked_by_user2"`
	} `bson:"blocked"`
}

type ChatWithUsers struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	User1ID primitive.ObjectID `bson:"user1_id"`
	User2ID primitive.ObjectID `bson:"user2_id"`
	// MessageIDs  []primitive.ObjectID `bson:"message_ids"`
	CreatedAt   time.Time          `bson:"created_at"`
	Users       []*User            `bson:"users"`
	LastMessage time.Time          `bson:"LastMessage"`
	StatusUser1 string             `bson:"status_user1"`
	StatusUser2 string             `bson:"status_user2"`
	NotifyA     primitive.ObjectID `bson:"NotifyA"`
	Blocked     struct {
		BlockedByUser1 bool `bson:"blocked_by_user1"`
		BlockedByUser2 bool `bson:"blocked_by_user2"`
	} `bson:"blocked"`
}

type User struct {
	ID       string  `bson:"_id,omitempty"`
	NameUser string  `bson:"NameUser"`
	Avatar   string  `bson:"Avatar"`
	Partner  Partner `bson:"Partner"`
}

type Partner struct {
	Active bool      `bson:"Active,omitempty"`
	Date   time.Time `bson:"Date,omitempty"`
}
