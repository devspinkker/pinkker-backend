package advertisements

import (
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// si una marca nos pide que imprimamos 20000 ads y nos paga 3000 (quedarian Budget=900
// el 70% es para pinkker) el calculo del valor PayPerPrint seria
// PayPerPrint= Budget/Impressions
type Advertisements struct {
	ID                     primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name                   string               `json:"Name" bson:"Name"`
	NameUser               string               `json:"NameUser" bson:"NameUser"`
	Destination            string               `json:"Destination" bson:"Destination"` // para Streams o para Muro o que ClipAds
	Categorie              string               `json:"Categorie" bson:"Categorie"`
	Impressions            int                  `json:"Impressions" bson:"Impressions"`
	ImpressionsMax         int                  `json:"ImpressionsMax" bson:"ImpressionsMax"`
	UrlVideo               string               `json:"UrlVideo" bson:"UrlVideo"`
	ReferenceLink          string               `json:"ReferenceLink" bson:"ReferenceLink"`
	PayPerPrint            float64              `json:"PayPerPrint" bson:"PayPerPrint"`
	Clicks                 int                  `json:"Clicks" bson:"Clicks"`
	ClicksMax              int                  `json:"ClicksMax" bson:"ClicksMax"`
	DocumentToBeAnnounced  primitive.ObjectID   `json:"DocumentToBeAnnounced" bson:"DocumentToBeAnnounced"`
	IdOfTheUsersWhoClicked []primitive.ObjectID `json:"IdOfTheUsersWhoClicked" bson:"IdOfTheUsersWhoClicked"`
	ClicksPerDay           []ClicksPerDay       `json:"ClicksPerDay" bson:"ClicksPerDay"`
	ImpressionsPerDay      []ImpressionsPerDay  `json:"ImpressionsPerDay" bson:"ImpressionsPerDay"`
	Timestamp              time.Time            `json:"Timestamp" bson:"Timestamp"`
	State                  string               `json:"State"  bson:"State"`
	ClipId                 primitive.ObjectID   `json:"ClipId"  bson:"ClipId"`
}

type ClicksPerDay struct {
	Date   string `json:"Date" bson:"Date"`
	Clicks int    `json:"Clicks" bson:"Clicks"`
}

type ImpressionsPerDay struct {
	Date        string `json:"Date" bson:"Date"`
	Impressions int    `json:"Impressions" bson:"Impressions"`
}

type AdvertisementGet struct {
	Code string `json:"Code"`
}
type AcceptPendingAds struct {
	Code     string `json:"Code"`
	NameUser string `json:"NameUser"`
}

type GetAdsUserCode struct {
	Code     string `json:"Code"`
	NameUser string `json:"NameUser" bson:"NameUser"`
}

type UpdateAdvertisement struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name                  string             `json:"Name" bson:"Name"`
	NameUser              string             `json:"NameUser" bson:"NameUser"`
	Destination           string             `json:"Destination" bson:"Destination"`
	Categorie             string             `json:"Categorie" bson:"Categorie"`
	ImpressionsMax        int                `json:"ImpressionsMax" bson:"ImpressionsMax"`
	UrlVideo              string             `json:"UrlVideo" bson:"UrlVideo"`
	ReferenceLink         string             `json:"ReferenceLink" bson:"ReferenceLink"`
	Code                  string             `json:"Code"`
	ClicksMax             int                `json:"ClicksMax"`
	DocumentToBeAnnounced primitive.ObjectID `json:"DocumentToBeAnnounced"`
}

type ClipAdsCreate struct {
	Name                  string             `json:"Name" bson:"Name"`
	NameUser              string             `json:"NameUser" bson:"NameUser"`
	Destination           string             `json:"Destination" bson:"Destination"` // ClipAds
	Categorie             string             `json:"Categorie" bson:"Categorie"`
	ImpressionsMax        int                `json:"ImpressionsMax" bson:"ImpressionsMax"`
	UrlVideo              string             `json:"UrlVideo" bson:"UrlVideo"`
	ReferenceLink         string             `json:"ReferenceLink" bson:"ReferenceLink"`
	Code                  string             `json:"Code"`
	ClicksMax             int                `json:"ClicksMax"`
	DocumentToBeAnnounced primitive.ObjectID `json:"DocumentToBeAnnounced"`
	ClipTitle             string             `json:"clipTitle" validate:"required,min=2,max=100"`
	TotalKey              string             `json:"totalKey" validate:"required"`
}

func (u *UpdateAdvertisement) Validate() error {
	validate := validator.New()

	validate.RegisterValidation("nonnegative", func(fl validator.FieldLevel) bool {
		value := fl.Field().Int()
		return value >= 0
	})

	// Validate struct
	err := validate.Struct(u)
	if err != nil {
		return err
	}

	if u.ImpressionsMax < 0 {
		return fmt.Errorf("ImpressionsMax must be non-negative")
	}
	if u.ClicksMax < 0 {
		return fmt.Errorf("ClicksMax must be non-negative")
	}

	return nil
}

type DeleteAdvertisement struct {
	ID   primitive.ObjectID `json:"ID" bson:"ID"`
	Code string             `json:"Code"`
}
