package tweetdomain

import (
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tweet struct {
	ID         primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Status     string               `json:"Status" bson:"Status"`
	TweetImage string               `json:"TweetImage" bson:"TweetImage,omitempty"`
	TimeStamp  time.Time            `json:"TimeStamp" bson:"TimeStamp"`
	UserID     primitive.ObjectID   `json:"UserID" bson:"UserID"`
	Likes      []primitive.ObjectID `json:"Likes" bson:"Likes"`
}

type TweetModelValidator struct {
	Status string `json:"status" validate:"required,min=3,max=100"`
}

func (u *TweetModelValidator) ValidateUser() error {
	validate := validator.New()
	return validate.Struct(u)
}

type TweetGetFollowReq struct {
	ID         primitive.ObjectID   `json:"id"`
	Status     string               `json:"Status"`
	TweetImage string               `json:"TweetImage,omitempty"`
	TimeStamp  time.Time            `json:"TimeStamp"`
	UserID     primitive.ObjectID   `json:"UserID"`
	Likes      []primitive.ObjectID `json:"Likes"`
	UserInfo   struct {
		FullName string `json:"FullName"`
		Avatar   string `json:"Avatar"`
		NameUser string `json:"NameUser"`
	} `json:"UserInfo"`
}
