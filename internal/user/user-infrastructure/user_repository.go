package userinfrastructure

import (
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	domain "PINKKER-BACKEND/internal/user/user-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"errors"
	"fmt"
	"strings"
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
func (u *UserRepository) GetUserBykey(key string) (*domain.User, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var FindUserInDb primitive.D
	FindUserInDb = bson.D{
		{Key: "KeyTransmission", Value: key},
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

func (u *UserRepository) FindEmailForOauth2Updata(user *domain.Google_callback_Complete_Profile_And_Username) (*domain.User, error) {
	NameUserLower := strings.ToLower(user.NameUser)
	_, err := u.FindNameUser(NameUserLower, "")
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
		GoMongoDBColl := u.mongoClient.Database("PINKKER-BACKEND")

		GoMongoDBCollUsers := GoMongoDBColl.Collection("Users")
		GoMongoDBCollStreams := GoMongoDBColl.Collection("Streams")

		filter := bson.M{"Email": user.Email}

		var existingUser *domain.User
		err = GoMongoDBCollUsers.FindOne(context.Background(), filter).Decode(&existingUser)
		if err != nil {
			return nil, err
		}
		if existingUser.NameUser != "" {
			return nil, errors.New("NameUser exists")
		}

		update := bson.M{
			"$set": bson.M{
				"NameUser":     user.NameUser,
				"PasswordHash": "",
				"Email":        user.Email,
				"Pais":         user.Pais,
				"Ciudad":       user.Ciudad,
				"Biography":    user.Biography,
				"HeadImage":    user.HeadImage,
				"BirthDate":    user.BirthDate,
				"Sex":          user.Sex,
				"Situation":    user.Situation,
				"ZodiacSign":   user.ZodiacSign,
			},
		}

		// Realizar la actualización
		_, err = GoMongoDBCollUsers.UpdateOne(context.Background(), filter, update)

		if err != nil {
			return nil, err
		}
		filterStream := bson.M{"StreamerID": existingUser.ID}
		updateStream := bson.M{
			"$set": bson.M{
				"Streamer": user.NameUser,
			},
		}

		_, err = GoMongoDBCollStreams.UpdateOne(context.Background(), filterStream, updateStream)

		if err != nil {
			return nil, err
		}
		return existingUser, nil
	}
	return nil, errors.New("nameuser exist")

}
func (u *UserRepository) EditProfile(profile domain.EditProfile, id primitive.ObjectID) error {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"Pais":       profile.Pais,
			"Ciudad":     profile.Ciudad,
			"Biography":  profile.Biography,
			"HeadImage":  profile.HeadImage,
			"BirthDate":  profile.BirthDateTime,
			"Sex":        profile.Sex,
			"Situation":  profile.Situation,
			"ZodiacSign": profile.ZodiacSign,
		},
	}
	_, err := GoMongoDBCollUsers.UpdateOne(context.TODO(), filter, update)
	return err
}
func (u *UserRepository) EditAvatar(avatar string, id primitive.ObjectID) error {
	GoMongoDB := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")
	GoMongoDBCollStreams := GoMongoDB.Collection("Streams")

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"Avatar": avatar,
		},
	}
	_, err := GoMongoDBCollUsers.UpdateOne(context.TODO(), filter, update)
	filterStream := bson.M{"StreamerID": id}
	updateStream := bson.M{
		"$set": bson.M{
			"StreamerAvatar": avatar,
		},
	}

	_, err = GoMongoDBCollStreams.UpdateOne(context.TODO(), filterStream, updateStream)

	return err
}
func (r *UserRepository) Subscription(Source, Destination primitive.ObjectID) error {
	usersCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	ctx, cancel := context.WithTimeout(context.Background(), 30*24*time.Hour)
	defer cancel()

	sourceUser, destUser, err := r.findUsersBy_ids(ctx, Source, Destination, usersCollection)
	if err != nil {
		return err
	}
	if sourceUser.Pixeles < 1000 {
		return errors.New("pixeles insufficient")
	}
	// Verificar si el usuario que recibe ya está suscrito
	var existingSubscription *userdomain.Subscription
	for _, subscription := range sourceUser.Subscriptions {
		if subscription.SubscriptionNameUser == destUser.NameUser {
			existingSubscription = &subscription
			break
		}
	}

	subscriptionStart := time.Now()
	subscriptionEnd := subscriptionStart.Add(30 * 24 * time.Hour)

	if existingSubscription == nil {
		r.addSubscription(sourceUser, destUser, subscriptionStart, subscriptionEnd)
		r.addSubscriber(destUser, sourceUser, subscriptionEnd)
	} else {
		r.updateSubscription(existingSubscription, subscriptionStart, subscriptionEnd)
	}

	if err := r.updateUserSource(ctx, sourceUser, usersCollection, Destination); err != nil {
		return err
	}

	err = r.updateUserDest(ctx, destUser, usersCollection)
	return err
}
func (r *UserRepository) findUsersBy_ids(ctx context.Context, source_id, dest_id primitive.ObjectID, usersCollection *mongo.Collection) (*userdomain.User, *userdomain.User, error) {
	var sourceUser userdomain.User
	filtersourceWallet := bson.M{
		"_id": source_id,
	}
	err := usersCollection.FindOne(ctx, filtersourceWallet).Decode(&sourceUser)
	if err != nil {
		return nil, nil, err
	}

	var destUser userdomain.User
	filterdestUserWallet := bson.M{
		"_id": dest_id,
	}

	err = usersCollection.FindOne(ctx, filterdestUserWallet).Decode(&destUser)
	if err != nil {
		return nil, nil, err
	}

	return &sourceUser, &destUser, nil
}
func (r *UserRepository) addSubscription(sourceUser *userdomain.User, destUser *userdomain.User, subscriptionStart, subscriptionEnd time.Time) {
	if sourceUser.Pixeles >= 1000 {
		sourceUser.Pixeles -= 1000
		subscription := userdomain.Subscription{
			SubscriptionNameUser: destUser.NameUser,
			SubscriptionStart:    subscriptionStart,
			SubscriptionEnd:      subscriptionEnd,
			MonthsSubscribed:     1, // Comienza en 1 mes
		}
		sourceUser.Subscriptions = append(sourceUser.Subscriptions, subscription)
	}
}

// Actualiza una suscripción existente
func (r *UserRepository) updateSubscription(subscription *userdomain.Subscription, subscriptionStart, subscriptionEnd time.Time) {

	subscription.SubscriptionStart = subscriptionStart
	subscription.SubscriptionEnd = subscriptionEnd
	// No es necesario actualizar MonthsSubscribed, ya que comenzamos desde 1
}

// Actualiza el usuario que da en MongoDB
func (r *UserRepository) updateUserSource(ctx context.Context, user *userdomain.User, usersCollection *mongo.Collection, Destination primitive.ObjectID) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"Subscriptions": user.Subscriptions,
			"Pixeles":       user.Pixeles,
		},
	}
	_, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	database := r.mongoClient.Database("PINKKER-BACKEND")

	// Obtén la información del stream
	streamFilter := bson.M{"StreamerID": Destination}
	streamSub := streamdomain.Stream{}
	err = database.Collection("Streams").FindOne(ctx, streamFilter).Decode(&streamSub)
	if err != nil {
		return err
	}

	// Actualiza la colección "UserInformationInAllRooms"
	updateRoom := bson.M{
		"$set": bson.M{
			"Rooms.$.Subscription":  "active",
			"Rooms.$.SubscribedAgo": time.Now(),
		},
	}

	_, err = database.Collection("UserInformationInAllRooms").UpdateOne(
		ctx,
		bson.M{"NameUser": user.NameUser, "Rooms.Room": streamSub.ID},
		updateRoom,
	)
	if err != nil {
		return err
	}

	return nil
}

// Actualiza el usuario que destino en MongoDB
func (r *UserRepository) updateUserDest(ctx context.Context, user *userdomain.User, usersCollection *mongo.Collection) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"Subscribers": user.Subscribers,
			"Pixeles":     user.Pixeles,
		},
	}
	valor, err := usersCollection.UpdateOne(ctx, filter, update)
	fmt.Println(valor)
	return err
}
func (r *UserRepository) addSubscriber(destUser *userdomain.User, sourceUser *userdomain.User, subscriptionEnd time.Time) {
	existingSubscriber := false
	for _, subscriber := range destUser.Subscribers {
		if subscriber.SubscriberNameUser == sourceUser.NameUser {
			existingSubscriber = true
			break
		}
	}

	if !existingSubscriber {
		subscriber := userdomain.Subscriber{
			SubscriberNameUser: sourceUser.NameUser,
			SubscriptionStart:  time.Now(),
			SubscriptionEnd:    subscriptionEnd,
		}
		destUser.Pixeles += 1000
		destUser.Subscribers = append(destUser.Subscribers, subscriber)
	}
}
