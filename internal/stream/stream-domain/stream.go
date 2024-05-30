package streamdomain

import (
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Stream struct {
	ID                     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	StreamerID             primitive.ObjectID `json:"streamerId" bson:"StreamerID"`
	Streamer               string             `json:"streamer" bson:"Streamer"`
	StreamerAvatar         string             `json:"streamer_avatar" bson:"StreamerAvatar,omitempty"`
	ViewerCount            int                `json:"ViewerCount"  bson:"ViewerCount,default:0"`
	Online                 bool               `json:"online" bson:"Online,default:false"`
	StreamTitle            string             `json:"stream_title" bson:"StreamTitle"`
	StreamCategory         string             `json:"stream_category" bson:"StreamCategory"`
	ImageCategorie         string             `json:"ImageCategorie" bson:"ImageCategorie"`
	StreamNotification     string             `json:"stream_notification" bson:"StreamNotification"`
	StreamTag              []string           `json:"stream_tag"  bson:"StreamTag,default:['Español']"`
	StreamLikes            []string           `json:"stream_likes" bson:"StreamLikes"`
	StreamIdiom            string             `json:"stream_idiom" default:"Español" bson:"StreamIdiom,default:'Español'"`
	StreamThumbnail        string             `json:"stream_thumbnail" bson:"StreamThumbnail"`
	StartDate              time.Time          `json:"start_date" bson:"StartDate"`
	Timestamp              time.Time          `json:"Timestamp" bson:"Timestamp"`
	EmotesChat             map[string]string  `json:"EmotesChat" bson:"EmotesChat"`
	ModChat                string             `json:"ModChat" bson:"ModChat"`
	ModSlowMode            int                `json:"ModSlowMode" bson:"ModSlowMode"`
	Banned                 bool               `json:"Banned" bson:"Banned"`
	TotalTimeOnlineSeconds int64              `json:"TotalTimeOnlineSeconds" bson:"TotalTimeOnlineSeconds"`
}

type UpdateStreamInfo struct {
	Date         int64    `json:"date"`
	Title        string   `json:"title" validate:"min=5,max=30"`
	Notification string   `json:"notification" validate:"min=5,max=30"`
	Category     string   `json:"category" validate:"min=3"`
	Tag          []string `json:"tag" `
	Idiom        string   `json:"idiom"`
}

func (u *UpdateStreamInfo) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type UpdateModChat struct {
	Mod string `json:"title" validate:"max=30"`
}

func (u *UpdateModChat) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type UpdateModChatSlowMode struct {
	ModSlowMode int `json:"ModSlowMode" validate:"min=0,max=30"`
}
type CommercialInStream struct {
	CommercialInStream int `json:"CommercialInStream"`
}

func (u *CommercialInStream) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}
func (u *UpdateModChatSlowMode) Validate() error {
	validate := validator.New()
	if err := validate.Struct(u); err != nil {
		return err
	}

	// if err := u.customModSlowModeValidator(); err != nil {
	// 	return err
	// }

	return nil
}

func (u *UpdateModChatSlowMode) customModSlowModeValidator() error {
	// allowedValues := map[int]bool{2: true, 5: true, 15: true, 30: true, 45: true, 60: true, 120: true}
	// if !allowedValues[u.ModSlowMode] {
	// 	return fmt.Errorf("ModSlowMode debe ser uno de los valores permitidos: 2, 5, 15, 30, 45, 60, 120")
	// }
	return nil
}

type Update_start_date struct {
	Date int    `json:"date"`
	Key  string `json:"keyTransmission"`
}
type CategoriesUpdate struct {
	Name       string `json:"Name" validate:"required"`
	Img        string `json:"img,omitempty"`
	Spectators int    `json:"spectators,omitempty"`
	TopColor   string `json:"TopColor,omitempty"`
	CodeAdmin  string `json:"CodeAdmin" validate:"required"`
	Delete     bool   `json:"Delete"`
}

func (u *CategoriesUpdate) Validate() error {
	validate := validator.New()
	if err := validate.Struct(u); err != nil {
		return err
	}

	return nil
}

type Categoria struct {
	Name       string   `json:"nombre"`
	Img        string   `json:"img,omitempty"`
	Spectators int      `json:"spectators,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	TopColor   string   `json:"TopColor,omitempty"`
}
