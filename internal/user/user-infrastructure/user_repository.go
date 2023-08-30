package userinfrastructure

import (
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	domain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewUserRepository(redisClient *redis.Client, mongoClient *mongo.Client) *UserRepository {
	return &UserRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}

func (u *UserRepository) SaveUserDB(User *domain.User) (primitive.ObjectID, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	insertResult, errInsertOne := GoMongoDBCollUsers.InsertOne(context.Background(), User)
	if errInsertOne != nil {
		return primitive.NilObjectID, errInsertOne
	}
	insertedID := insertResult.InsertedID.(primitive.ObjectID)
	return insertedID, nil
}
func (u *UserRepository) FindNameUser(NameUser string, Email string) (*domain.User, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var FindUserInDb primitive.D
	if Email == "" {
		FindUserInDb = bson.D{
			{Key: "NameUser", Value: NameUser},
		}
	} else {
		FindUserInDb = bson.D{
			{
				Key: "$or",
				Value: bson.A{
					bson.D{{Key: "NameUser", Value: NameUser}},
					bson.D{{Key: "Email", Value: Email}},
				},
			},
		}
	}
	var findUserInDbExist *domain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&findUserInDbExist)
	return findUserInDbExist, errCollUsers
}
func (u *UserRepository) FindUserById(id primitive.ObjectID) (*domain.User, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var FindUserInDb primitive.D
	FindUserInDb = bson.D{
		{Key: "_id", Value: id},
	}
	var FindUserById *domain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&FindUserById)
	return FindUserById, errCollUsers
}

func (u *UserRepository) SendConfirmationEmail(Email string, Token string) error {

	// confirmationLink := fmt.Sprintf("https://tudominio.com/confirm?token=%s", Token)

	// from := mail.NewEmail("Nombre de remitente", "noreply@tudominio.com")
	// subject := "Confirmación de correo electrónico"
	// to := mail.NewEmail("", Email)
	// content := mail.NewContent("text/plain", "Por favor, confirma tu correo electrónico haciendo clic en el siguiente enlace: "+confirmationLink)
	// message := mail.NewV3MailInit(from, subject, to, content)
	// SENDGRIDAPIKEY := config.SENDGRIDAPIKEY
	// sg := sendgrid.NewSendClient(SENDGRIDAPIKEY)

	// _, err := sg.Send(message)
	return nil
}
func (u *UserRepository) UpdateConfirmationEmailToken(user *domain.User) error {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	FindUserInDb := bson.D{
		{Key: "NameUser", Value: user.NameUser},
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "EmailConfirmation", Value: true}}}}
	_, err := GoMongoDBCollUsers.UpdateOne(context.Background(), FindUserInDb, update)
	if err != nil {
		return err
	}

	return nil
}

// create Stream documment
func (u *UserRepository) CreateStreamUser(user *domain.User, id primitive.ObjectID) error {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	newStream := streamdomain.Stream{
		StreamerID:         id,
		Streamer:           user.NameUser,
		StreamTitle:        "Your Title",
		StreamCategory:     "Charlando",
		StreamNotification: user.NameUser + " is live!",
		StreamerAvatar:     user.Avatar,
		StreamTag:          []string{"Español"},
		StreamIdiom:        "Español",
		StreamLikes:        []string{},
		Timestamp:          time.Now(),
		EmotesChat:         map[string]string{},
	}
	_, errInsertOne := GoMongoDBCollUsers.InsertOne(context.Background(), newStream)
	return errInsertOne
}

// follow
func (u *UserRepository) FollowUser(IdUserTokenP, followedUserID primitive.ObjectID) error {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	// Agregar el followedUserID al campo Following del usuario que sigue
	updateFollowing := bson.D{
		{Key: "$addToSet", Value: bson.D{{Key: "Following", Value: followedUserID}}},
	}
	_, errUpdateFollowing := GoMongoDBCollUsers.UpdateOne(context.Background(), bson.D{{Key: "_id", Value: IdUserTokenP}}, updateFollowing)
	if errUpdateFollowing != nil {
		return errUpdateFollowing
	}

	// Agregar el IdUserTokenP al campo Followers del usuario seguido
	updateFollowers := bson.D{
		{Key: "$addToSet", Value: bson.D{{Key: "Followers", Value: IdUserTokenP}}},
	}
	_, errUpdateFollowers := GoMongoDBCollUsers.UpdateOne(context.Background(), bson.D{{Key: "_id", Value: followedUserID}}, updateFollowers)
	if errUpdateFollowers != nil {
		return errUpdateFollowers
	}

	return nil
}

func (u *UserRepository) UnfollowUser(userID, unfollowedUserID primitive.ObjectID) error {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	updateFollowing := bson.D{
		{Key: "$pull", Value: bson.D{{Key: "Following", Value: unfollowedUserID}}},
	}
	_, errUpdateFollowing := GoMongoDBCollUsers.UpdateOne(context.Background(), bson.D{{Key: "_id", Value: userID}}, updateFollowing)
	if errUpdateFollowing != nil {
		return errUpdateFollowing
	}

	updateFollowers := bson.D{
		{Key: "$pull", Value: bson.D{{Key: "Followers", Value: userID}}},
	}
	_, errUpdateFollowers := GoMongoDBCollUsers.UpdateOne(context.Background(), bson.D{{Key: "_id", Value: unfollowedUserID}}, updateFollowers)
	if errUpdateFollowers != nil {
		return errUpdateFollowers
	}

	return nil
}
