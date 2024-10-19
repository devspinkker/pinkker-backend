package userdomain

import (
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	"regexp"
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Avatar            string                 `json:"Avatar" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png" bson:"Avatar"`
	FullName          string                 `json:"FullName" bson:"FullName"`
	NameUser          string                 `json:"NameUser" bson:"NameUser"`
	PasswordHash      string                 `json:"passwordHash" bson:"PasswordHash"`
	Pais              string                 `json:"Pais" bson:"Pais"`
	Subscriptions     []primitive.ObjectID   `bson:"Subscriptions"`
	Subscribers       []primitive.ObjectID   `bson:"Subscribers"`
	Clips             []primitive.ObjectID   `bson:"Clips,omitempty"`
	ClipsLikes        []primitive.ObjectID   `bson:"ClipsLikes,omitempty"`
	Ciudad            string                 `json:"Ciudad" bson:"Ciudad"`
	Email             string                 `json:"Email" bson:"Email"`
	EmailConfirmation bool                   `json:"EmailConfirmation" bson:"EmailConfirmation,default:false"`
	Role              int                    `json:"role" bson:"Role,default:0"`
	KeyTransmission   string                 `json:"keyTransmission,omitempty" bson:"KeyTransmission"`
	Biography         string                 `json:"biography" default:"Bienvenido a pinkker! actualiza tu biografía en ajustes de cuenta." bson:"Biography"`
	Look              string                 `json:"look" default:"h_std_cc_3032_7_0-undefined-undefined.ch-215-62-78.hd-180-10.lg-270-110" bson:"Look"`
	LookImage         string                 `json:"lookImage" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png" bson:"LookImage"`
	Banner            string                 `json:"Banner"  bson:"Banner"`
	HeadImage         string                 `json:"headImage" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png" bson:"headImage"`
	Color             string                 `json:"color" bson:"Color"`
	BirthDate         time.Time              `json:"birthDate" bson:"BirthDate"`
	Pixeles           float64                `json:"Pixeles,default:0.0" bson:"Pixeles,default:0.0"`
	CustomAvatar      bool                   `json:"customAvatar,omitempty" bson:"CustomAvatar"`
	CountryInfo       map[string]interface{} `json:"countryInfo,omitempty" bson:"CountryInfo"`
	Partner           struct {
		Active bool      `json:"active,omitempty" bson:"Active,omitempty"`
		Date   time.Time `json:"date,omitempty" bson:"Date,omitempty"`
	} `json:"Partner,omitempty" bson:"Partner"`
	EditProfiile struct {
		NameUser  time.Time `json:"NameUser,omitempty" bson:"NameUser,omitempty"`
		Biography time.Time `json:"Biography,omitempty" bson:"Biography,omitempty"`
	} `json:"EditProfiile,omitempty" bson:"EditProfiile"`
	SocialNetwork struct {
		Facebook  string `json:"facebook,omitempty" bson:"facebook"`
		Twitter   string `json:"twitter,omitempty" bson:"twitter"`
		Instagram string `json:"instagram,omitempty" bson:"instagram"`
		Youtube   string `json:"youtube,omitempty" bson:"youtube"`
		Tiktok    string `json:"tiktok,omitempty" bson:"tiktok"`
	} `json:"socialnetwork,omitempty" bson:"socialnetwork"`
	Cmt                      string                            `json:"cmt,omitempty" bson:"Cmt"`
	Verified                 bool                              `json:"verified,omitempty" bson:"Verified"`
	Website                  string                            `json:"website,omitempty" bson:"Website"`
	Phone                    string                            `json:"phone,omitempty" bson:"Phone"`
	Sex                      string                            `json:"sex,omitempty" bson:"Sex"`
	Situation                string                            `json:"situation,omitempty" bson:"Situation"`
	UserFriendsNotifications int                               `json:"userFriendsNotifications,omitempty" bson:"UserFriendsNotifications"`
	Following                map[primitive.ObjectID]FollowInfo `json:"Following" bson:"Following"`
	Followers                map[primitive.ObjectID]FollowInfo `json:"Followers" bson:"Followers"`
	Timestamp                time.Time                         `json:"Timestamp" bson:"Timestamp"`
	Likes                    []primitive.ObjectID              `json:"Likes" bson:"Likes"`
	Wallet                   string                            `json:"Wallet" bson:"Wallet"`
	Online                   bool                              `json:"Online,omitempty" bson:"Online,omitempty" default:"false"`
	ClipsComment             []primitive.ObjectID              `json:"ClipsComment" bson:"ClipsComment,omitempty"`
	CategoryPreferences      map[string]float64                `json:"categoryPreferences" bson:"categoryPreferences"`
	PanelAdminPinkker        struct {
		Level int       `json:"Level,omitempty" bson:"Level" default:"0"`
		Asset bool      `json:"Asset,omitempty" bson:"Asset,omitempty" default:"false"`
		Code  string    `json:"Code,omitempty" bson:"Code"`
		Date  time.Time `json:"date,omitempty" bson:"Date,omitempty"`
	} `json:"PanelAdminPinkker,omitempty" bson:"PanelAdminPinkker"`
	Banned           bool                 `json:"Banned" bson:"Banned"`
	TOTPSecret       string               `json:"TOTPSecret" bson:"TOTPSecret"`
	LastConnection   time.Time            `json:"LastConnection" bson:"LastConnection"`
	PinkkerPrime     PinkkerPrime         `json:"PinkkerPrime" bson:"PinkkerPrime"`
	InCommunities    []primitive.ObjectID `json:"InCommunities" bson:"InCommunities"`
	OwnerCommunities []primitive.ObjectID `json:"OwnerCommunities" bson:"OwnerCommunities"`
}
type PinkkerPrime struct {
	MonthsSubscribed  int       `bson:"MonthsSubscribed"`
	SubscriptionStart time.Time `bson:"SubscriptionStart"`
	SubscriptionEnd   time.Time `bson:"SubscriptionEnd"`
}
type FollowInfo struct {
	Since         time.Time `json:"since" bson:"since"`
	Notifications bool      `json:"notifications" bson:"notifications"`
	Email         string    `json:"Email" bson:"Email"`
}

type Subscription struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty"`
	SubscriptionNameUser string             `bson:"SubscriptionNameUser"`
	SourceUserID         primitive.ObjectID `bson:"sourceUserID"`
	DestinationUserID    primitive.ObjectID `bson:"destinationUserID"`
	SubscriptionStart    time.Time          `bson:"SubscriptionStart"`
	SubscriptionEnd      time.Time          `bson:"SubscriptionEnd"`
}

type FollowInfoRes struct {
	Since         time.Time `json:"since" bson:"since"`
	Notifications bool      `json:"notifications" bson:"notifications"`
	Email         string    `json:"Email" bson:"Email"`
	NameUser      string    `json:"NameUser" bson:"NameUser"`
	Avatar        string    `json:"Avatar" bson:"Avatar"`
}

type UserModelValidator struct {
	FullName      string    `json:"fullName" validate:"required,min=5,max=70"`
	NameUser      string    `json:"NameUser" validate:"nameuser" `
	Password      string    `json:"password" validate:"required,min=8"`
	Pais          string    `json:"Pais" `
	Ciudad        string    `json:"Ciudad"`
	Email         string    `json:"email" validate:"required,email"`
	Instagram     string    `json:"instagram" default:""`
	Twitter       string    `json:"twitter" default:""`
	Youtube       string    `json:"youtube" default:""`
	Wallet        string    `json:"Wallet" default:""`
	BirthDate     string    `json:"BirthDate" default:""`
	BirthDateTime time.Time `json:"-" bson:"BirthDate"`
}

func (u *UserModelValidator) ValidateUser() error {
	validate := validator.New()

	validate.RegisterValidation("nameuser", nameUserValidator)

	if u.BirthDate != "" {
		_, err := time.Parse("2006-01-02", u.BirthDate)
		if err != nil {
			return err
		}

		birthDate, _ := time.Parse("2006-01-02", u.BirthDate)
		u.BirthDateTime = birthDate
	}

	return validate.Struct(u)
}

type SocialNetwork struct {
	Facebook  string `json:"facebook,omitempty" bson:"facebook"`
	Twitter   string `json:"twitter,omitempty" bson:"twitter"`
	Instagram string `json:"instagram,omitempty" bson:"instagram"`
	Youtube   string `json:"youtube,omitempty" bson:"youtube"`
	Tiktok    string `json:"tiktok,omitempty" bson:"tiktok"`
}

type PanelAdminPinkkerInfoUserReq struct {
	Code     string             `json:"Code,omitempty" bson:"Code"`
	IdUser   primitive.ObjectID `json:"IdUser,omitempty" bson:"IdUser"`
	NameUser string             `json:"NameUser,omitempty" bson:"NameUser"`
}
type CreateAdmin struct {
	Code     string             `json:"Code,omitempty" bson:"Code"`
	IdUser   primitive.ObjectID `json:"IdUser,omitempty" bson:"IdUser"`
	NameUser string             `json:"NameUser,omitempty" bson:"NameUser"`
	Level    int                `json:"Level,omitempty" bson:"Level"`
	NewCode  string             `json:"NewCode,omitempty" bson:"NewCode" validate:"required,min=5,max=20" `
}
type ChangeNameUser struct {
	Code           string             `json:"Code,omitempty" bson:"Code" `
	IdUser         primitive.ObjectID `json:"IdUser,omitempty" bson:"IdUser"`
	NameUserNew    string             `json:"NameUserNew,omitempty" bson:"NameUserNew" validate:"NameUserNew"`
	NameUserRemove string             `json:"NameUserRemove,omitempty" bson:"NameUserRemove"`
}
type ReqGetUserByNameUser struct {
	User     *GetUser             `json:"user"`
	Stream   *streamdomain.Stream `json:"stream"`
	UserInfo *UserInfo            `json:"UserInfo"`
}

func nameUserValidator(fl validator.FieldLevel) bool {
	nameUser := fl.Field().String()

	// Verifica la longitud
	if len(nameUser) < 5 || len(nameUser) > 20 {
		return false
	}

	// Verifica que solo contenga caracteres alfanuméricos
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(nameUser) {
		return false
	}

	return true
}

// Valida la estructura
func (u *ChangeNameUser) ValidateUser() error {
	validate := validator.New()

	// Registro de validación personalizada
	validate.RegisterValidation("NameUserNew", nameUserValidator)

	// Validar estructura con la etiqueta personalizada
	return validate.Struct(u)
}

type ReqGoogle_callback_NameUserConfirm struct {
	NameUser string `json:"NameUser" validate:"required,min=5,max=20"`
	Pais     string `json:"Pais" validate:"required"`
	Ciudad   string `json:"Ciudad" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

func (u *ReqGoogle_callback_NameUserConfirm) ValidateUser() error {
	validate := validator.New()
	return validate.Struct(u)
}

type LoginValidatorStruct struct {
	NameUser string `json:"NameUser" validate:"required,max=70"`
	Password string `json:"password" validate:"required,min=8"`
}
type LoginTOTPSecret struct {
	NameUser string `json:"NameUser" validate:"required,max=70"`
	Password string `json:"password" validate:"required,min=8"`
	Totpcode string `json:"totp_code" `
}

func (L *LoginTOTPSecret) LoginTOTPSecret() error {
	validate := validator.New()
	return validate.Struct(L)
}

func (L *LoginValidatorStruct) LoginValidator() error {
	validate := validator.New()
	return validate.Struct(L)
}

type Req_Recover_lost_password struct {
	Mail string `json:"mail" validate:"required,max=70"`
}
type ReqRestorePassword struct {
	Code     string `json:"code"`
	Password string `json:"password" validate:"required,min=8"`
}
type DeleteGoogleAuthenticator struct {
	Code string `json:"code"`
}
type GetRecommended struct {
	ExcludeIDs []primitive.ObjectID `json:"ExcludeIDs" validate:"required"`
}
type GetUser struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Avatar   string             `json:"Avatar" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png" bson:"Avatar"`
	FullName string             `json:"FullName" bson:"FullName"`
	NameUser string             `json:"NameUser" bson:"NameUser"`
	Pais     string             `json:"Pais" bson:"Pais"`
	// Subscriptions   []primitive.ObjectID   `bson:"Subscriptions"`
	// Subscribers     []primitive.ObjectID   `bson:"Subscribers"`
	// Clips           []primitive.ObjectID   `bson:"Clips,omitempty"`
	// ClipsLikes      []primitive.ObjectID   `bson:"ClipsLikes,omitempty"`
	Ciudad          string                 `json:"Ciudad" bson:"Ciudad"`
	Email           string                 `json:"Email" bson:"Email"`
	Role            int                    `json:"role" bson:"Role,default:0"`
	KeyTransmission string                 `json:"keyTransmission,omitempty" bson:"KeyTransmission"`
	Biography       string                 `json:"biography" default:"Bienvenido a pinkker! actualiza tu biografía en ajustes de cuenta." bson:"Biography"`
	Look            string                 `json:"look" default:"h_std_cc_3032_7_0-undefined-undefined.ch-215-62-78.hd-180-10.lg-270-110" bson:"Look"`
	LookImage       string                 `json:"lookImage" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png" bson:"LookImage"`
	Banner          string                 `json:"Banner" default:"https://res.cloudinary.com/dcj8krp42/image/upload/v1712283573/categorias/logo_trazado_pndioh.png" bson:"Banner"`
	HeadImage       string                 `json:"headImage" default:"https://res.cloudinary.com/pinkker/image/upload/v1680478837/foto_default_obyind.png" bson:"headImage"`
	Online          bool                   `json:"Online" bson:"Online"`
	Color           string                 `json:"color" bson:"Color"`
	BirthDate       time.Time              `json:"birthDate" bson:"BirthDate"`
	CustomAvatar    bool                   `json:"customAvatar,omitempty" bson:"CustomAvatar"`
	CountryInfo     map[string]interface{} `json:"countryInfo,omitempty" bson:"CountryInfo"`
	Partner         struct {
		Active bool      `json:"active,omitempty" bson:"Active,omitempty"`
		Date   time.Time `json:"date,omitempty" bson:"Date,omitempty"`
	} `json:"Partner,omitempty" bson:"Partner"`
	// Suscribers    []string `json:"suscribers,omitempty" bson:"Suscribers"`
	SocialNetwork struct {
		Facebook  string `json:"facebook,omitempty" bson:"facebook"`
		Twitter   string `json:"twitter,omitempty" bson:"twitter"`
		Instagram string `json:"instagram,omitempty" bson:"instagram"`
		Youtube   string `json:"youtube,omitempty" bson:"youtube"`
		Tiktok    string `json:"tiktok,omitempty" bson:"tiktok"`
	} `json:"socialnetwork,omitempty" bson:"socialnetwork"`
	Verified                 bool   `json:"verified,omitempty" bson:"Verified"`
	Website                  string `json:"website,omitempty" bson:"Website"`
	Phone                    string `json:"phone,omitempty" bson:"Phone"`
	Sex                      string `json:"sex,omitempty" bson:"Sex"`
	Situation                string `json:"situation,omitempty" bson:"Situation"`
	UserFriendsNotifications int    `json:"userFriendsNotifications,omitempty" bson:"UserFriendsNotifications"`
	// Following                map[primitive.ObjectID]FollowInfo `json:"Following" bson:"Following"`
	// Followers                map[primitive.ObjectID]FollowInfo `json:"Followers" bson:"Followers"`
	FollowersCount int `json:"FollowersCount" bson:"FollowersCount"`
	FollowingCount int `json:"FollowingCount" bson:"FollowingCount"`

	SubscribersCount int       `json:"SubscribersCount" bson:"SubscribersCount"`
	Timestamp        time.Time `json:"Timestamp" bson:"Timestamp"`
	// Likes                    []primitive.ObjectID              `json:"Likes" bson:"Likes"`
	Wallet string `json:"Wallet" bson:"Wallet"`
	// ClipsComment             []primitive.ObjectID              `json:"ClipsComment" bson:"ClipsComment"`
	CategoryPreferences map[string]float64   `json:"categoryPreferences" bson:"categoryPreferences"`
	Banned              bool                 `json:"Banned" bson:"Banned"`
	IsFollowedByUser    bool                 `json:"isFollowedByUser" bson:"isFollowedByUser"`
	PinkkerPrime        PinkkerPrime         `json:"PinkkerPrime" bson:"PinkkerPrime"`
	InCommunities       []primitive.ObjectID `json:"InCommunities" bson:"InCommunities"`
}
type UserInfoOAuth2 struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}
type EditProfile struct {
	Pais      string `json:"Pais" bson:"Pais"`
	Ciudad    string `json:"Ciudad" bson:"Ciudad"`
	Biography string `json:"biography" validate:"max=600"`
	HeadImage string `json:"headImage"`

	BirthDate     string    `json:"birthDate"`
	BirthDateTime time.Time `json:"-" bson:"BirthDate"`
	Sex           string    `json:"sex,omitempty"`
	Situation     string    `json:"situation,omitempty"`
	ZodiacSign    string    `json:"ZodiacSign,omitempty"`
}

func (u *EditProfile) ValidateEditProfile() error {
	validate := validator.New()

	if u.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", u.BirthDate)
		if err != nil {
			return err
		}
		u.BirthDateTime = birthDate
	}

	return validate.Struct(u)
}

type Google_callback_Complete_Profile_And_Username struct {
	NameUser   string    `json:"nameUser" validate:"required,min=5,max=20"`
	Email      string    `json:"email" validate:"required,email"`
	Password   string    `json:"password" validate:"required,min=8"`
	Pais       string    `json:"pais" bson:"Pais"`
	Ciudad     string    `json:"ciudad" bson:"Ciudad"`
	Biography  string    `json:"biography" validate:"max=600"`
	HeadImage  string    `json:"headImage"`
	BirthDate  time.Time `json:"birthDate"`
	Sex        string    `json:"sex,omitempty"`
	Situation  string    `json:"situation,omitempty"`
	ZodiacSign string    `json:"zodiacSign,omitempty"`
}

func (u *Google_callback_Complete_Profile_And_Username) ValidateUser() error {
	validate := validator.New()
	if u.BirthDate.IsZero() || u.BirthDate.String() == "" {
		u.BirthDate = time.Now()
	}
	return validate.Struct(u)
}

type InfoUserInRoom struct {
	ID       primitive.ObjectID       `json:"id" bson:"_id,omitempty"`
	NameUser string                   `json:"nameuser" bson:"NameUser"`
	Color    string                   `json:"Color" bson:"Color"`
	Rooms    []map[string]interface{} `json:"rooms" bson:"Rooms"`
}

// este documento tiene todas los Rooms o chats en los que interactuo  Nameuser
type InfoUser struct {
	ID       primitive.ObjectID       `bson:"_id,omitempty"`
	Nameuser string                   `bson:"NameUser"`
	Color    string                   `bson:"Color"`
	Rooms    []map[string]interface{} `bson:"Rooms"`
}

// Rooms tiene esto como interface
//
//	newRoom := map[string]interface{}{
//		"Room":                 roomID,// primitive.ObjectID
//		"Vip":                  false,
//		"Color":                randomColor, //string
//		"Moderator":            false,
//		"Verified":             verified,// bool
//		"Subscription":         primitive.ObjectID{},
//		"Baneado":              false,
//		"TimeOut":              time.Now(),
//		"EmblemasChat":         userInfo.EmblemasChat,
//		"Following":            domain.FollowInfo{},
//		"StreamerChannelOwner": userInfo.StreamerChannelOwner,  //bool
//		"LastMessage":          time.Now(),
//	}
type SubscriptionInfo struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty"`
	SubscriptionNameUser string             `bson:"SubscriptionNameUser"`
	SourceUserID         primitive.ObjectID `bson:"sourceUserID"`
	DestinationUserID    primitive.ObjectID `bson:"destinationUserID"`
	SubscriptionStart    time.Time          `bson:"SubscriptionStart"`
	SubscriptionEnd      time.Time          `bson:"SubscriptionEnd"`
	MonthsSubscribed     int                `bson:"MonthsSubscribed"`
	Notified             bool               `bson:"Notified"`
	Text                 string             `bson:"Text"`
}
type UserInfo struct { // ROOMS
	Room                 primitive.ObjectID
	Color                string
	Vip                  bool
	Verified             bool
	Moderator            bool
	Subscription         primitive.ObjectID
	SubscriptionInfo     SubscriptionInfo
	Baneado              bool
	TimeOut              time.Time
	EmblemasChat         map[string]string
	Identidad            string
	Following            FollowInfo
	StreamerChannelOwner bool
	LastMessage          time.Time
}
