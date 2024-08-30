package PinkkerProfitPerMonthdomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Week struct {
	Impressions int     `json:"impressions" bson:"impressions"`
	Clicks      int     `json:"clicks" bson:"clicks"`
	Pixels      float64 `json:"pixels" bson:"pixels"`
}

type PinkkerProfitPerMonth struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	Weeks     map[string]Week    `json:"weeks" bson:"weeks"`
	Total     float64            `json:"total" bson:"total"`
}
