package voddomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vod struct {
	StreamerId         primitive.ObjectID `json:"streamerId" bson:"StreamerId"`
	Streamer           string             `json:"streamer" bson:"Streamer"`
	URL                string             `json:"url" bson:"URL"`
	StreamTitle        string             `json:"stream_title" bson:"StreamTitle"`
	StreamCategory     string             `json:"stream_category" bson:"StreamCategory"`
	StreamNotification string             `json:"stream_notification" bson:"StreamNotification"`
	StreamTag          []string           `json:"stream_tag,omitempty" bson:"StreamTag,omitempty"`
	StreamIdiom        string             `json:"stream_idiom,omitempty" bson:"StreamIdiom,omitempty"`
	StreamThumbnail    string             `json:"stream_thumbnail" bson:"StreamThumbnail"`
	StartDate          time.Time          `json:"start_date" bson:"StartDate"`
	Views              int                `json:"views,omitempty" bson:"Views,omitempty"`
	Likes              []string           `json:"likes,omitempty" bson:"Likes,omitempty"`
	Dislikes           []string           `json:"dislikes,omitempty" bson:"Dislikes,omitempty"`
	Points             int                `json:"points,omitempty" bson:"Points,omitempty"`
	Timestamps         struct {
		CreatedAt int64 `json:"createdAt,omitempty" bson:"CreatedAt,omitempty"`
		UpdatedAt int64 `json:"updatedAt,omitempty" bson:"UpdatedAt,omitempty"`
	} `json:"timestamps,omitempty" bson:"Timestamps,omitempty"`
}
