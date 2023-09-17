package tweetdomain

import (
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Post struct {
	ID        primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Status    string               `json:"Status" bson:"Status"`
	PostImage string               `json:"PostImage" bson:"PostImage,omitempty"`
	TimeStamp time.Time            `json:"TimeStamp" bson:"TimeStamp"`
	UserID    primitive.ObjectID   `json:"UserID" bson:"UserID"`
	Likes     []primitive.ObjectID `json:"Likes" bson:"Likes"`
	Comments  []primitive.ObjectID `json:"Comments" bson:"Comments"`
	RePosts   []primitive.ObjectID `json:"RePosts" bson:"RePosts"`
}
type PostComment struct {
	ID        primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	CommentTo primitive.ObjectID   `json:"CommentTo" bson:"CommentTo"`
	Comments  []primitive.ObjectID `json:"Comments" bson:"Comments"`
	Status    string               `json:"Status" bson:"Status"`
	PostImage string               `json:"PostImage" bson:"PostImage,omitempty"`
	TimeStamp time.Time            `json:"TimeStamp" bson:"TimeStamp"`
	UserID    primitive.ObjectID   `json:"UserID" bson:"UserID"`
	Likes     []primitive.ObjectID `json:"Likes" bson:"Likes"`
	RePosts   []primitive.ObjectID `json:"RePosts" bson:"RePosts"`
}
type RePost struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"UserID" bson:"UserID"`
	OriginalPost primitive.ObjectID `json:"OriginalPost" bson:"OriginalPost"`
	TimeStamp    time.Time          `json:"TimeStamp" bson:"TimeStamp"`
}
type CitaPost struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID       primitive.ObjectID `json:"UserID" bson:"UserID"`
	OriginalPost primitive.ObjectID `json:"OriginalPost" bson:"OriginalPost"`
	TimeStamp    time.Time          `json:"TimeStamp" bson:"TimeStamp"`
	Status       string             `json:"Status" bson:"Status"`
}
type TweetModelValidator struct {
	Status string `json:"status" validate:"required,min=3,max=100"`
}
type TweetCommentModelValidator struct {
	Status    string             `json:"status" validate:"required,min=3,max=100"`
	CommentTo primitive.ObjectID `json:"CommentTo" validate:"required"`
}
type CitaPostModelValidator struct {
	Status       string             `json:"status" validate:"required,min=3,max=100"`
	OriginalPost primitive.ObjectID `json:"OriginalPost" validate:"required"`
}

func (u *TweetCommentModelValidator) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}
func (u *TweetModelValidator) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}
func (u *CitaPostModelValidator) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type TweetGetFollowReq struct {
	ID        primitive.ObjectID   `json:"id"`
	Status    string               `json:"Status"`
	PostImage string               `json:"PostImage,omitempty"`
	TimeStamp time.Time            `json:"TimeStamp"`
	UserID    primitive.ObjectID   `json:"UserID"`
	Likes     []primitive.ObjectID `json:"Likes"`
	UserInfo  struct {
		FullName string `json:"FullName"`
		Avatar   string `json:"Avatar"`
		NameUser string `json:"NameUser"`
	} `json:"UserInfo"`
}
type TweetCommentsGetReq struct {
	ID        primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	CommentBy primitive.ObjectID   `json:"CommentBy" bson:"CommentBy"`
	Comments  []primitive.ObjectID `json:"Comments" bson:"Comments"`
	Status    string               `json:"Status" bson:"Status"`
	PostImage string               `json:"PostImage" bson:"PostImage,omitempty"`
	TimeStamp time.Time            `json:"TimeStamp" bson:"TimeStamp"`
	UserID    primitive.ObjectID   `json:"UserID" bson:"UserID"`
	Likes     []primitive.ObjectID `json:"Likes" bson:"Likes"`
	UserInfo  struct {
		FullName string `json:"FullName"`
		Avatar   string `json:"Avatar"`
		NameUser string `json:"NameUser"`
	} `json:"UserInfo"`
}
