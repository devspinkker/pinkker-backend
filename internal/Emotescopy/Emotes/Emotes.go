package EmotesDomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmotePair struct {
	Name string `json:"name" bson:"name"`
	URL  string `json:"url" bson:"url"`
}

// Emote representa un emote individual
type Emote struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Emotes    []EmotePair        `json:"emotes" bson:"emotes"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	Type      string             `json:"type" bson:"type"`                         // seccion de la que ocupa, global o es de pinkker o sub etc
	UserID    primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"` // Solo para emotes de usuario
}

type EmoteUpdate struct {
	ID     primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	Emotes EmotePair           `json:"emotes" bson:"emotes"`
	Type   string              `json:"type" bson:"type"`
	UserID *primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"` // Solo para emotes de usuario
	Code   string              `json:"Code" bson:"Code"`
	Name   string              `json:"name" bson:"name"`
}

type EmoteUpdateOrCreate struct {
	Name      string `json:"name" bson:"name"`
	TypeEmote string `json:"typeEmote" bson:"typeEmote"`
}
type GetEmoteIdUserandType struct {
	IdUser    primitive.ObjectID `json:"IdUser" bson:"IdUserIdUser"`
	TypeEmote string             `json:"typeEmote" bson:"typeEmote"`
}

// func (u *Emotes) Validate() error {
// 	validate := validator.New()
// 	return validate.Struct(u)
// }
