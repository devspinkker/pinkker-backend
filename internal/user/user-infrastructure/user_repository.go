package userinfrastructure

import (
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	domain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/authGoogleAuthenticator"
	"PINKKER-BACKEND/pkg/helpers"
	"PINKKER-BACKEND/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (u *UserRepository) SetTOTPSecret(ctx context.Context, userID primitive.ObjectID, secret string) error {
	existingSecret, err := u.GetTOTPSecret(ctx, userID)
	if err != nil {
		return err
	}
	if existingSecret != "" {
		return nil
	}

	usersCollection := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"TOTPSecret": secret}}
	_, err = usersCollection.UpdateOne(ctx, filter, update)
	return err
}

func (u *UserRepository) GetTOTPSecret(ctx context.Context, userID primitive.ObjectID) (string, error) {
	usersCollection := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	filter := bson.M{"_id": userID}
	var user domain.User
	err := usersCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return "", err
	}
	return user.TOTPSecret, nil
}

func (u *UserRepository) ValidateTOTPCode(ctx context.Context, userID primitive.ObjectID, code string) (bool, error) {
	secret, err := u.GetTOTPSecret(ctx, userID)
	if err != nil {
		return false, err
	}
	return authGoogleAuthenticator.ValidateCode(secret, code), nil
}

func (u *UserRepository) FindNameUserNoSensitiveInformation(NameUser string, Email string) (*domain.GetUser, error) {
	var filter primitive.D
	if Email == "" {
		filter = bson.D{
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "NameUser", Value: NameUser}},
				bson.D{{Key: "NameUser", Value: primitive.Regex{Pattern: NameUser, Options: "i"}}},
			}},
		}
	} else {
		filter = bson.D{
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "NameUser", Value: NameUser}},
				bson.D{{Key: "Email", Value: Email}},
			}},
		}
	}
	return u.getUser(filter)
}
func (u *UserRepository) ChangeNameUser(changeNameUser domain.ChangeNameUser) error {

	ctx := context.TODO()
	db := u.mongoClient.Database("PINKKER-BACKEND")
	if !u.doesUserExist(ctx, db, changeNameUser.NameUserRemove) {
		return fmt.Errorf("NameUserRemove does not exist")
	}
	if u.doesUserExist(ctx, db, changeNameUser.NameUserNew) {
		return fmt.Errorf("NameUserNew already exists")
	}

	err := u.updateUserNames(ctx, db, changeNameUser)
	if err != nil {
		return err
	}

	err = u.updateStreamerNames(ctx, db, changeNameUser)
	if err != nil {
		return err
	}

	err = u.updateClips(ctx, db, changeNameUser)
	if err != nil {
		return err
	}
	err = u.updateUserInformationInAllRooms(ctx, db, changeNameUser)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserRepository) ChangeNameUserCodeAdmin(changeNameUser domain.ChangeNameUser, id primitive.ObjectID) error {
	err := u.AutCode(id, changeNameUser.Code)
	if err != nil {
		return err
	}
	ctx := context.TODO()
	db := u.mongoClient.Database("PINKKER-BACKEND")
	if !u.doesUserExist(ctx, db, changeNameUser.NameUserRemove) {
		return fmt.Errorf("NameUserRemove does not exist")
	}
	if u.doesUserExist(ctx, db, changeNameUser.NameUserNew) {
		return fmt.Errorf("NameUserNew already exists")
	}

	err = u.updateUserNames(ctx, db, changeNameUser)
	if err != nil {
		return err
	}

	err = u.updateStreamerNames(ctx, db, changeNameUser)
	if err != nil {
		return err
	}

	err = u.updateClips(ctx, db, changeNameUser)
	if err != nil {
		return err
	}
	err = u.updateUserInformationInAllRooms(ctx, db, changeNameUser)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserRepository) doesUserExist(ctx context.Context, db *mongo.Database, nameUser string) bool {
	GoMongoDBCollUsers := db.Collection("Users")
	filter := bson.M{"NameUser": nameUser}
	count, err := GoMongoDBCollUsers.CountDocuments(ctx, filter)
	if err != nil {
		return false
	}
	return count > 0
}
func (u *UserRepository) updateUserInformationInAllRooms(ctx context.Context, db *mongo.Database, changeNameUser domain.ChangeNameUser) error {
	GoMongoDBCollUsers := db.Collection("UserInformationInAllRooms")
	userFilterTemp := bson.M{"NameUser": changeNameUser.NameUserRemove}
	updateTemp := bson.M{"$set": bson.M{"NameUser": changeNameUser.NameUserNew}}
	_, err := GoMongoDBCollUsers.UpdateOne(ctx, userFilterTemp, updateTemp)
	if err != nil {
		return fmt.Errorf("error updating user collection to NameUserNew: %v", err)
	}

	return nil
}
func (u *UserRepository) updateUserNames(ctx context.Context, db *mongo.Database, changeNameUser domain.ChangeNameUser) error {
	GoMongoDBCollUsers := db.Collection("Users")

	userFilterTemp := bson.M{"NameUser": changeNameUser.NameUserRemove}
	updateTemp := bson.M{"$set": bson.M{"NameUser": changeNameUser.NameUserNew}}
	_, err := GoMongoDBCollUsers.UpdateOne(ctx, userFilterTemp, updateTemp)
	if err != nil {
		return fmt.Errorf("error updating user collection to NameUserNew: %v", err)
	}

	return nil
}
func (u *UserRepository) updateClips(ctx context.Context, db *mongo.Database, changeNameUser domain.ChangeNameUser) error {
	GoMongoDBCollUsers := db.Collection("Clips")

	userFilterTemp := bson.M{"NameUser": changeNameUser.NameUserRemove}
	updateTemp := bson.M{"$set": bson.M{"NameUser": changeNameUser.NameUserNew}}
	_, err := GoMongoDBCollUsers.UpdateOne(ctx, userFilterTemp, updateTemp)
	if err != nil {
		return fmt.Errorf("error updating user collection to NameUserNew: %v", err)
	}

	return nil
}

func (u *UserRepository) updateStreamerNames(ctx context.Context, db *mongo.Database, changeNameUser domain.ChangeNameUser) error {
	GoMongoDBCollStreams := db.Collection("Streams")

	streamFilterTemp := bson.M{"Streamer": changeNameUser.NameUserRemove}
	updateStreamTemp := bson.M{"$set": bson.M{"Streamer": changeNameUser.NameUserNew}}
	_, err := GoMongoDBCollStreams.UpdateOne(ctx, streamFilterTemp, updateStreamTemp)
	if err != nil {
		return fmt.Errorf("error updating stream collection to NameUserNew: %v", err)
	}

	return nil
}

func (s *UserRepository) SubscribeToRoom(roomID string) *redis.PubSub {
	sub := s.redisClient.Subscribe(context.Background(), roomID)
	return sub
}

func (s *UserRepository) CloseSubscription(sub *redis.PubSub) error {
	return sub.Close()
}
func (u *UserRepository) PanelAdminPinkkerInfoUser(PanelAdminPinkkerInfoUserReq domain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) (*domain.User, streamdomain.Stream, error) {
	err := u.AutCode(id, PanelAdminPinkkerInfoUserReq.Code)
	if err != nil {
		return &domain.User{}, streamdomain.Stream{}, err
	}
	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStream := db.Collection("Streams")

	ctx := context.TODO()

	var userFilter bson.D
	if PanelAdminPinkkerInfoUserReq.IdUser != primitive.NilObjectID {
		userFilter = bson.D{{Key: "_id", Value: PanelAdminPinkkerInfoUserReq.IdUser}}
	} else if PanelAdminPinkkerInfoUserReq.NameUser != "" {
		userFilter = bson.D{{Key: "NameUser", Value: PanelAdminPinkkerInfoUserReq.NameUser}}

	} else {
		return &domain.User{}, streamdomain.Stream{}, errors.New("IdUser and NameUser are empty")
	}
	userResult, err := u.getFullUser(userFilter)
	if err != nil {
		return &domain.User{}, streamdomain.Stream{}, err
	}
	streamFilter := bson.M{"StreamerID": userResult.ID}
	var streamResult streamdomain.Stream
	err = GoMongoDBCollStream.FindOne(ctx, streamFilter).Decode(&streamResult)
	if err != nil {
		return &domain.User{}, streamdomain.Stream{}, err
	}

	return userResult, streamResult, nil
}
func (u *UserRepository) PanelAdminPinkkerbanStreamer(PanelAdminPinkkerInfoUserReq domain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) error {
	err := u.AutCode(id, PanelAdminPinkkerInfoUserReq.Code)
	if err != nil {
		return err
	}
	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")
	GoMongoDBCollStream := db.Collection("Streams")

	ctx := context.TODO()

	var userFilter bson.M
	if PanelAdminPinkkerInfoUserReq.IdUser != primitive.NilObjectID {
		userFilter = bson.M{"_id": PanelAdminPinkkerInfoUserReq.IdUser}
	} else if PanelAdminPinkkerInfoUserReq.NameUser != "" {
		userFilter = bson.M{"NameUser": PanelAdminPinkkerInfoUserReq.NameUser}
	} else {
		return errors.New("IdUser and NameUser are empty")
	}
	var userResult domain.User
	err = GoMongoDBCollUsers.FindOne(ctx, userFilter).Decode(&userResult)
	if err != nil {
		return err
	}
	streamFilter := bson.M{"StreamerID": userResult.ID}
	var streamResult streamdomain.Stream
	err = GoMongoDBCollStream.FindOne(ctx, streamFilter).Decode(&streamResult)
	if err != nil {
		return err
	}

	update := bson.M{"$set": bson.M{"Banned": true}}
	_, err = GoMongoDBCollStream.UpdateOne(ctx, streamFilter, update)
	if err != nil {
		return err
	}
	_, err = GoMongoDBCollUsers.UpdateOne(ctx, userFilter, update)
	if err != nil {
		return err
	}

	return nil
}
func (u *UserRepository) PanelAdminRemoveBanStreamer(PanelAdminPinkkerInfoUserReq domain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) error {
	err := u.AutCode(id, PanelAdminPinkkerInfoUserReq.Code)
	if err != nil {
		return err
	}
	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")
	GoMongoDBCollStream := db.Collection("Streams")

	ctx := context.TODO()

	var userFilter bson.M
	if PanelAdminPinkkerInfoUserReq.IdUser != primitive.NilObjectID {
		userFilter = bson.M{"_id": PanelAdminPinkkerInfoUserReq.IdUser}
	} else if PanelAdminPinkkerInfoUserReq.NameUser != "" {
		userFilter = bson.M{"NameUser": PanelAdminPinkkerInfoUserReq.NameUser}
	} else {
		return errors.New("IdUser and NameUser are empty")
	}
	var userResult domain.User
	err = GoMongoDBCollUsers.FindOne(ctx, userFilter).Decode(&userResult)
	if err != nil {
		return err
	}
	streamFilter := bson.M{"StreamerID": userResult.ID}
	var streamResult streamdomain.Stream
	err = GoMongoDBCollStream.FindOne(ctx, streamFilter).Decode(&streamResult)
	if err != nil {
		return err
	}

	update := bson.M{"$set": bson.M{"Banned": false}}
	_, err = GoMongoDBCollStream.UpdateOne(ctx, streamFilter, update)
	if err != nil {
		return err
	}
	_, err = GoMongoDBCollUsers.UpdateOne(ctx, userFilter, update)
	if err != nil {
		return err
	}

	return nil
}
func (u *UserRepository) PanelAdminPinkkerPartnerUser(PanelAdminPinkkerInfoUserReq domain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) error {
	// Autenticar el código
	err := u.AutCode(id, PanelAdminPinkkerInfoUserReq.Code)
	if err != nil {
		return err
	}
	ctx := context.TODO()

	// Conectar a la base de datos y obtener la colección de usuarios
	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")

	// Crear el filtro para encontrar al usuario
	var userFilter bson.M
	if PanelAdminPinkkerInfoUserReq.IdUser != primitive.NilObjectID {
		userFilter = bson.M{"_id": PanelAdminPinkkerInfoUserReq.IdUser}
	} else if PanelAdminPinkkerInfoUserReq.NameUser != "" {
		userFilter = bson.M{"NameUser": PanelAdminPinkkerInfoUserReq.NameUser}
	} else {
		return errors.New("IdUser and NameUser are empty")
	}

	var userResult domain.User
	err = GoMongoDBCollUsers.FindOne(ctx, userFilter).Decode(&userResult)
	if err != nil {
		return err
	}

	newActiveState := !userResult.Partner.Active

	update := bson.M{
		"$set": bson.M{
			"Partner.Active": newActiveState,
			"Partner.Date":   time.Now(),
		},
	}

	_, err = GoMongoDBCollUsers.UpdateOne(ctx, userFilter, update)
	if err != nil {
		return err
	}

	return nil
}
func (u *UserRepository) CreateAdmin(CreateAdmin domain.CreateAdmin, id primitive.ObjectID) error {
	err := u.AutCode(id, CreateAdmin.Code)
	if err != nil {
		return err
	}
	ctx := context.TODO()

	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")

	var userFilter bson.M
	if CreateAdmin.IdUser != primitive.NilObjectID {
		userFilter = bson.M{"_id": CreateAdmin.IdUser}
	} else if CreateAdmin.NameUser != "" {
		userFilter = bson.M{"NameUser": CreateAdmin.NameUser}
	} else {
		return errors.New("IdUser and NameUser are empty")
	}

	var userResult domain.User
	err = GoMongoDBCollUsers.FindOne(ctx, userFilter).Decode(&userResult)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"PanelAdminPinkker.Level": CreateAdmin.Level,
			"PanelAdminPinkker.Asset": true,
			"PanelAdminPinkker.Code":  CreateAdmin.NewCode,
			"PanelAdminPinkker.Date":  time.Now(),
		},
	}

	_, err = GoMongoDBCollUsers.UpdateOne(ctx, userFilter, update)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) AutCode(id primitive.ObjectID, code string) error {
	db := u.mongoClient.Database("PINKKER-BACKEND")
	collectionUsers := db.Collection("Users")
	var User domain.User

	err := collectionUsers.FindOne(context.Background(), bson.M{"_id": id}).Decode(&User)
	if err != nil {
		return err
	}

	if User.PanelAdminPinkker.Level != 1 || !User.PanelAdminPinkker.Asset || User.PanelAdminPinkker.Code != code {
		return fmt.Errorf("usuario no autorizado")
	}
	return nil
}

func (u *UserRepository) SaveUserRedis(User *domain.User) (string, error) {

	code := helpers.GenerateRandomCode()

	// Convertir el usuario a formato JSON
	userJSON, errMarshal := json.Marshal(User)
	if errMarshal != nil {
		return "", errMarshal
	}

	// Almacenar en Redis con clave como el código
	errSet := u.redisClient.Set(context.Background(), code, userJSON, 5*time.Minute).Err()
	if errSet != nil {
		return "", errSet
	}
	return code, nil
}
func (u *UserRepository) GetUserByCodeFromRedis(code string) (*domain.User, error) {

	userJSON, errGet := u.redisClient.Get(context.Background(), code).Result()
	if errGet != nil {
		return nil, errGet
	}

	var user domain.User
	errUnmarshal := json.Unmarshal([]byte(userJSON), &user)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}

	_, errDel := u.redisClient.Del(context.Background(), code).Result()
	if errDel != nil {
		return &user, nil
	}
	return &user, nil
}

func (u *UserRepository) SaveUser(User *domain.User) (primitive.ObjectID, error) {

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
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "NameUser", Value: NameUser}},
				bson.D{{Key: "NameUser", Value: primitive.Regex{Pattern: NameUser, Options: "i"}}},
			}},
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

func (u *UserRepository) GetUserByNameUserIndex(NameUser string) ([]*domain.GetUser, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "NameUser", Value: 1}},
	}
	_, err := GoMongoDBCollUsers.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "NameUser", Value: primitive.Regex{Pattern: NameUser, Options: "i"}}}

	findOptions := options.Find().SetLimit(10)

	cursor, err := GoMongoDBCollUsers.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []*domain.GetUser
	for cursor.Next(context.Background()) {
		var user domain.GetUser
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (u *UserRepository) FindUserById(id primitive.ObjectID) (*domain.User, error) {
	filter := bson.D{{Key: "_id", Value: id}}
	return u.getFullUser(filter)
}

func (u *UserRepository) GetUserBykey(key string) (*domain.GetUser, error) {
	filter := bson.D{{Key: "KeyTransmission", Value: key}}
	return u.getUser(filter)
}

func (u *UserRepository) GetUserByCmt(Cmt string) (*domain.User, error) {
	filter := bson.D{{Key: "Cmt", Value: Cmt}}
	return u.getFullUser(filter)
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
		StreamCategory:     "Just Chatting",
		ImageCategorie:     "https://res.cloudinary.com/dcj8krp42/image/upload/v1708649172/categorias/IRL_aiusyf.jpg",
		StreamNotification: user.NameUser + " is live!",
		StreamerAvatar:     user.Avatar,
		StreamTag:          []string{"Español"},
		StreamIdiom:        "Español",
		StreamLikes:        []string{},
		Timestamp:          time.Now(),
		EmotesChat:         map[string]string{},
		StreamThumbnail:    "https://res.cloudinary.com/dcj8krp42/image/upload/v1711393933/gvnemflnz904jeawxwd7.png",
		ModChat:            "",
		ModSlowMode:        0,
		Banned:             false,
	}
	_, errInsertOne := GoMongoDBCollUsers.InsertOne(context.Background(), newStream)
	return errInsertOne
}
func (u *UserRepository) EditSocialNetworks(SocialNetwork domain.SocialNetwork, id primitive.ObjectID) error {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"socialnetwork.facebook":  SocialNetwork.Facebook,
			"socialnetwork.twitter":   SocialNetwork.Twitter,
			"socialnetwork.instagram": SocialNetwork.Instagram,
			"socialnetwork.youtube":   SocialNetwork.Youtube,
			"socialnetwork.tiktok":    SocialNetwork.Tiktok,
		},
	}

	_, err := GoMongoDBCollUsers.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// follow
func (u *UserRepository) FollowUser(IdUserTokenP, followedUserID primitive.ObjectID) error {
	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")

	filterFollowe := bson.M{"_id": followedUserID}

	var userFolloer domain.User
	err := GoMongoDBCollUsers.FindOne(context.Background(), filterFollowe).Decode(&userFolloer)
	if err != nil {
		return err
	}
	filterToken := bson.M{"_id": IdUserTokenP}
	var usertoken domain.User
	err = GoMongoDBCollUsers.FindOne(context.Background(), filterToken).Decode(&usertoken)
	if err != nil {
		return err
	}
	Followingadd := domain.FollowInfo{
		Since:         time.Now(),
		Notifications: true,
		Email:         userFolloer.Email,
	}

	// Agregar followedUserID al mapa Following de IdUserTokenP
	filter := bson.M{"_id": IdUserTokenP}
	update := bson.M{"$set": bson.M{"Following." + followedUserID.Hex(): Followingadd}}

	_, err = GoMongoDBCollUsers.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	// Agregar IdUserTokenP al mapa Followers de followedUserID
	Followersadd := domain.FollowInfo{
		Since:         time.Now(),
		Notifications: true,
		Email:         usertoken.Email,
	}

	filter = bson.M{"_id": followedUserID}
	update = bson.M{"$set": bson.M{"Followers." + IdUserTokenP.Hex(): Followersadd}}

	_, err = GoMongoDBCollUsers.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	GoMongoDBCollInformationInAllRooms := db.Collection("UserInformationInAllRooms")

	var StreamInfo streamdomain.Stream
	filter = bson.M{"Streamer": userFolloer.NameUser}
	GoMongoDBCollStreams := db.Collection("Streams")
	err = GoMongoDBCollStreams.FindOne(context.Background(), filter).Decode(&StreamInfo)
	if err != nil {
		return err
	}
	filter = bson.M{"NameUser": usertoken.NameUser}
	var userInfo domain.InfoUser
	err = GoMongoDBCollInformationInAllRooms.FindOne(context.Background(), filter).Decode(&userInfo)
	if err == mongo.ErrNoDocuments {
		defaultUserFields := map[string]interface{}{
			"Room":             StreamInfo.ID,
			"Color":            "#00ccb3",
			"Vip":              false,
			"Verified":         false,
			"Moderator":        false,
			"Subscription":     primitive.ObjectID{},
			"SubscriptionInfo": domain.SubscriptionInfo{},
			"Baneado":          false,
			"TimeOut":          time.Now(),
			"EmblemasChat": map[string]string{
				"Vip":       "",
				"Moderator": "",
				"Verified":  "",
			},
		}
		userInfo = domain.InfoUser{
			Nameuser: usertoken.NameUser,
			Color:    "#00ccb3",
			Rooms:    []map[string]interface{}{defaultUserFields},
		}
		_, err := GoMongoDBCollInformationInAllRooms.InsertOne(context.Background(), userInfo)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	roomExists := false
	for _, room := range userInfo.Rooms {
		if room["Room"] == StreamInfo.ID {
			roomExists = true
			room["Following"] = Followingadd
			break
		}
	}

	if !roomExists {
		newRoom := map[string]interface{}{
			"Room":         StreamInfo.ID,
			"Vip":          false,
			"Color":        "#00ccb3",
			"Moderator":    false,
			"Verified":     false,
			"Subscription": primitive.ObjectID{},
			"Baneado":      false,
			"TimeOut":      time.Now(),
			"EmblemasChat": map[string]string{
				"Vip":       "",
				"Moderator": "",
				"Verified":  "",
			},
			"Following": Followingadd,
		}

		userInfo.Rooms = append(userInfo.Rooms, newRoom)

	}

	_, err = GoMongoDBCollInformationInAllRooms.UpdateOne(context.Background(), filter, bson.M{"$set": userInfo})
	if err != nil {
		return err
	}

	return nil
}
func (u *UserRepository) UnfollowUser(userID, unfollowedUserID primitive.ObjectID) error {
	db := u.mongoClient.Database("PINKKER-BACKEND")

	GoMongoDBCollUsers := db.Collection("Users")

	var userToken domain.User
	err := GoMongoDBCollUsers.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&userToken)
	if err != nil {
		return err
	}

	// Eliminar unfollowedUserID del mapa Following
	delete(userToken.Following, unfollowedUserID)

	_, err = GoMongoDBCollUsers.ReplaceOne(context.Background(), bson.M{"_id": userID}, userToken)
	if err != nil {
		return err
	}

	// Eliminar userID del mapa Followers del usuario que está siendo seguido
	var userUnf domain.User
	err = GoMongoDBCollUsers.FindOne(context.Background(), bson.M{"_id": unfollowedUserID}).Decode(&userUnf)
	if err != nil {
		return err
	}

	delete(userUnf.Followers, userID)

	_, err = GoMongoDBCollUsers.ReplaceOne(context.Background(), bson.M{"_id": unfollowedUserID}, userUnf)
	if err != nil {
		return err
	}

	var StreamInfo streamdomain.Stream
	filter := bson.M{"Streamer": userUnf.NameUser}
	GoMongoDBCollStreams := db.Collection("Streams")
	err = GoMongoDBCollStreams.FindOne(context.Background(), filter).Decode(&StreamInfo)
	if err != nil {
		return err
	}
	GoMongoDBCollInformationInAllRooms := db.Collection("UserInformationInAllRooms")

	filter = bson.M{"NameUser": userToken.NameUser}
	update := bson.M{"$set": bson.M{"Rooms.$[elem].Following": domain.FollowInfo{}}}
	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem.Room": StreamInfo.ID}},
	}
	opts := options.UpdateOptions{
		ArrayFilters: &arrayFilters,
	}
	_, err = GoMongoDBCollInformationInAllRooms.UpdateOne(context.Background(), filter, update, &opts)
	if err != nil {
		return err
	}

	return nil
}
func (u *UserRepository) DeleteRedisUserChatInOneRoom(userToDelete, IdRoom primitive.ObjectID) error {
	GoMongoDBColl := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreams := GoMongoDBColl.Collection("Streams")
	filter := bson.M{"StreamerID": IdRoom}
	var stream streamdomain.Stream
	err := GoMongoDBCollStreams.FindOne(context.Background(), filter).Decode(&stream)
	if err != nil {
		return err
	}
	GoMongoDBCollUsers := GoMongoDBColl.Collection("Users")

	filter = bson.M{"_id": userToDelete}

	var userFolloer domain.User
	err = GoMongoDBCollUsers.FindOne(context.Background(), filter).Decode(&userFolloer)
	if err != nil {
		return err
	}
	userHashKey := "userInformation:" + userFolloer.NameUser + ":inTheRoom:" + stream.ID.Hex()
	_, err = u.redisClient.Del(context.Background(), userHashKey).Result()
	return err
}
func (u *UserRepository) GetWebSocketClientsInRoom(roomID string) ([]*websocket.Conn, error) {
	clients, err := utils.NewChatService().GetWebSocketClientsInRoom(roomID)

	return clients, err
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
func (u *UserRepository) EditPasswordHast(passwordHash string, id primitive.ObjectID) error {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"PasswordHash": passwordHash,
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
	if err != nil {
		return err
	}
	filterStream := bson.M{"StreamerID": id}
	updateStream := bson.M{
		"$set": bson.M{
			"StreamerAvatar": avatar,
		},
	}

	_, err = GoMongoDBCollStreams.UpdateOne(context.TODO(), filterStream, updateStream)

	return err
}
func (u *UserRepository) RedisSaveAccountRecoveryCode(code string, user domain.User) error {
	userJSON, errMarshal := json.Marshal(user)
	if errMarshal != nil {
		return errMarshal
	}

	err := u.redisClient.Set(context.Background(), code, userJSON, 5*time.Minute).Err()

	return err
}

func (u *UserRepository) getUser(filter bson.D) (*domain.GetUser, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "FollowersCount", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$objectToArray", Value: bson.D{
						{Key: "$ifNull", Value: bson.A{"$Followers", bson.D{}}},
					}},
				}},
			}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Followers", Value: 0}, // Excluir el campo Followers si es necesario
		}}},
	}

	cursor, err := GoMongoDBCollUsers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var user domain.GetUser
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserRepository) getFullUser(filter bson.D) (*domain.User, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "FollowersCount", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$objectToArray", Value: bson.D{
						{Key: "$ifNull", Value: bson.A{"$Followers", bson.D{}}},
					}},
				}},
			}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Followers", Value: 0}, // Excluir el campo Followers si es necesario
		}}},
	}

	cursor, err := GoMongoDBCollUsers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var user domain.User
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
	}

	return &user, nil
}
