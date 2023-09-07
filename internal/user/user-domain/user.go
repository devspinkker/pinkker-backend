package userdomain

import (
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	SubscriptionNameUser string    `bson:"SubscriptionNameUser"`
	SubscriptionStart    time.Time `bson:"SubscriptionStart"`
	SubscriptionEnd      time.Time `bson:"SubscriptionEnd"`
	MonthsSubscribed     int       `bson:"MonthsSubscribed"`
}
type Subscriber struct {
	SubscriberNameUser string    `bson:"SubscriberNameUser"`
	SubscriptionStart  time.Time `bson:"SubscriptionStart"`
	SubscriptionEnd    time.Time `bson:"SubscriptionEnd"`
	MonthsSubscribed   int       `bson:"MonthsSubscribed"`
}
type User struct {
	ID                primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Avatar            string                 `json:"Avatar" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png" bson:"Avatar,omitempty"`
	FullName          string                 `json:"FullName" bson:"FullName"`
	NameUser          string                 `json:"NameUser" bson:"NameUser"`
	PasswordHash      string                 `json:"passwordHash" bson:"PasswordHash"`
	Pais              string                 `json:"Pais" bson:"Pais"`
	Subscriptions     []Subscription         `bson:"Subscriptions"`
	Subscribers       []Subscriber           `bson:"Subscribers"`
	Ciudad            string                 `json:"Ciudad" bson:"Ciudad"`
	Email             string                 `json:"Email" bson:"Email"`
	EmailConfirmation bool                   `json:"EmailConfirmation" bson:"EmailConfirmation,default:false"`
	Role              int                    `json:"role" bson:"Role,default:0"`
	KeyTransmission   string                 `json:"keyTransmission,omitempty" bson:"KeyTransmission,omitempty"`
	Biography         string                 `json:"biography" default:"Bienvenido a pinkker! actualiza tu biografía en ajustes de cuenta." bson:"Biography"`
	Look              string                 `json:"look" default:"h_std_cc_3032_7_0-undefined-undefined.ch-215-62-78.hd-180-10.lg-270-110" bson:"Look"`
	LookImage         string                 `json:"lookImage" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png" bson:"LookImage"`
	HeadImage         string                 `json:"headImage" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png" bson:"headImage"`
	Color             string                 `json:"color" bson:"Color"`
	BirthDate         time.Time              `json:"birthDate" bson:"BirthDate"`
	Pixeles           float64                `json:"Pixeles,default:0.0" bson:"Pixeles,default:0.0"`
	CustomAvatar      bool                   `json:"customAvatar,omitempty" bson:"CustomAvatar,omitempty"`
	CountryInfo       map[string]interface{} `json:"countryInfo,omitempty" bson:"CountryInfo,omitempty"`
	PinkkerPrime      struct {
		Active bool      `json:"active,omitempty" bson:"Active,omitempty"`
		Date   time.Time `json:"date,omitempty" bson:"Date,omitempty"`
	} `json:"pinkkerPrime,omitempty" bson:"PinkkerPrime,omitempty"`
	Suscribers    []string `json:"suscribers,omitempty" bson:"Suscribers,omitempty"`
	SocialNetwork struct {
		Facebook  string `json:"facebook,omitempty" bson:"facebook,omitempty"`
		Twitter   string `json:"twitter,omitempty" bson:"twitter,omitempty"`
		Instagram string `json:"instagram,omitempty" bson:"instagram,omitempty"`
		Youtube   string `json:"youtube,omitempty" bson:"youtube,omitempty"`
		Tiktok    string `json:"tiktok,omitempty" bson:"tiktok,omitempty"`
	} `json:"socialnetwork,omitempty" bson:"socialnetwork,omitempty"`
	Cmt                      string               `json:"cmt,omitempty" bson:"Cmt,omitempty"`
	Verified                 bool                 `json:"verified,omitempty" bson:"Verified"`
	Website                  string               `json:"website,omitempty" bson:"Website,omitempty"`
	Phone                    string               `json:"phone,omitempty" bson:"Phone,omitempty"`
	Sex                      string               `json:"sex,omitempty" bson:"Sex,omitempty"`
	Situation                string               `json:"situation,omitempty" bson:"Situation,omitempty"`
	UserFriendsNotifications int                  `json:"userFriendsNotifications,omitempty" bson:"UserFriendsNotifications,omitempty"`
	Following                []primitive.ObjectID `json:"Following" bson:"Following"`
	Followers                []primitive.ObjectID `json:"Followers" bson:"Followers"`
	Timestamp                time.Time            `json:"Timestamp" bson:"Timestamp"`
	Likes                    []primitive.ObjectID `json:"Likes" bson:"Likes"`
	Wallet                   string               `json:"Wallet" bson:"Wallet"`
}

type UserModelValidator struct {
	FullName  string `json:"fullName" validate:"required,min=8,max=70"`
	NameUser  string `json:"NameUser" validate:"required,min=5,max=20"`
	Password  string `json:"password" validate:"required,min=8"`
	Pais      string `json:"Pais" validate:"required"`
	Ciudad    string `json:"Ciudad" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Instagram string `json:"instagram" default:""`
	Twitter   string `json:"twitter" default:""`
	Youtube   string `json:"youtube" default:""`
	Wallet    string `json:"Wallet" default:""`
}

func (u *UserModelValidator) ValidateUser() error {
	validate := validator.New()
	return validate.Struct(u)
}

type LoginValidatorStruct struct {
	NameUser string `json:"NameUser" validate:"required,max=70"`
	Password string `json:"password" validate:"required,min=8"`
}

func (L *LoginValidatorStruct) LoginValidator() error {
	validate := validator.New()
	return validate.Struct(L)
}

type GetUser struct {
	ID                primitive.ObjectID     `json:"id"`
	Avatar            string                 `json:"Avatar"`
	FullName          string                 `json:"FullName" `
	NameUser          string                 `json:"NameUser" `
	Pais              string                 `json:"Pais"`
	Ciudad            string                 `json:"Ciudad"`
	Email             string                 `json:"Email" `
	EmailConfirmation bool                   `json:"EmailConfirmation" `
	Role              int                    `json:"role" `
	KeyTransmission   string                 `json:"keyTransmission,omitempty" `
	Biography         string                 `json:"biography" default:"Bienvenido a pinkker! actualiza tu biografía en ajustes de cuenta."`
	Look              string                 `json:"look" default:"h_std_cc_3032_7_0-undefined-undefined.ch-215-62-78.hd-180-10.lg-270-110"`
	LookImage         string                 `json:"lookImage" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png"`
	HeadImage         string                 `json:"headImage" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png"`
	Color             string                 `json:"color" `
	BirthDate         time.Time              `json:"birthDate" `
	Pixeles           float64                `json:"Pixeles,default:0.0" `
	CustomAvatar      bool                   `json:"customAvatar,omitempty"`
	CountryInfo       map[string]interface{} `json:"countryInfo,omitempty"`
	PinkkerPrime      struct {
		Active bool      `json:"active,omitempty`
		Date   time.Time `json:"date,omitempty"`
	} `json:"pinkkerPrime,omitempty"`
	Suscribers       []string `json:"suscribers,omitempty" `
	Subscriptions    []string `json:"subscriptions,omitempty" `
	SuscriptionPrice int      `json:"suscriptionPrice,default:300"`
	SocialNetwork    struct {
		Facebook  string `json:"facebook,omitempty"`
		Twitter   string `json:"twitter,omitempty" `
		Instagram string `json:"instagram,omitempty" `
		Youtube   string `json:"youtube,omitempty" `
		Tiktok    string `json:"tiktok,omitempty"`
	} `json:"socialnetwork,omitempty"`
	Cmt                      string               `json:"cmt,omitempty" `
	Verified                 bool                 `json:"verified,omitempty" `
	Website                  string               `json:"website,omitempty" `
	Phone                    string               `json:"phone,omitempty" `
	Sex                      string               `json:"sex,omitempty" `
	Situation                string               `json:"situation,omitempty" `
	UserFriendsNotifications int                  `json:"userFriendsNotifications,omitempty"`
	Following                []primitive.ObjectID `json:"Following"`
	Followers                []primitive.ObjectID `json:"Followers"`
	Timestamps               struct {
		CreatedAt int64 `json:"createdAt,omitempty" `
		UpdatedAt int64 `json:"updatedAt,omitempty" `
	} `json:"timestamps,omitempty" `
	Likes []primitive.ObjectID `json:"Likes"`
}
