package communitiesdomain

import (
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Community struct {
	ID            primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	CommunityName string               `json:"communityName" bson:"CommunityName" validate:"required"`
	Description   string               `json:"description" bson:"Description"`
	CreatorID     primitive.ObjectID   `json:"creatorID" bson:"CreatorID"`
	Members       []primitive.ObjectID `json:"members" bson:"Members"`
	Mods          []primitive.ObjectID `json:"mods" bson:"Mods"`
	BannedUsers   []primitive.ObjectID `json:"bannedUsers" bson:"BannedUsers"` // Usuarios expulsados
	Rules         string               `json:"rules" bson:"Rules"`
	IsPrivate     bool                 `json:"isPrivate" bson:"IsPrivate"`
	Categories    []string             `json:"categories" bson:"Categories"`
	CreatedAt     time.Time            `json:"createdAt" bson:"CreatedAt"`
	UpdatedAt     time.Time            `json:"updatedAt" bson:"UpdatedAt"`
}

func (u *Community) ValidateReqCreateCommunities() error {

	validate := validator.New()
	return validate.Struct(u)
}

type PostGetCommunityReq struct {
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
		Online   bool   `json:"Online"`
	} `json:"UserInfo"`
	OriginalPostData *PostGetCommunityReq `json:"OriginalPostData"`
	Views            int                  `json:"Views" bson:"Views"`
	IsLikedByID      bool                 `json:"isLikedByID" bson:"isLikedByID"`
	LikeCount        int                  `json:"likeCount" bson:"likeCount"`
	RePostsCount     int                  `json:"RePostsCount" bson:"RePostsCount"`
	CommentsCount    int                  `json:"CommentsCount" bson:"CommentsCount"`
}

type CommunityDetails struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CommunityName string             `json:"communityName" bson:"CommunityName"`
	Description   string             `json:"description" bson:"Description"`
	Creator       CreatorInfo        `json:"creator" bson:"creator"`
	IsPrivate     bool               `json:"isPrivate" bson:"IsPrivate"`
	MembersCount  int                `json:"membersCount" bson:"membersCount"`
	CreatedAt     time.Time          `json:"createdAt" bson:"CreatedAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"UpdatedAt"`
	Categories    []string           `json:"categories" bson:"Categories"`
}

type CreatorInfo struct {
	UserID   primitive.ObjectID `json:"userID" bson:"userID"`
	Avatar   string             `json:"avatar" bson:"avatar"`
	Banner   string             `json:"banner" bson:"banner"`
	NameUser string             `json:"nameUser" bson:"nameUser"`
}
