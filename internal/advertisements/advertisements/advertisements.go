package advertisements

import (
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// si una marca nos pide que imprimamos 20000 ads y nos paga 3000 (quedarian Budget=900
// el 70% es para pinkker) el calculo del valor PayPerPrint seria
// PayPerPrint= Budget/Impressions
type Advertisements struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name                  string             `json:"Name" bson:"Name"`
	Destination           string             `json:"Destination" bson:"Destination"` // para stream o para muro o que
	Categorie             string             `json:"Categorie" bson:"Categorie"`
	Impressions           int                `json:"Impressions" bson:"Impressions"`
	ImpressionsMax        int                `json:"ImpressionsMax" bson:"ImpressionsMax"`
	UrlVideo              string             `json:"UrlVideo" bson:"UrlVideo"`
	ReferenceLink         string             `json:"ReferenceLink" bson:"ReferenceLink"`
	PayPerPrint           float64            `json:"PayPerPrint" bson:"PayPerPrint"`
	Clicks                int                `json:"Clicks" bson:"Clicks"`
	ClicksMax             int                `json:"ClicksMax" bson:"ClicksMax"`
	DocumentToBeAnnounced primitive.ObjectID `json:"DocumentToBeAnnounced" bson:"DocumentToBeAnnounced"`
}

type AdvertisementGet struct {
	Code string `json:"Code"`
}
type UpdateAdvertisement struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name                  string             `json:"Name" bson:"Name"`
	Destination           string             `json:"Destination" bson:"Destination"` // para stream o para muro o que
	Categorie             string             `json:"Categorie" bson:"Categorie"`
	ImpressionsMax        int                `json:"ImpressionsMax" bson:"ImpressionsMax"`
	UrlVideo              string             `json:"UrlVideo" bson:"UrlVideo"`
	ReferenceLink         string             `json:"ReferenceLink" bson:"ReferenceLink"`
	Code                  string             `json:"Code"`
	ClicksMax             int                `json:"ClicksMax"`
	DocumentToBeAnnounced primitive.ObjectID `json:"DocumentToBeAnnounced"`
}

func (u *UpdateAdvertisement) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type DeleteAdvertisement struct {
	ID   primitive.ObjectID `json:"ID" bson:"ID"`
	Code string             `json:"Code"`
}
