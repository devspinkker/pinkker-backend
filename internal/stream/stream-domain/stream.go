package streamdomain

import (
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Stream struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StreamerID         primitive.ObjectID `json:"streamerId" bson:"StreamerID"`
	Streamer           string             `json:"streamer" bson:"Streamer"`
	StreamerAvatar     string             `json:"streamer_avatar" bson:"StreamerAvatar,omitempty"`
	ViewerCount        int                `json:"ViewerCount"  bson:"ViewerCount,default:0"`
	Online             bool               `json:"online" bson:"Online,default:false"`
	StreamTitle        string             `json:"stream_title" bson:"StreamTitle"`
	StreamCategory     string             `json:"stream_category" bson:"StreamCategory"`
	StreamNotification string             `json:"stream_notification" bson:"StreamNotification"`
	StreamTag          []string           `json:"stream_tag"  bson:"StreamTag,default:['Español']"`
	StreamLikes        []string           `json:"stream_likes" bson:"StreamLikes"`
	StreamIdiom        string             `json:"stream_idiom" default:"Español" bson:"StreamIdiom,default:'Español'"`
	StreamThumbnail    string             `json:"stream_thumbnail" bson:"StreamThumbnail"`
	StartDate          time.Time          `json:"start_date" bson:"StartDate"`
	Timestamp          time.Time          `json:"Timestamp" bson:"Timestamp"`
	EmotesChat         map[string]string  `json:"EmotesChat" bson:"EmotesChat"`
}
type UpdateStreamInfo struct {
	Date         int64    `json:"date"`
	Title        string   `json:"title" validate:"min=5,max=30"`
	Notification string   `json:"notification" validate:"min=5,max=30"`
	Category     string   `json:"category" validate:"min=3"`
	Tag          []string `json:"tag" `
	Idiom        string   `json:"idiom"`
}

func (u *UpdateStreamInfo) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type Update_start_date struct {
	Date int    `json:"date"`
	Key  string `json:"keyTransmission"`
}
type Categoria struct {
	Name       string   `json:"nombre"`
	Img        string   `json:"img,omitempty"`
	Spectators int      `json:"spectators,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	TopColor   []string `json:"TopColor,omitempty"`
}
