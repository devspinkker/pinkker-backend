package StreamSummarydomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StreamSummary struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EndOfStream          time.Time          `json:"EndOfStream" bson:"EndOfStream"`
	AverageViewers       int                `json:"AverageViewers" bson:"AverageViewers"`
	AverageViewersByTime map[string]int     `json:"AverageViewersByTime" bson:"AverageViewersByTime"`
	MaxViewers           int                `json:"MaxViewers" bson:"MaxViewers"`
	NewFollowers         int                `json:"NewFollowers" bson:"NewFollowers"`
	NewSubscriptions     int                `json:"NewSubscriptions" bson:"NewSubscriptions"`
	Advertisements       int                `json:"Advertisements" bson:"Advertisements"`
	StartOfStream        time.Time          `json:"StartOfStream" bson:"StartOfStream"`
	StreamerID           primitive.ObjectID `json:"StreamerID" bson:"StreamerID"`
	StartFollowersCount  int                `json:"StartFollowersCount" bson:"StartFollowersCount"`
	StartSubsCount       int                `json:"StartSubsCount" bson:"StartSubsCount"`
}
type UpdateStreamSummary struct {
	AverageViewers   int `json:"AverageViewers" bson:"AverageViewers"`
	MaxViewers       int `json:"MaxViewers" bson:"MaxViewers"`
	NewFollowers     int `json:"NewFollowers" bson:"NewFollowers"`
	NewSubscriptions int `json:"NewSubscriptions" bson:"NewSubscriptions"`
	Advertisements   int `json:"Advertisements" bson:"Advertisements"`
}
type AddAds struct {
	StreamerID primitive.ObjectID `json:"StreamerID" bson:"StreamerID"`
}
type AverageViewers struct {
	StreamerID primitive.ObjectID `json:"StreamerID" bson:"StreamerID"`
}