package PinkkerProfitPerMonthdomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Week struct {
	Impressions            int     `json:"impressions" bson:"impressions"`
	Clicks                 int     `json:"clicks" bson:"clicks"`
	Pixeles                float64 `json:"pixels" bson:"pixeles"`
	PinkkerPrime           float64 `json:"PinkkerPrime" bson:"pinkkerPrime"`
	CommunityBuy           float64 `json:"communityBuy" bson:"communityBuy"`
	PaidCommunities        float64 `json:"PaidCommunities" bson:"PaidCommunities"`
	CommissionsSuscripcion float64 `json:"CommissionsSuscripcion" bson:"CommissionsSuscripcion"`
	CommissionsDonation    float64 `json:"CommissionsDonation" bson:"CommissionsDonation"`
	CommissionsCommunity   float64 `json:"CommissionsCommunity" bson:"CommissionsCommunity"`
}

type PinkkerProfitPerMonth struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
	Weeks     map[string]Week    `json:"weeks" bson:"weeks"`
	Total     float64            `json:"total" bson:"total"`
}

func NewDefaultWeek() Week {
	return Week{
		Impressions:            0,
		Clicks:                 0,
		Pixeles:                0.0,
		PinkkerPrime:           0,
		CommunityBuy:           0,
		PaidCommunities:        0,
		CommissionsSuscripcion: 0,
		CommissionsDonation:    0,
		CommissionsCommunity:   0,
	}
}
