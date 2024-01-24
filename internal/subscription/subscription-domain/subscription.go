package subscriptiondomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty"`
	SubscriptionNameUser string             `bson:"SubscriptionNameUser"`
	SourceUserID         primitive.ObjectID `bson:"sourceUserID"`
	DestinationUserID    primitive.ObjectID `bson:"destinationUserID"`
	SubscriptionStart    time.Time          `bson:"SubscriptionStart"`
	SubscriptionEnd      time.Time          `bson:"SubscriptionEnd"`
	MonthsSubscribed     int                `bson:"MonthsSubscribed"`
	Notified             bool               `bson:"Notified"`
	Text                 string             `bson:"Text"`
}

type Subscriber struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	SubscriberNameUser string             `bson:"SubscriberNameUser"`
	SourceUserID       primitive.ObjectID `bson:"sourceUserID"`
	DestinationUserID  primitive.ObjectID `bson:"destinationUserID"`
	SubscriptionStart  time.Time          `bson:"SubscriptionStart"`
	SubscriptionEnd    time.Time          `bson:"SubscriptionEnd"`
	Notified           bool               `bson:"Notified"`
	Text               string             `bson:"Text"`
}
type ResSubscriber struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	SubscriberNameUser string             `bson:"SubscriberNameUser"`
	SourceUserID       primitive.ObjectID `bson:"sourceUserID"`
	DestinationUserID  primitive.ObjectID `bson:"destinationUserID"`
	SubscriptionStart  time.Time          `bson:"SubscriptionStart"`
	SubscriptionEnd    time.Time          `bson:"SubscriptionEnd"`
	Notified           bool               `bson:"Notified"`
	Text               string             `bson:"Text"`
	FromUserInfo       struct {
		Avatar   string `json:"Avatar"`
		NameUser string `json:"NameUser"`
	} `json:"FromUserInfo"`
}
type ReqCreateSuscribirse struct {
	ToUser primitive.ObjectID `json:"ToUser" validate:"required"`
	Text   string             `json:"Text" validate:"required,min=2,max=70"`
}
