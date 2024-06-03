package advertisements

import (
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Advertisements struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name          string             `json:"Name" bson:"Name"`
	Destination   string             `json:"Destination" bson:"Destination"` // para stream o para muro o que
	Categorie     string             `json:"Categorie" bson:"Categorie"`
	Impressions   int                `json:"Impressions" bson:"Impressions"`
	UrlVideo      string             `json:"UrlVideo" bson:"UrlVideo"`
	ReferenceLink string             `json:"ReferenceLink" bson:"ReferenceLink"`
	PayPerPrint   float64            `json:"PayPerPrint" bson:"PayPerPrint"`
}
type AdvertisementGet struct {
	Code string `json:"Code"`
}
type UpdateAdvertisement struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name          string             `json:"Name" bson:"Name"`
	Destination   string             `json:"Destination" bson:"Destination"` // para stream o para muro o que
	Categorie     string             `json:"Categorie" bson:"Categorie"`
	Impressions   int                `json:"Impressions" bson:"Impressions"`
	UrlVideo      string             `json:"UrlVideo" bson:"UrlVideo"`
	ReferenceLink string             `json:"ReferenceLink" bson:"ReferenceLink"`
	PayPerPrint   float64            `json:"PayPerPrint" bson:"PayPerPrint"`
	Code          string             `json:"Code"`
}

func (u *UpdateAdvertisement) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type DeleteAdvertisement struct {
	ID   primitive.ObjectID `json:"ID" bson:"ID"`
	Code string             `json:"Code"`
}
