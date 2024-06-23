package Chatsdomain

import "time"

type Message struct {
	ID         string    `bson:"_id,omitempty"`
	SenderID   string    `bson:"sender_id"`
	ReceiverID string    `bson:"receiver_id"`
	Content    string    `bson:"content"`
	Seen       bool      `bson:"seen"`
	Notified   bool      `bson:"notified"`
	CreatedAt  time.Time `bson:"created_at"`
}

type Chat struct {
	ID         string    `bson:"_id,omitempty"`
	User1ID    string    `bson:"user1_id"`
	User2ID    string    `bson:"user2_id"`
	MessageIDs []string  `bson:"message_ids"`
	CreatedAt  time.Time `bson:"created_at"`
}
