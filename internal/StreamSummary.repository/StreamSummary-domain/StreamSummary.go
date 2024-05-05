package StreamSummarydomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StreamSummary struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StreamDuration   int                `json:"StreamDuration" bson:"StreamDuration"`
	AverageViewers   int                `json:"AverageViewers" bson:"AverageViewers"`
	MaxViewers       int                `json:"MaxViewers" bson:"MaxViewers"`
	NewFollowers     int                `json:"NewFollowers" bson:"NewFollowers"`
	NewSubscriptions int                `json:"NewSubscriptions" bson:"NewSubscriptions"`
	Advertisements   int                `json:"Advertisements" bson:"Advertisements"`
	Date             time.Time          `json:"Date" bson:"Date"`
	StreamerID       primitive.ObjectID `json:"StreamerID" bson:"StreamerID"`
}

type UpdateStreamSummary struct {
	AverageViewers   int `json:"AverageViewers" bson:"AverageViewers"`
	MaxViewers       int `json:"MaxViewers" bson:"MaxViewers"`
	NewFollowers     int `json:"NewFollowers" bson:"NewFollowers"`
	NewSubscriptions int `json:"NewSubscriptions" bson:"NewSubscriptions"`
	Advertisements   int `json:"Advertisements" bson:"Advertisements"`
}
