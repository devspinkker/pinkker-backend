package StreamSummarydomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StreamSummary struct {
	Title                string             `json:"Title" bson:"Title"`
	StreamThumbnail      string             `json:"StreamThumbnail" bson:"StreamThumbnail"`
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
	StreamCategory       string             `json:"stream_category" bson:"StreamCategory"`
	Admoney              float64            `json:"Admoney" bson:"Admoney"`
	SubscriptionsMoney   float64            `json:"SubscriptionsMoney" bson:"SubscriptionsMoney"`
	DonationsMoney       float64            `json:"DonationsMoney" bson:"DonationsMoney"`
	TotalMoney           float64            `json:"TotalMoney" bson:"TotalMoney"`
}
type StreamSummaryGet struct {
	Title               string             `json:"Title" bson:"Title"`
	StreamThumbnail     string             `json:"StreamThumbnail" bson:"StreamThumbnail"`
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EndOfStream         time.Time          `json:"EndOfStream" bson:"EndOfStream"`
	MaxViewers          int                `json:"MaxViewers" bson:"MaxViewers"`
	StartOfStream       time.Time          `json:"StartOfStream" bson:"StartOfStream"`
	StreamerID          primitive.ObjectID `json:"StreamerID" bson:"StreamerID"`
	StartFollowersCount int                `json:"StartFollowersCount" bson:"StartFollowersCount"`
	StartSubsCount      int                `json:"StartSubsCount" bson:"StartSubsCount"`
	StreamCategory      string             `json:"stream_category" bson:"StreamCategory"`
	UserInfo            UserInfo           `json:"UserInfo" bson:"UserInfo"`
}

type UserInfo struct {
	Avatar   string `json:"Avatar" bson:"Avatar"`
	FullName string `json:"FullName" bson:"FullName"`
	NameUser string `json:"NameUser" bson:"NameUser"`
}

type UpdateStreamSummary struct {
	AverageViewers   int `json:"AverageViewers" bson:"AverageViewers"`
	MaxViewers       int `json:"MaxViewers" bson:"MaxViewers"`
	NewFollowers     int `json:"NewFollowers" bson:"NewFollowers"`
	NewSubscriptions int `json:"NewSubscriptions" bson:"NewSubscriptions"`
	Advertisements   int `json:"Advertisements" bson:"Advertisements"`
}
type AddAds struct {
	StreamerID       primitive.ObjectID `json:"StreamerID" bson:"StreamerID"`
	AdvertisementsId primitive.ObjectID `json:"AdvertisementsId" bson:"AdvertisementsId"`
}
type AverageViewers struct {
	StreamerID primitive.ObjectID `json:"StreamerID" bson:"StreamerID"`
}

type Earnings struct {
	Admoney            float64 `json:"Admoney"`
	SubscriptionsMoney float64 `json:"SubscriptionsMoney"`
	DonationsMoney     float64 `json:"DonationsMoney"`
}
type EarningsPerDay struct {
	Date     time.Time
	Earnings Earnings
}
