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
	IdUser                 primitive.ObjectID   `json:"IdUser" bson:"IdUser"`
	Destination            string               `json:"Destination" bson:"Destination"` // para Streams o para Muro o que ClipAds
	Category               string               `json:"Category" bson:"Category"`       // preferencia del creador en hacia donde se dirija el ad de stream
	Impressions            int                  `json:"Impressions" bson:"Impressions"`
	ImpressionsMax         int                  `json:"ImpressionsMax" bson:"ImpressionsMax"`
	UrlVideo               string               `json:"UrlVideo" bson:"UrlVideo"` // para stream
	ReferenceLink          string               `json:"ReferenceLink" bson:"ReferenceLink"`
	PayPerPrint            float64              `json:"PayPerPrint" bson:"PayPerPrint"`
	Clicks                 int                  `json:"Clicks" bson:"Clicks"`
	ClicksMax              int                  `json:"ClicksMax" bson:"ClicksMax"`
	DocumentToBeAnnounced  primitive.ObjectID   `json:"DocumentToBeAnnounced" bson:"DocumentToBeAnnounced"`   // documento de muro al que quiere promocionar
	IdOfTheUsersWhoClicked []primitive.ObjectID `json:"IdOfTheUsersWhoClicked" bson:"IdOfTheUsersWhoClicked"` // en caso de ser clicks para mantener un conteo de que usuario dio like
	ClicksPerDay           []ClicksPerDay       `json:"ClicksPerDay" bson:"ClicksPerDay"`
	ImpressionsPerDay      []ImpressionsPerDay  `json:"ImpressionsPerDay" bson:"ImpressionsPerDay"`
	Timestamp              time.Time            `json:"Timestamp" bson:"Timestamp"`
	State                  string               `json:"State"  bson:"State"` // aceptado es que esta siendo funcional
	ClipId                 primitive.ObjectID   `json:"ClipId"  bson:"ClipId"`
	// el anuncio puede ser de tipo muro y ser mostrado solo en comunidades, las comunides a la que se muestra
	// tendra un tiempo en que expira el ad
	CommunityId          primitive.ObjectID `json:"CommunityId"  bson:"CommunityId"`
	EndAdCommunity       time.Time          `json:"EndAdCommunity"  bson:"EndAdCommunity"`
	PricePTotalCommunity float64            `json:"PricePTotalCommunity"  bson:"PricePTotalCommunity"`
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
	Category              string             `json:"Category" bson:"Category"`
	ImpressionsMax        int                `json:"ImpressionsMax" bson:"ImpressionsMax"`
	UrlVideo              string             `json:"UrlVideo" bson:"UrlVideo"`
	ReferenceLink         string             `json:"ReferenceLink" bson:"ReferenceLink"`
	Code                  string             `json:"Code"`
	ClicksMax             int                `json:"ClicksMax"`
	DocumentToBeAnnounced primitive.ObjectID `json:"DocumentToBeAnnounced"`
}
type UpdateAdvertisementCommunityId struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name                  string             `json:"Name" bson:"Name"`
	NameUser              string             `json:"NameUser" bson:"NameUser"`
	Destination           string             `json:"Destination" bson:"Destination"`
	Category              string             `json:"Category" bson:"Category"`
	ImpressionsMax        int                `json:"ImpressionsMax" bson:"ImpressionsMax"`
	UrlVideo              string             `json:"UrlVideo" bson:"UrlVideo"`
	ReferenceLink         string             `json:"ReferenceLink" bson:"ReferenceLink"`
	Code                  string             `json:"Code"`
	ClicksMax             int                `json:"ClicksMax"`
	DocumentToBeAnnounced primitive.ObjectID `json:"DocumentToBeAnnounced"`
	CommunityId           primitive.ObjectID `json:"CommunityId"`
	DaysTheAd             time.Time          `json:"Days"`
}
type ClipAdsCreate struct {
	Name                  string             `json:"Name" bson:"Name"`
	NameUser              string             `json:"NameUser" bson:"NameUser"`
	Destination           string             `json:"Destination" bson:"Destination"` // ClipAds
	Category              string             `json:"Category" bson:"Category"`
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
func (u *UpdateAdvertisementCommunityId) Validate() error {
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
