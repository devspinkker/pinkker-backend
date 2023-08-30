package donationdomain

import (
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Donation struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FromUser  primitive.ObjectID `json:"FromUser" bson:"FromUser"`
	ToUser    primitive.ObjectID `json:"ToUser" bson:"ToUser"`
	Pixeles   float64            `json:"Pixeles" bson:"Pixeles"`
	Text      string             `json:"Text" bson:"Text"`
	TimeStamp time.Time          `json:"TimeStamp" bson:"TimeStamp"`
	Notified  bool               `json:"Notified"  bson:"Notified,default:false"`
}
type ReqCreateDonation struct {
	ToUser  primitive.ObjectID `json:"ToUser" validate:"required"`
	Pixeles float64            `json:"Pixeles" validate:"gt=1,required"`
	Text    string             `json:"Text" validate:"required,min=2,max=70"`
}

type ResDonation struct {
	ID           primitive.ObjectID `json:"id"`
	FromUser     primitive.ObjectID `json:"FromUser"`
	ToUser       primitive.ObjectID `json:"ToUser"`
	Pixeles      float64            `json:"Pixeles"`
	Text         string             `json:"Text"`
	TimeStamp    time.Time          `json:"TimeStamp"`
	Notified     bool               `json:"Notified"`
	FromUserInfo struct {
		Avatar   string `json:"Avatar"`
		NameUser string `json:"NameUser"`
	} `json:"FromUserInfo"`
}

func (u *ReqCreateDonation) ValidateReqCreateDonation() error {

	validate := validator.New()
	return validate.Struct(u)
}
