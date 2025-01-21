package clipdomain

import (
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Clip struct {
	ID              primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	NameUserCreator string               `json:"nameUserCreator" bson:"NameUserCreator"`
	IDCreator       primitive.ObjectID   `json:"idCreator" bson:"IDCreator"`
	NameUser        string               `json:"NameUser" bson:"NameUser"`
	StreamThumbnail string               `json:"streamThumbnail" bson:"StreamThumbnail"`
	Category        string               `json:"category" bson:"Category"`
	UserID          primitive.ObjectID   `json:"UserID" bson:"UserID"`
	Avatar          string               `json:"Avatar" bson:"Avatar"`
	ClipTitle       string               `json:"clipTitle" bson:"ClipTitle"`
	URL             string               `json:"url" bson:"url"`
	URLS            []string             `json:"urls" bson:"urls"`
	Likes           []primitive.ObjectID `json:"likes" bson:"Likes"`
	Duration        int                  `json:"duration" bson:"duration"`
	Views           int                  `json:"views" bson:"views"`
	Cover           string               `json:"cover" bson:"cover"`
	Comments        []primitive.ObjectID `json:"comments" bson:"Comments"`
	Type            string               `json:"Type" bson:"Type"`
	Timestamps      struct {
		CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
		UpdatedAt time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	} `json:"timestamps,omitempty" bson:"timestamps,omitempty"`
	IdOfTheUsersWhoViewed []primitive.ObjectID `json:"IdOfTheUsersWhoViewed" bson:"IdOfTheUsersWhoViewed"`
	M3U8Content           string               `json:"m3u8Content" bson:"m3u8Content"` // Campo para almacenar el m3u8

}
type ClipCategoryInfo struct {
	Category string `bson:"Category"`
}

type GetClip struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	NameUserCreator string             `json:"nameUserCreator" bson:"NameUserCreator"`
	IDCreator       primitive.ObjectID `json:"idCreator" bson:"IDCreator"`
	NameUser        string             `json:"NameUser" bson:"NameUser"`
	StreamThumbnail string             `json:"streamThumbnail" bson:"StreamThumbnail"`
	Category        string             `json:"category" bson:"Category"`
	UserID          primitive.ObjectID `json:"UserID" bson:"UserID"`
	Avatar          string             `json:"Avatar" bson:"Avatar"`
	ClipTitle       string             `json:"clipTitle" bson:"ClipTitle"`
	URL             string             `json:"url" bson:"url"`
	Duration        int                `json:"duration" bson:"duration"`
	Views           int                `json:"views" bson:"views"`
	Cover           string             `json:"cover" bson:"cover"`
	Timestamps      struct {
		CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
		UpdatedAt time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	} `json:"timestamps,omitempty" bson:"timestamps,omitempty"`
	IsLikedByID   bool               `json:"isLikedByID" bson:"isLikedByID"`
	LikeCount     int                `json:"likeCount" bson:"likeCount"`
	Type          string             `json:"Type" bson:"Type"`
	CommentsCount int                `json:"CommentsCount" bson:"CommentsCount"`
	AdId          primitive.ObjectID `json:"AdId" bson:"AdId"`
	M3U8Content   string             `json:"m3u8Content" bson:"m3u8Content"` // Campo para almacenar el m3u8

}

type ClipRequest struct {
	TsUrls    []string `json:"tsUrls"`
	ClipTitle string   `json:"clipTitle" validate:"required,min=2,max=100"`
	TotalKey  string   `json:"totalKey" validate:"required"`
}

func (u *ClipRequest) ValidateClipRequest() error {

	if len(u.TsUrls) > 8 || len(u.TsUrls) < 2 {
		return errors.New("TsUrls cannot have more than 5 elements")
	}

	const urlPrefix = "https://www.pinkker.tv/8002/stream"
	for _, url := range u.TsUrls {
		if !strings.HasPrefix(url, urlPrefix) {
			return errors.New("all TsUrls must start with " + urlPrefix)
		}
	}

	// Validar otros campos usando el validador estándar
	validate := validator.New()
	return validate.Struct(u)
}

type GetClipId struct {
	ClipId primitive.ObjectID `json:"ClipId" validate:"required"`
}

func (u *GetClipId) GetClipIdValidate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type GetRecommended struct {
	ExcludeIDs []primitive.ObjectID `json:"ExcludeIDs" validate:"required"`
}

func (u *GetRecommended) GetRecommended() error {
	validate := validator.New()
	if reflect.TypeOf(u.ExcludeIDs).Elem() != reflect.TypeOf(primitive.ObjectID{}) {
		return errors.New("Clip debe ser del tipo primitive.ObjectID")
	}
	return validate.Struct(u)
}

type ClipComment struct {
	ID        primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	ClipID    primitive.ObjectID   `json:"clipId,omitempty" bson:"clipId,omitempty"`
	UserID    primitive.ObjectID   `json:"UserID,omitempty" bson:"UserID,omitempty"`
	NameUser  string               `json:"NameUser,omitempty" bson:"nameUser,omitempty"`
	Comment   string               `json:"comment,omitempty" bson:"comment,omitempty"`
	CreatedAt time.Time            `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	Likes     []primitive.ObjectID `json:"likes" bson:"Likes"`
}

type CommentClip struct {
	CommentClip string             `json:"CommentClip" validate:"required,min=2,max=100"`
	IdClip      primitive.ObjectID `json:"IdClip" validate:"required"`
}
type CommentClipId struct {
	IdClip primitive.ObjectID `json:"IdClip" validate:"required"`
}

func (u *CommentClip) ValidateCommentClip() error {

	validate := validator.New()
	return validate.Struct(u)
}

type ClipCommentGet struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ClipID      primitive.ObjectID `json:"clipId,omitempty" bson:"clipId,omitempty"`
	UserID      primitive.ObjectID `json:"UserID" bson:"UserID"`
	NameUser    string             `json:"NameUser" bson:"nameUser"`
	Comment     string             `json:"comment,omitempty" bson:"comment,omitempty"`
	CreatedAt   time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	FullName    string             `json:"FullName"`
	IsLikedByID bool               `json:"isLikedByID" bson:"isLikedByID"`
	LikeCount   int                `json:"likeCount" bson:"likeCount"`
	Avatar      string             `json:"Avatar"`
}
