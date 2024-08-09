package tweetdomain

import (
	"errors"
	"reflect"
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Trend struct {
	Hashtag  string    `json:"hashtag" bson:"Hashtags"`
	Count    int       `json:"count"  bson:"count"`
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
}
type Post struct {
	ID        primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Type      string               `json:"Type" default:"Post" bson:"Type"`
	Status    string               `json:"Status" bson:"Status"`
	PostImage string               `json:"PostImage" bson:"PostImage"`
	TimeStamp time.Time            `json:"TimeStamp" bson:"TimeStamp"`
	UserID    primitive.ObjectID   `json:"UserID" bson:"UserID"`
	Likes     []primitive.ObjectID `json:"Likes" bson:"Likes"`
	Comments  []primitive.ObjectID `json:"Comments" bson:"Comments"`
	RePosts   []primitive.ObjectID `json:"RePosts" bson:"RePosts"`
	Hashtags  []string             `json:"hashtags" bson:"Hashtags"`
	Views     int                  `json:"Views" bson:"Views"`
}
type PostComment struct {
	ID           primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Type         string               `json:"Type" default:"PostComment" bson:"Type"`
	OriginalPost primitive.ObjectID   `json:"OriginalPost" bson:"OriginalPost"`
	Comments     []primitive.ObjectID `json:"Comments" bson:"Comments"`
	Status       string               `json:"Status" bson:"Status"`
	PostImage    string               `json:"PostImage" bson:"PostImage,omitempty"`
	TimeStamp    time.Time            `json:"TimeStamp" bson:"TimeStamp"`
	UserID       primitive.ObjectID   `json:"UserID" bson:"UserID"`
	Likes        []primitive.ObjectID `json:"Likes" bson:"Likes"`
	RePosts      []primitive.ObjectID `json:"RePosts" bson:"RePosts"`
	Hashtags     []string             `json:"hashtags" bson:"Hashtags"`
	Views        int                  `json:"Views" bson:"Views"`
}
type RePost struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type         string             `json:"Type" default:"RePost" bson:"Type"`
	UserID       primitive.ObjectID `json:"UserID" bson:"UserID"`
	OriginalPost primitive.ObjectID `json:"OriginalPost" bson:"OriginalPost"`
	TimeStamp    time.Time          `json:"TimeStamp" bson:"TimeStamp"`
}
type CitaPost struct {
	ID           primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Type         string               `json:"Type" default:"CitaPost" bson:"Type"`
	UserID       primitive.ObjectID   `json:"UserID" bson:"UserID"`
	OriginalPost primitive.ObjectID   `json:"OriginalPost" bson:"OriginalPost"`
	TimeStamp    time.Time            `json:"TimeStamp" bson:"TimeStamp"`
	Status       string               `json:"Status" bson:"Status"`
	Likes        []primitive.ObjectID `json:"Likes" bson:"Likes"`
	RePosts      []primitive.ObjectID `json:"RePosts" bson:"RePosts"`
	Comments     []primitive.ObjectID `json:"Comments" bson:"Comments"`
	PostImage    string               `json:"PostImage" bson:"PostImage"`
	Hashtags     []string             `json:"hashtags" bson:"Hashtags"`
	Views        int                  `json:"Views" bson:"Views"`
}
type TweetModelValidator struct {
	Status string `json:"status" validate:"required,min=3,max=100"`
}
type TweetCommentModelValidator struct {
	Status       string             `json:"status" validate:"required,min=3,max=100"`
	OriginalPost primitive.ObjectID `json:"OriginalPost" validate:"required"`
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

type OriginalPostReference struct {
	ID     primitive.ObjectID `json:"id"`
	Status string             `json:"Status"`
}

type TweetGetFollowReq struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Status    string             `json:"Status" bson:"Status"`
	PostImage string             `json:"PostImage" bson:"PostImage"`
	TimeStamp time.Time          `json:"TimeStamp"  bson:"TimeStamp"`
	UserID    primitive.ObjectID `json:"UserID" bson:"UserID"`
	// Comments     []primitive.ObjectID `json:"Comments" bson:"Comments"`
	OriginalPost primitive.ObjectID `json:"OriginalPost"`
	Type         string             `json:"Type" bson:"Type"`
	Hashtags     []string           `json:"hashtags" bson:"Hashtags"`
	UserInfo     struct {
		FullName string `json:"FullName"`
		Avatar   string `json:"Avatar"`
		NameUser string `json:"NameUser"`
	} `json:"UserInfo"`
	OriginalPostData *TweetGetFollowReq `json:"OriginalPostData"`
	Views            int                `json:"Views" bson:"Views"`
	IsLikedByID      bool               `json:"isLikedByID" bson:"isLikedByID"`
	LikeCount        int                `json:"likeCount" bson:"likeCount"`
	RePostsCount     int                `json:"RePostsCount" bson:"RePostsCount"`
	CommentsCount    int                `json:"CommentsCount" bson:"CommentsCount"`
}
type TweetCommentsGetReq struct {
	ID           primitive.ObjectID   `json:"_id" bson:"_id"`
	Status       string               `json:"Status" bson:"Status"`
	PostImage    string               `json:"PostImage" bson:"PostImage"`
	TimeStamp    time.Time            `json:"TimeStamp"  bson:"TimeStamp"`
	UserID       primitive.ObjectID   `json:"UserID" bson:"UserID"`
	Likes        []primitive.ObjectID `json:"Likes"`
	Comments     []primitive.ObjectID `json:"Comments" bson:"Comments"`
	RePosts      []primitive.ObjectID `json:"RePosts" bson:"RePosts"`
	OriginalPost primitive.ObjectID   `json:"OriginalPost"`
	Type         string               `json:"Type" bson:"Type"`
	Hashtags     []string             `json:"hashtags" bson:"hashtags"`
	Views        int                  `json:"Views" bson:"Views"`
	UserInfo     struct {
		FullName string `json:"FullName"`
		Avatar   string `json:"Avatar"`
		NameUser string `json:"NameUser"`
	} `json:"UserInfo"`
}
type PostAds struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Status    string             `json:"Status" bson:"Status"`
	PostImage string             `json:"PostImage" bson:"PostImage"`
	TimeStamp time.Time          `json:"TimeStamp"  bson:"TimeStamp"`
	UserID    primitive.ObjectID `json:"UserID" bson:"UserID"`
	// Comments     []primitive.ObjectID `json:"Comments" bson:"Comments"`
	Type     string   `json:"Type" bson:"Type"`
	Hashtags []string `json:"hashtags" bson:"Hashtags"`
	UserInfo struct {
		FullName string `json:"FullName"`
		Avatar   string `json:"Avatar"`
		NameUser string `json:"NameUser"`
	} `json:"UserInfo"`
	OriginalPostData *TweetGetFollowReq `json:"OriginalPostData"`
	Views            int                `json:"Views" bson:"Views"`
	IsLikedByID      bool               `json:"isLikedByID" bson:"isLikedByID"`
	LikeCount        int                `json:"likeCount" bson:"likeCount"`
	RePostsCount     int                `json:"RePostsCount" bson:"RePostsCount"`
	CommentsCount    int                `json:"CommentsCount" bson:"CommentsCount"`
	//ads
	AdvertisementsId primitive.ObjectID `json:"AdvertisementsId" bson:"AdvertisementsId"`
	ReferenceLink    string             `json:"ReferenceLink" bson:"ReferenceLink"`
}
type GetRecommended struct {
	ExcludeIDs []primitive.ObjectID `json:"ExcludeIDs" validate:"required"`
}

func (u *GetRecommended) GetRecommended() error {
	validate := validator.New()
	if reflect.TypeOf(u.ExcludeIDs).Elem() != reflect.TypeOf(primitive.ObjectID{}) {
		return errors.New("clip debe ser del tipo primitive.ObjectId")
	}
	return validate.Struct(u)
}
