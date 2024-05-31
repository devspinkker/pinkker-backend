package withdrawalsdomain

import (
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WithdrawalRequests struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Destination      string             `json:"Destination" bson:"Destination"`
	AcceptedBy       primitive.ObjectID `json:"AcceptedBy" bson:"AcceptedBy"`
	RequesteNameUser string             `json:"RequesteNameUser" bson:"RequesteNameUser"`
	RequestedBy      primitive.ObjectID `json:"RequestedBy" bson:"RequestedBy"`
	Amount           float64            `json:"Amount" bson:"Amount"`
	TimeStamp        time.Time          `json:"TimeStamp" bson:"TimeStamp"`
	Notified         bool               `json:"Notified"  bson:"Notified,default:false"`
	State            string             `json:"State"  bson:"State"`
	TextReturn       string             `json:"TextReturn"  bson:"TextReturn"`
}
type WithdrawalRequestReq struct {
	Amount string `json:"amount"`
	Cbu    string `json:"cbu"`
}

func (u *WithdrawalRequestReq) ValidateReq() error {

	validate := validator.New()
	return validate.Struct(u)
}

type WithdrawalRequestGet struct {
	Code string `json:"Code"`
	Page int    `json:"PagePage"`
}

func (u *WithdrawalRequestGet) ValidateReqCreateDonation() error {

	validate := validator.New()
	return validate.Struct(u)
}

type AcceptWithdrawal struct {
	Code                 string             `json:"Code"`
	WithdrawalRequestsId primitive.ObjectID `json:"WithdrawalRequestsId"`
}

func (u *AcceptWithdrawal) Validate() error {

	validate := validator.New()
	return validate.Struct(u)
}
