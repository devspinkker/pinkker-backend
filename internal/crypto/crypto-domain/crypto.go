package cryptodomain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Clip struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StreamerID    primitive.ObjectID `json:"streamerId" bson:"streamerId"`
	UserID        primitive.ObjectID `json:"userId" bson:"userId"`
	ClipName      string             `json:"clipName" bson:"clipName"`
	URL           string             `json:"url" bson:"url"`
	Likes         []string           `json:"likes" bson:"likes"`
	Duration      int                `json:"duration" bson:"duration"`
	Views         int                `json:"views" bson:"views"`
	Cover         string             `json:"cover" bson:"cover"`
	TotalLikes    int                `json:"totalLikes" bson:"totalLikes"`
	TotalRetweets int                `json:"totalRetweets" bson:"totalRetweets"`
	TotalComments int                `json:"totalComments" bson:"totalComments"`
	Timestamps    struct {
		CreatedAt int64 `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
		UpdatedAt int64 `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	} `json:"timestamps,omitempty" bson:"timestamps,omitempty"`
}
