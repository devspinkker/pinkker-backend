package clipdomain

import (
	"time"

	"github.com/go-playground/validator"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Clip struct {
	ID              primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	NameUserCreator string               `json:"nameUserCreator" bson:"NameUserCreator"`
	IDCreator       primitive.ObjectID   `json:"idCreator" bson:"IDCreator"`
	NameUser        string               `json:"streamerId" bson:"StreamerID"`
	StreamThumbnail string               `json:"streamThumbnail" bson:"StreamThumbnail"`
	Category        string               `json:"category" bson:"Category"`
	UserID          primitive.ObjectID   `json:"userId" bson:"UserID"`
	Avatar          string               `json:"Avatar" bson:"Avatar"`
	ClipTitle       string               `json:"clipTitle" bson:"ClipTitle"`
	URL             string               `json:"url" bson:"url"`
	Likes           []primitive.ObjectID `json:"likes" bson:"likes"`
	Duration        int                  `json:"duration" bson:"duration"`
	Views           int                  `json:"views" bson:"views"`
	Cover           string               `json:"cover" bson:"cover"`
	Comments        []primitive.ObjectID `json:"comments" bson:"Comments"`
	Timestamps      struct {
		CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
		UpdatedAt time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	} `json:"timestamps,omitempty" bson:"timestamps,omitempty"`
}
type ClipRequest struct {
	Video     []byte `json:"video" validate:"required"`
	Start     int    `json:"start" validate:"required"`
	End       int    `json:"end" validate:"required"`
	ClipTitle string `json:"clipTitle" validate:"required,min=2,max=100"`
	TotalKey  string `json:"totalKey" validate:"required"`
}

func (u *ClipRequest) ValidateClipRequest() error {
	if err := validateDuration(u.Start, u.End); err != nil {
		return err
	}

	validate := validator.New()
	return validate.Struct(u)
}

func validateDuration(start, end int) error {
	if start < 0 || start > end || end < 10 {
		return errors.New("start y end deben ser valores no negativos y start debe ser menor o igual que end, y end debe ser mayor o igual a 10")
	}

	duration := end - start
	minDuration := 10
	maxDuration := 60
	if duration < minDuration || duration > maxDuration {
		return errors.New("la duración del clip debe estar entre 10 y 60 segundos")
	}

	return nil
}

type GetClipId struct {
	ClipId primitive.ObjectID `json:"ClipId" validate:"required"`
}

func (u *GetClipId) GetClipIdValidate() error {
	validate := validator.New()
	return validate.Struct(u)
}
