package userinfrastructure

import (
	"PINKKER-BACKEND/config"
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	donationdomain "PINKKER-BACKEND/internal/donation/donation"
	notificationsdomain "PINKKER-BACKEND/internal/notifications/notifications"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	domain "PINKKER-BACKEND/internal/user/user-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/authGoogleAuthenticator"
	"PINKKER-BACKEND/pkg/helpers"
	"PINKKER-BACKEND/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
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

// GetFollowersPaginated recupera de la base de datos los followers de un usuario y los devuelve paginados de a 6000.
// El parámetro 'page' es el número de página (empezando en 0).
func (u *UserRepository) GetFollowersPaginated(ctx context.Context, userID primitive.ObjectID, page int) ([]userdomain.Follower, error) {
	// Validar el parámetro de página.
	if page < 0 {
		return nil, errors.New("el número de página debe ser un entero no negativo")
	}

	// Obtener el documento del usuario, proyectando solo el campo Followers.
	collection := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var result struct {
		Followers map[primitive.ObjectID]userdomain.FollowInfo `bson:"Followers"`
	}
	err := collection.FindOne(ctx, bson.M{"_id": userID}, options.FindOne().SetProjection(bson.M{"Followers": 1})).Decode(&result)
	if err != nil {
		return nil, err
	}

	// Convertir el mapa de followers a un slice para facilitar la paginación.
	var followers []userdomain.Follower
	for id, info := range result.Followers {
		followers = append(followers, userdomain.Follower{
			ID:         id,
			FollowInfo: info,
		})
	}

	// Opcional: Ordenar los followers por la fecha de inicio del seguimiento (campo Since).
	sort.Slice(followers, func(i, j int) bool {
		return followers[i].FollowInfo.Since.Before(followers[j].FollowInfo.Since)
	})

	// Aplicar paginación: 6000 elementos por página.
	const pageSize = 6000
	start := page * pageSize
	if start >= len(followers) {
		// No hay más resultados.
		return []userdomain.Follower{}, nil
	}
	end := start + pageSize
	if end > len(followers) {
		end = len(followers)
	}

	return followers[start:end], nil
}
func (u *UserRepository) PurchasePinkkerPrime(userID primitive.ObjectID) (bool, error) {
	PinkkerPrimeCostStr := config.PinkkerPrimeCost()
	PinkkerPrimeCost, err := strconv.Atoi(PinkkerPrimeCostStr)
	if err != nil {
		return false, fmt.Errorf("error converting PinkkerPrime cost to integer: %v", err)
	}

	var result struct {
		Pixeles      float64 `bson:"Pixeles"`
		PinkkerPrime struct {
			SubscriptionEnd time.Time `bson:"SubscriptionEnd"`
		} `bson:"PinkkerPrime"`
	}

	collection := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	err = collection.FindOne(
		context.TODO(),
		bson.M{"_id": userID},
		options.FindOne().SetProjection(bson.M{"Pixeles": 1, "PinkkerPrime.SubscriptionEnd": 1}),
	).Decode(&result)

	if err != nil {
		return false, fmt.Errorf("error finding user: %v", err)
	}

	// Verificar si el usuario ya tiene una suscripción activa
	if time.Now().Before(result.PinkkerPrime.SubscriptionEnd) {
		return false, fmt.Errorf("user already has an active PinkkerPrime subscription until %v", result.PinkkerPrime.SubscriptionEnd)
	}

	if result.Pixeles < float64(PinkkerPrimeCost) {
		return false, fmt.Errorf("not enough Pixeles. User has %.2f Pixeles, but PinkkerPrime costs %d", result.Pixeles, PinkkerPrimeCost)
	}
	update := bson.M{
		"$inc": bson.M{"Pixeles": -PinkkerPrimeCost, "PinkkerPrime.MonthsSubscribed": 1},
		"$set": bson.M{
			"PinkkerPrime.SubscriptionStart": time.Now(),
			"PinkkerPrime.SubscriptionEnd":   time.Now().AddDate(0, 1, 0),
		},
	}

	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": userID}, update)
	if err != nil {
		return false, fmt.Errorf("error updating user with PinkkerPrime: %v", err)
	}
	u.updatePinkkerProfitPerMonth(context.TODO(), PinkkerPrimeCost)
	return true, nil
}
func (r *UserRepository) updatePinkkerProfitPerMonth(ctx context.Context, PinkkerPrimeCost int) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollMonthly := GoMongoDB.Collection("PinkkerProfitPerMonth")

	currentTime := time.Now()
	currentMonth := int(currentTime.Month())
	currentYear := currentTime.Year()

	currentDay := helpers.GetDayOfMonth(currentTime)

	startOfMonth := time.Date(currentYear, time.Month(currentMonth), 1, 0, 0, 0, 0, time.UTC)
	startOfNextMonth := time.Date(currentYear, time.Month(currentMonth+1), 1, 0, 0, 0, 0, time.UTC)

	// Filtro para el documento del mes actual
	monthlyFilter := bson.M{
		"timestamp": bson.M{
			"$gte": startOfMonth,
			"$lt":  startOfNextMonth,
		},
	}

	_, err := GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, bson.M{
		"$setOnInsert": bson.M{
			"timestamp":          currentTime,
			"days." + currentDay: PinkkerProfitPerMonthdomain.NewDefaultDay(),
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	monthlyUpdate := bson.M{
		"$inc": bson.M{
			"total":                                PinkkerPrimeCost,
			"days." + currentDay + ".pinkkerPrime": PinkkerPrimeCost,
		},
	}

	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserRepository) IsUserBlocked(nameUser string) (bool, error) {
	blockKey := fmt.Sprintf("login_blocked:%s", nameUser)

	_, err := u.redisClient.Get(context.Background(), blockKey).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error checking if user is blocked in Redis: %v", err)
	}

	return true, nil
}
func (u *UserRepository) HandleLoginFailure(nameUser string) error {
	key := fmt.Sprintf("login_failures:%s", nameUser)

	failures, err := u.redisClient.Get(context.Background(), key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("error getting login failures from Redis: %v", err)
	}

	failureCount := 0
	if failures != "" {
		failureCount, err = strconv.Atoi(failures)
		if err != nil {
			return fmt.Errorf("error converting login failures to integer: %v", err)
		}
	}

	failureCount++

	if failureCount >= 5 {
		// Set a block expiration for 15 minutes
		blockKey := fmt.Sprintf("login_blocked:%s", nameUser)
		_, err = u.redisClient.Set(context.Background(), blockKey, "blocked", 15*time.Minute).Result()
		if err != nil {
			return fmt.Errorf("error setting block key in Redis: %v", err)
		}
	} else {
		_, err = u.redisClient.Set(context.Background(), key, failureCount, 15*time.Minute).Result()
		if err != nil {
			return fmt.Errorf("error setting login failures in Redis: %v", err)
		}
	}

	return nil
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
	projection := bson.M{"TOTPSecret": 1, "_id": 0}
	var result struct {
		TOTPSecret string `bson:"TOTPSecret"`
	}
	err := usersCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.TOTPSecret, nil
}

func (u *UserRepository) ValidateTOTPCode(ctx context.Context, userID primitive.ObjectID, code string) (bool, error) {
	secret, err := u.GetTOTPSecret(ctx, userID)
	if err != nil {
		return false, err
	}
	return authGoogleAuthenticator.ValidateCode(secret, code), nil
}

func (u *UserRepository) DeleteGoogleAuthenticator(id primitive.ObjectID) error {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"TOTPSecret": "",
		},
	}
	_, err := GoMongoDBCollUsers.UpdateOne(context.TODO(), filter, update)
	return err
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

	err = u.updateUserNamesAdmin(ctx, db, changeNameUser)
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

	filter := bson.M{"NameUser": bson.M{"$regex": "^" + nameUser + "$", "$options": "i"}}

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

	// Estructura que contiene la propiedad NameUser en EditProfiile
	var existingUser struct {
		EditProfiile struct {
			NameUser time.Time `bson:"NameUser,omitempty"`
		} `bson:"EditProfiile"`
	}

	// Filtrar el usuario por su nombre de usuario actual
	userFilterTemp := bson.M{"NameUser": changeNameUser.NameUserRemove}
	err := GoMongoDBCollUsers.FindOne(ctx, userFilterTemp).Decode(&existingUser)
	if err != nil {
		return fmt.Errorf("error finding user with NameUser: %v", err)
	}

	// Verificar si han pasado más de 60 días desde la última actualización del nombre de usuario
	timeSinceLastChange := time.Since(existingUser.EditProfiile.NameUser)
	if timeSinceLastChange < 60*24*time.Hour {
		return fmt.Errorf("no puedes actualizar el nombre de usuario hasta que pasen 60 días desde el último cambio")
	}

	// Si han pasado más de 60 días, actualizamos el nombre de usuario y la fecha de actualización
	updateTemp := bson.M{
		"$set": bson.M{
			"NameUser":              changeNameUser.NameUserNew,
			"EditProfiile.NameUser": time.Now(), // Actualizamos la fecha de la última modificación del nombre de usuario
		},
	}

	_, err = GoMongoDBCollUsers.UpdateOne(ctx, userFilterTemp, updateTemp)
	if err != nil {
		return fmt.Errorf("error updating user collection to NameUserNew: %v", err)
	}

	return nil
}

func (u *UserRepository) updateUserNamesAdmin(ctx context.Context, db *mongo.Database, changeNameUser domain.ChangeNameUser) error {
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
func (u *UserRepository) PanelAdminPinkkerInfoUser(PanelAdminPinkkerInfoUserReq domain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) (*userdomain.GetUser, streamdomain.Stream, error) {
	err := u.AutCode(id, PanelAdminPinkkerInfoUserReq.Code)
	if err != nil {
		return nil, streamdomain.Stream{}, err
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
		return nil, streamdomain.Stream{}, errors.New("IdUser and NameUser are empty")
	}
	userResult, err := u.getUser(userFilter)
	if err != nil {
		return nil, streamdomain.Stream{}, err
	}
	streamFilter := bson.M{"StreamerID": userResult.ID}
	var streamResult streamdomain.Stream
	err = GoMongoDBCollStream.FindOne(ctx, streamFilter).Decode(&streamResult)
	if err != nil {
		return nil, streamdomain.Stream{}, err
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

	return u.getFullUser(FindUserInDb)
}
func (u *UserRepository) FindNameUserInternalOperation(NameUser string, Email string) (*domain.User, error) {
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

	return u.getFullUserInternalOperations(FindUserInDb)
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
func (u *UserRepository) GetUserBanInstream(key string) (bool, error) {
	filter := bson.D{{Key: "KeyTransmission", Value: key}}

	projection := bson.D{{Key: "Banned", Value: 1}, {Key: "_id", Value: 0}}

	var result struct {
		Banned bool `bson:"Banned"`
	}

	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	err := GoMongoDBCollUsers.FindOne(context.TODO(), filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return result.Banned, nil
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
		StreamCategory:     "IRL",
		ImageCategorie:     "https://www.pinkker.tv/uploads/images/categories/IRL.webp",
		StreamNotification: user.NameUser + " is live!",
		StreamerAvatar:     user.Avatar,
		StreamTag:          []string{"Español"},
		StreamIdiom:        "Español",
		StreamLikes:        []string{},
		Timestamp:          time.Now(),
		EmotesChat:         map[string]string{},
		StreamThumbnail:    "https://www.pinkker.tv/static/media/Banner2.1e5d7c2f65e418a490f3.jpg",
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

// get follows notis
func (u *UserRepository) GetRecentFollowsLastConnection(IdUserTokenP primitive.ObjectID, page int) ([]domain.FollowInfoRes, error) {
	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")

	limit := 10
	skip := (page - 1) * limit

	// Pipeline de agregación
	pipeline := bson.A{
		// 1. Filtramos por el usuario con el ID proporcionado
		bson.M{"$match": bson.M{"_id": IdUserTokenP}},
		// 2. Proyectamos los campos necesarios y convertimos el campo Followers a un arreglo
		bson.M{"$project": bson.M{
			"LastConnection": 1,
			"Followers": bson.M{
				"$objectToArray": "$Followers",
			},
		}},
		// 3. "Unwind" para descomponer el arreglo de Followers en documentos individuales
		bson.M{"$unwind": "$Followers"},
		// 4. Filtramos los seguidores que tienen la fecha 'since' mayor a la fecha 'LastConnection'
		bson.M{"$match": bson.M{
			"$expr": bson.M{
				"$gt": []interface{}{
					"$Followers.v.since", "$LastConnection",
				},
			},
		}},
		// 5. Convertimos Followers.k a ObjectId si no lo es ya
		bson.M{"$addFields": bson.M{
			"followerId": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$eq": bson.A{bson.M{"$type": "$Followers.k"}, "objectId"}},
					"then": "$Followers.k",
					"else": bson.M{"$toObjectId": "$Followers.k"},
				},
			},
		}},
		// 6. Lookup para obtener el NameUser de la colección Users basado en el ID del seguidor
		bson.M{"$lookup": bson.M{
			"from":         "Users",        // Colección Users
			"localField":   "followerId",   // ID convertido del seguidor
			"foreignField": "_id",          // Campo _id de la colección Users
			"as":           "FollowerInfo", // Nombre del campo para la información del usuario
		}},
		// 7. Descomponemos el array FollowerInfo para obtener el primer documento
		bson.M{"$unwind": bson.M{
			"path":                       "$FollowerInfo",
			"preserveNullAndEmptyArrays": true, // En caso de que no haya coincidencia
		}},
		// 8. Ordenamos los resultados por la fecha de 'since' en orden descendente
		bson.M{"$sort": bson.M{"Followers.v.since": -1}},
		// 9. Aplicamos el skip para la paginación
		bson.M{"$skip": skip},
		// 10. Aplicamos el limit para limitar la cantidad de resultados
		bson.M{"$limit": limit},
		// 11. Proyectamos los campos finales que queremos devolver
		bson.M{"$project": bson.M{
			"Email":         "$Followers.v.Email",
			"since":         "$Followers.v.since",
			"notifications": "$Followers.v.notifications",
			"NameUser":      "$FollowerInfo.NameUser", // Nombre del seguidor
			"Avatar":        "$FollowerInfo.Avatar",   // Nombre del seguidor

		}},
	}

	// Ejecutamos el pipeline de agregación
	cursor, err := GoMongoDBCollUsers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []domain.FollowInfoRes
	for cursor.Next(context.Background()) {
		var followInfo domain.FollowInfoRes
		if err := cursor.Decode(&followInfo); err != nil {
			return nil, err
		}
		results = append(results, followInfo)
	}

	// Revisamos si hubo error en el cursor
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (d *UserRepository) AllMyPixelesDonorsLastConnection(id primitive.ObjectID, page int) ([]donationdomain.ResDonation, error) {
	GoMongoDBCollDonations := d.mongoClient.Database("PINKKER-BACKEND").Collection("Donations")

	limit := 10
	skip := (page - 1) * limit

	// Pipeline de agregación
	pipeline := bson.A{
		// 1. Filtrar donaciones hechas al usuario destino (ToUser)
		bson.M{"$match": bson.M{"ToUser": id}},
		// 2. Lookup para obtener información del usuario destino (ToUser) y su LastConnection
		bson.M{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "ToUser",
			"foreignField": "_id",
			"as":           "toUserInfo",
		}},
		// 3. Descomponer el array de usuarios destino (ToUser)
		bson.M{"$unwind": "$toUserInfo"},
		// 4. Filtrar donaciones hechas después de la última conexión del usuario destino usando $expr
		// para comparar el campo TimeStamp con LastConnection
		bson.M{"$match": bson.M{
			"$expr": bson.M{
				"$gt": bson.A{"$TimeStamp", "$toUserInfo.LastConnection"},
			},
		}},
		// 5. Lookup para obtener información del usuario que hizo la donación (FromUser)
		bson.M{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "FromUser",
			"foreignField": "_id",
			"as":           "fromUserInfo",
		}},
		// 6. Descomponer el array de usuarios que hicieron la donación
		bson.M{"$unwind": "$fromUserInfo"},
		// 7. Proyectar los campos que queremos devolver, incluyendo los detalles de FromUser
		bson.M{"$project": bson.M{
			"FromUser":              "$FromUser",
			"fromUserInfo.Avatar":   "$fromUserInfo.Avatar",
			"fromUserInfo.NameUser": "$fromUserInfo.NameUser",
			"Pixeles":               1,
			"Text":                  1,
			"TimeStamp":             1,
			"Notified":              1,
			"ToUser":                1,
			"_id":                   1,
		}},
		// 8. Ordenar por la fecha de donación en orden descendente
		bson.M{"$sort": bson.M{"TimeStamp": -1}},
		// 9. Aplicar el skip para la paginación
		bson.M{"$skip": skip},
		// 10. Limitar la cantidad de resultados
		bson.M{"$limit": limit},
	}

	// Ejecutar el pipeline de agregación
	cursor, err := GoMongoDBCollDonations.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var donations []donationdomain.ResDonation
	for cursor.Next(context.Background()) {
		var donation donationdomain.ResDonation
		if err := cursor.Decode(&donation); err != nil {
			return nil, err
		}
		donations = append(donations, donation)
	}

	// Verificar si hubo error en el cursor
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Si no se encontraron resultados, devolver un error
	if len(donations) == 0 {
		return nil, errors.New("no documents found")
	}

	return donations, nil
}
func (r *UserRepository) GetSubsChatLastConnection(id primitive.ObjectID, page int) ([]subscriptiondomain.ResSubscriber, error) {
	GoMongoDBCollSubscribers := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")

	limit := 10
	skip := (page - 1) * limit

	// Definimos el pipeline para la consulta de agregación
	pipeline := bson.A{
		// 1. Lookup para obtener el usuario de destino (destinationUserID)
		bson.M{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "destinationUserID",
			"foreignField": "_id",
			"as":           "user",
		}},
		// 2. Unwind para descomponer el array de usuarios
		bson.M{"$unwind": "$user"},
		// 3. Match para comparar si el TimeStamp es mayor que LastConnection
		bson.M{"$match": bson.M{
			"destinationUserID": id,
			"$expr": bson.M{
				"$gt": bson.A{"$TimeStamp", "$user.LastConnection"},
			},
		}},
		// 4. Ordenamos por la fecha de inicio de suscripción en orden descendente
		bson.M{"$sort": bson.M{"TimeStamp": -1}},
		// 5. Aplicamos el skip para la paginación
		bson.M{"$skip": skip},
		// 6. Limitamos la cantidad de resultados
		bson.M{"$limit": limit},
		// 7. Lookup para obtener información del usuario que inició la suscripción (sourceUserID)
		bson.M{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "sourceUserID",
			"foreignField": "_id",
			"as":           "FromUserInfo",
		}},
		// 8. Unwind para descomponer el array de usuarios
		bson.M{"$unwind": "$FromUserInfo"},
		// 9. Proyectamos los campos necesarios para la respuesta final
		bson.M{"$project": bson.M{
			"SubscriberNameUser":    "$SubscriberNameUser",
			"FromUserInfo.Avatar":   "$FromUserInfo.Avatar",
			"FromUserInfo.NameUser": "$FromUserInfo.NameUser",
			"SubscriptionStart":     "$SubscriptionStart",
			"SubscriptionEnd":       "$SubscriptionEnd",
			"Notified":              "$Notified",
			"Text":                  "$Text",
			"id":                    "$_id",
		}},
	}

	// Ejecutamos el pipeline de agregación
	cursor, err := GoMongoDBCollSubscribers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var subscribers []subscriptiondomain.ResSubscriber
	for cursor.Next(context.Background()) {
		var subscriber subscriptiondomain.ResSubscriber
		if err := cursor.Decode(&subscriber); err != nil {
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}

	// Revisamos si hubo error en el cursor
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Si no se encontraron resultados, devolvemos un error
	if len(subscribers) == 0 {
		return nil, errors.New("no documents found")
	}

	return subscribers, nil
}

func (u *UserRepository) UpdateLastConnection(userID primitive.ObjectID) error {
	db := u.mongoClient.Database("PINKKER-BACKEND")
	usersCollection := db.Collection("Users")

	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"LastConnection": time.Now(),
		},
	}

	_, err := usersCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// last notificaciones

// esto se va
func (u *UserRepository) GetRecentFollowsBeforeFirstConnection(IdUserTokenP primitive.ObjectID, page int) ([]domain.FollowInfoRes, error) {
	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")

	limit := 10
	skip := (page - 1) * limit

	// Pipeline de agregación
	pipeline := bson.A{
		// 1. Filtramos por el usuario con el ID proporcionado
		bson.M{"$match": bson.M{"_id": IdUserTokenP}},
		// 2. Proyectamos los campos necesarios y convertimos el campo Followers a un arreglo
		bson.M{"$project": bson.M{
			"LastConnection": 1,
			"Followers": bson.M{
				"$objectToArray": "$Followers",
			},
		}},
		// 3. "Unwind" para descomponer el arreglo de Followers en documentos individuales
		bson.M{"$unwind": "$Followers"},
		// 4. Filtramos los seguidores que tienen la fecha 'since' mayor a la fecha 'LastConnection'
		bson.M{"$match": bson.M{
			"$expr": bson.M{
				"$lt": bson.A{"$Followers.v.since", "$LastConnection"},
			},
		}},
		// 5. Convertimos Followers.k a ObjectId si no lo es ya
		bson.M{"$addFields": bson.M{
			"followerId": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$eq": bson.A{bson.M{"$type": "$Followers.k"}, "objectId"}},
					"then": "$Followers.k",
					"else": bson.M{"$toObjectId": "$Followers.k"},
				},
			},
		}},
		// 6. Lookup para obtener el NameUser de la colección Users basado en el ID del seguidor
		bson.M{"$lookup": bson.M{
			"from":         "Users",        // Colección Users
			"localField":   "followerId",   // ID convertido del seguidor
			"foreignField": "_id",          // Campo _id de la colección Users
			"as":           "FollowerInfo", // Nombre del campo para la información del usuario
		}},
		// 7. Descomponemos el array FollowerInfo para obtener el primer documento
		bson.M{"$unwind": bson.M{
			"path":                       "$FollowerInfo",
			"preserveNullAndEmptyArrays": true, // En caso de que no haya coincidencia
		}},
		// 8. Ordenamos los resultados por la fecha de 'since' en orden descendente
		bson.M{"$sort": bson.M{"Followers.v.since": -1}},
		// 9. Aplicamos el skip para la paginación
		bson.M{"$skip": skip},
		// 10. Aplicamos el limit para limitar la cantidad de resultados
		bson.M{"$limit": limit},
		// 11. Proyectamos los campos finales que queremos devolver
		bson.M{"$project": bson.M{
			"Email":         "$Followers.v.Email",
			"since":         "$Followers.v.since",
			"notifications": "$Followers.v.notifications",
			"NameUser":      "$FollowerInfo.NameUser", // Nombre del seguidor
			"Avatar":        "$FollowerInfo.Avatar",   // Nombre del seguidor

		}},
	}

	cursor, err := GoMongoDBCollUsers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var follows []domain.FollowInfoRes
	for cursor.Next(context.Background()) {
		var follow domain.FollowInfoRes
		if err := cursor.Decode(&follow); err != nil {
			return nil, err
		}
		follows = append(follows, follow)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(follows) == 0 {
		return nil, errors.New("no documents found")
	}

	return follows, nil
}

func (d *UserRepository) AllMyPixelesDonorsBeforeFirstConnection(id primitive.ObjectID, page int) ([]donationdomain.ResDonation, error) {
	GoMongoDBCollDonations := d.mongoClient.Database("PINKKER-BACKEND").Collection("Donations")

	limit := 10
	skip := (page - 1) * limit

	// Pipeline de agregación
	// Pipeline de agregación
	pipeline := bson.A{
		// 1. Filtrar donaciones hechas al usuario destino (ToUser)
		bson.M{"$match": bson.M{"ToUser": id}},
		// 2. Lookup para obtener información del usuario destino (ToUser) y su LastConnection
		bson.M{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "ToUser",
			"foreignField": "_id",
			"as":           "toUserInfo",
		}},
		// 3. Descomponer el array de usuarios destino (ToUser)
		bson.M{"$unwind": "$toUserInfo"},
		// 4. Filtrar donaciones hechas después de la última conexión del usuario destino usando $expr
		// para comparar el campo TimeStamp con LastConnection
		bson.M{"$match": bson.M{
			"$expr": bson.M{
				"$lt": bson.A{"$TimeStamp", "$toUserInfo.LastConnection"},
			},
		}},
		// 5. Lookup para obtener información del usuario que hizo la donación (FromUser)
		bson.M{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "FromUser",
			"foreignField": "_id",
			"as":           "fromUserInfo",
		}},
		// 6. Descomponer el array de usuarios que hicieron la donación
		bson.M{"$unwind": "$fromUserInfo"},
		// 7. Proyectar los campos que queremos devolver, incluyendo los detalles de FromUser
		bson.M{"$project": bson.M{
			"FromUser":              "$FromUser",
			"fromUserInfo.Avatar":   "$fromUserInfo.Avatar",
			"fromUserInfo.NameUser": "$fromUserInfo.NameUser",
			"Pixeles":               1,
			"Text":                  1,
			"TimeStamp":             1,
			"Notified":              1,
			"ToUser":                1,
			"_id":                   1,
		}},
		// 8. Ordenar por la fecha de donación en orden descendente
		bson.M{"$sort": bson.M{"TimeStamp": -1}},
		// 9. Aplicar el skip para la paginación
		bson.M{"$skip": skip},
		// 10. Limitar la cantidad de resultados
		bson.M{"$limit": limit},
	}

	cursor, err := GoMongoDBCollDonations.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var donations []donationdomain.ResDonation
	for cursor.Next(context.Background()) {
		var donation donationdomain.ResDonation
		if err := cursor.Decode(&donation); err != nil {
			return nil, err
		}
		donations = append(donations, donation)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(donations) == 0 {
		return nil, errors.New("no documents found")
	}

	return donations, nil
}

func (r *UserRepository) GetSubsChatBeforeFirstConnection(id primitive.ObjectID, page int) ([]subscriptiondomain.ResSubscriber, error) {
	GoMongoDBCollSubscribers := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")

	limit := 10
	skip := (page - 1) * limit

	// Definimos el pipeline para la consulta de agregación
	// Definimos el pipeline para la consulta de agregación
	pipeline := bson.A{
		// 1. Lookup para obtener el usuario de destino (destinationUserID)
		bson.M{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "destinationUserID",
			"foreignField": "_id",
			"as":           "user",
		}},
		// 2. Unwind para descomponer el array de usuarios
		bson.M{"$unwind": "$user"},
		// 3. Match para comparar si el TimeStamp es mayor que LastConnection
		bson.M{"$match": bson.M{
			"destinationUserID": id,
			"$expr": bson.M{
				"$lt": bson.A{"$TimeStamp", "$user.LastConnection"},
			},
		}},
		// 4. Ordenamos por la fecha de inicio de suscripción en orden descendente
		bson.M{"$sort": bson.M{"TimeStamp": -1}},
		// 5. Aplicamos el skip para la paginación
		bson.M{"$skip": skip},
		// 6. Limitamos la cantidad de resultados
		bson.M{"$limit": limit},
		// 7. Lookup para obtener información del usuario que inició la suscripción (sourceUserID)
		bson.M{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "sourceUserID",
			"foreignField": "_id",
			"as":           "FromUserInfo",
		}},
		// 8. Unwind para descomponer el array de usuarios
		bson.M{"$unwind": "$FromUserInfo"},
		// 9. Proyectamos los campos necesarios para la respuesta final
		bson.M{"$project": bson.M{
			"SubscriberNameUser":    "$SubscriberNameUser",
			"FromUserInfo.Avatar":   "$FromUserInfo.Avatar",
			"FromUserInfo.NameUser": "$FromUserInfo.NameUser",
			"SubscriptionStart":     "$SubscriptionStart",
			"SubscriptionEnd":       "$SubscriptionEnd",
			"Notified":              "$Notified",
			"Text":                  "$Text",
			"id":                    "$_id",
		}},
	}

	cursor, err := GoMongoDBCollSubscribers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var subscribers []subscriptiondomain.ResSubscriber
	for cursor.Next(context.Background()) {
		var subscriber subscriptiondomain.ResSubscriber
		if err := cursor.Decode(&subscriber); err != nil {
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if len(subscribers) == 0 {
		return nil, errors.New("no documents found")
	}

	return subscribers, nil
}

// follow
func (u *UserRepository) FollowUser(IdUserTokenP, followedUserID primitive.ObjectID) (string, error) {
	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")

	// Buscar al usuario seguido (followedUserID)
	filterFollowe := bson.M{"_id": followedUserID}

	var userFolloer domain.GetUser
	err := GoMongoDBCollUsers.FindOne(context.Background(), filterFollowe).Decode(&userFolloer)
	if err != nil {
		return "", err
	}

	// Buscar al usuario que sigue (IdUserTokenP)
	filterToken := bson.M{"_id": IdUserTokenP}
	var usertoken domain.GetUser
	err = GoMongoDBCollUsers.FindOne(context.Background(), filterToken).Decode(&usertoken)
	if err != nil {
		return "", err
	}
	// Obtener el Avatar del usuario seguido
	avatar := usertoken.Avatar

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
		return "", err
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
		return "", err
	}

	GoMongoDBCollInformationInAllRooms := db.Collection("UserInformationInAllRooms")

	var StreamInfo streamdomain.Stream
	filter = bson.M{"Streamer": userFolloer.NameUser}
	GoMongoDBCollStreams := db.Collection("Streams")
	err = GoMongoDBCollStreams.FindOne(context.Background(), filter).Decode(&StreamInfo)
	if err != nil {
		return "", err
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
			return "", err
		}
	} else if err != nil {
		return "", err
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
		return "", err
	}

	// Devolver el Avatar del usuario seguido
	return avatar, nil
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
func (u *UserRepository) IsFollowing(IdUserTokenP, followedUserID primitive.ObjectID) (bool, error) {
	db := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")

	// Consulta MongoDB para verificar si el usuario sigue a followedUserID
	filter := bson.M{
		"_id":                             followedUserID,
		"Following." + IdUserTokenP.Hex(): bson.M{"$exists": true},
	}

	count, err := GoMongoDBCollUsers.CountDocuments(context.Background(), filter)
	if err != nil {
		return false, err
	}

	// Si se encontró al menos un documento, significa que lo sigue
	return count > 0, nil
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
				"PasswordHash": user.Password,
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

	var existingUser struct {
		EditProfiile struct {
			Biography time.Time `bson:"Biography,omitempty"`
		} `bson:"EditProfiile"`
	}

	filter := bson.M{"_id": id}
	err := GoMongoDBCollUsers.FindOne(context.TODO(), filter).Decode(&existingUser)
	if err != nil {
		return err
	}

	// timeSinceLastChange := time.Since(existingUser.EditProfiile.Biography)
	// if timeSinceLastChange < 15*24*time.Hour {
	// 	return fmt.Errorf("no puedes actualizar la biografía hasta que pasen 60 días desde el último cambio")
	// }

	update := bson.M{
		"$set": bson.M{
			"Pais":                   profile.Pais,
			"Ciudad":                 profile.Ciudad,
			"Biography":              profile.Biography,
			"EditProfiile.Biography": time.Now(),
			"HeadImage":              profile.HeadImage,
			"BirthDate":              profile.BirthDateTime,
			"Sex":                    profile.Sex,
			"Situation":              profile.Situation,
			"ZodiacSign":             profile.ZodiacSign,
		},
	}

	_, err = GoMongoDBCollUsers.UpdateOne(context.TODO(), filter, update)
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
func (u *UserRepository) EditBanner(Banner string, id primitive.ObjectID) error {
	GoMongoDB := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreams := GoMongoDB.Collection("Users")
	filterStream := bson.M{"_id": id}
	updateStream := bson.M{
		"$set": bson.M{
			"Banner": Banner,
		},
	}

	_, err := GoMongoDBCollStreams.UpdateOne(context.TODO(), filterStream, updateStream)

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

func (u *UserRepository) RedisSaveChangeGoogleAuthenticatorCode(code string, user domain.User) error {
	userJSON, errMarshal := json.Marshal(user)
	if errMarshal != nil {
		return errMarshal
	}

	err := u.redisClient.Set(context.Background(), code, userJSON, 5*time.Minute).Err()

	return err
}

func (u *UserRepository) RedisGetChangeGoogleAuthenticatorCode(code string) (*domain.User, error) {

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

func (u *UserRepository) getUser(filter bson.D) (*userdomain.GetUser, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	currentTime := time.Now()

	pipeline := mongo.Pipeline{
		// Filtra el usuario basado en el filtro proporcionado
		bson.D{{Key: "$match", Value: filter}},
		// Agrega campos adicionales como FollowersCount, FollowingCount, SubscribersCount
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "FollowersCount", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$ifNull", Value: bson.A{
						bson.D{{Key: "$objectToArray", Value: "$Followers"}},
						bson.A{},
					}},
				}},
			}},
			{Key: "FollowingCount", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$ifNull", Value: bson.A{
						bson.D{{Key: "$objectToArray", Value: "$Following"}},
						bson.A{},
					}},
				}},
			}},
			{Key: "SubscribersCount", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$ifNull", Value: bson.A{"$Subscribers", bson.A{}}},
				}},
			}},
		}}},
		// Realiza un lookup en la colección de suscripciones
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Subscriptions"},
			{Key: "let", Value: bson.D{{Key: "userID", Value: "$_id"}}}, // Pasa el ID del usuario actual
			{Key: "pipeline", Value: mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{
						{Key: "$and", Value: bson.A{
							bson.D{{Key: "$eq", Value: bson.A{"$destinationUserID", "$$userID"}}}, // Coincide el userID con el destinationUserID
							bson.D{{Key: "$gt", Value: bson.A{"$SubscriptionEnd", currentTime}}},  // Verifica que la suscripción esté activa
						}},
					}},
				}}},
				bson.D{{Key: "$count", Value: "activeSubscriptionsCount"}},
			}},
			{Key: "as", Value: "SubscriptionData"},
		}}},
		// Agrega el SubscriptionCount desde SubscriptionData
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "SubscriptionCount", Value: bson.D{
				{Key: "$ifNull", Value: bson.A{
					bson.D{{Key: "$arrayElemAt", Value: bson.A{"$SubscriptionData.activeSubscriptionsCount", 0}}},
					0,
				}},
			}},
		}}},
		// Proyección para excluir campos innecesarios
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Followers", Value: 0},
			{Key: "Subscribers", Value: 0},
			{Key: "SubscriptionData", Value: 0}, // Excluir los datos de lookup
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
func (u *UserRepository) getFullUserInternalOperations(filter bson.D) (*domain.User, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Followers", Value: 0},
			{Key: "Subscribers", Value: 0},
			{Key: "ClipsComment", Value: 0},
			{Key: "Following", Value: 0},
			{Key: "ClipsLikes", Value: 0},
			{Key: "Subscriptions", Value: 0},
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
	} else {
		// Si no se encuentra un usuario, devuelve mongo.ErrNoDocuments
		return nil, mongo.ErrNoDocuments
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

			{Key: "Subscribers", Value: 0}, //new
			{Key: "PasswordHash", Value: 0},

			{Key: "TOTPSecret", Value: 0},
			{Key: "ClipsComment", Value: 0},
			{Key: "Following", Value: 0},
			{Key: "ClipsLikes", Value: 0},
			{Key: "Subscriptions", Value: 0},
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
	} else {
		return nil, mongo.ErrNoDocuments
	}

	return &user, nil
}
func (u *UserRepository) getActiveSubscriptions(userID primitive.ObjectID) (int, error) {
	GoMongoDBCollSubscriptions := u.mongoClient.Database("PINKKER-BACKEND").Collection("Subscriptions")

	currentTime := time.Now()

	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "destinationUserID", Value: userID},
				{Key: "SubscriptionEnd", Value: bson.D{{Key: "$gt", Value: currentTime}}}, // Filtra solo suscripciones activas
			}},
		},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "activeSubscriptionsCount", Value: bson.D{{Key: "$sum", Value: 1}}},
			}},
		},
	}

	cursor, err := GoMongoDBCollSubscriptions.Aggregate(context.Background(), pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.Background())

	var result struct {
		ActiveSubscriptionsCount int `bson:"activeSubscriptionsCount"`
	}

	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
	}

	return result.ActiveSubscriptionsCount, nil
}

func (u *UserRepository) GetFollowsUser(ctx context.Context, idT primitive.ObjectID, userCollection *mongo.Collection) ([]primitive.ObjectID, error) {

	// Pipeline para obtener los usuarios seguidos por el usuario actual (idT)
	userPipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.M{"_id": idT}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Following", Value: bson.D{{Key: "$objectToArray", Value: "$Following"}}},
		}}},

		bson.D{{Key: "$unwind", Value: "$Following"}},
		bson.D{{Key: "$limit", Value: 100}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Following.k", Value: 1},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{}},
			{Key: "followingIDs", Value: bson.D{{Key: "$push", Value: "$Following.k"}}},
		}}},
	}

	// Obtener la lista de usuarios seguidos
	cursor, err := userCollection.Aggregate(ctx, userPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var userResult struct {
		FollowingIDs []primitive.ObjectID `bson:"followingIDs"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&userResult); err != nil {
			return nil, err
		}
	}

	return userResult.FollowingIDs, nil
}
func (u *UserRepository) GetRecommendedUsers(idT primitive.ObjectID, excludeIDs []primitive.ObjectID, limit int) ([]userdomain.GetUser, error) {
	ctx := context.Background()
	db := u.mongoClient.Database("PINKKER-BACKEND")
	collUsers := db.Collection("Users")

	excludedIDs := append(excludeIDs, idT)
	excludeFilter := bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: excludedIDs}}}}

	last24Hours := time.Now().Add(-24 * time.Hour)

	followingIDs, err := u.GetFollowsUser(ctx, idT, collUsers)
	if err != nil {
		return nil, err
	}
	relevantUsers, err := u.getRelevantUsers(ctx, idT, collUsers, excludeFilter, last24Hours, limit, followingIDs, *collUsers)
	if err != nil {
		return nil, err
	}

	// Calcular el nuevo límite para el pipeline secundario
	newLimit := limit - len(relevantUsers)
	if newLimit > 0 {
		var recommendedUserIDs []primitive.ObjectID
		for _, user := range relevantUsers {
			recommendedUserIDs = append(recommendedUserIDs, user.ID)
		}

		// Actualizar el filtro de exclusión
		excludeFilter := bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "$nin", Value: append(excludedIDs, recommendedUserIDs...)},
			}},
		}

		// Obtener usuarios aleatorios si no se ha cumplido el límite
		randomUsers, err := u.getRandomUsers(ctx, idT, collUsers, excludeFilter, newLimit, followingIDs)
		if err != nil {
			return nil, err
		}
		relevantUsers = append(relevantUsers, randomUsers...)
	}

	return relevantUsers, nil
}

func (u *UserRepository) getRelevantUsers(ctx context.Context, idT primitive.ObjectID, collUsers *mongo.Collection, excludeFilter bson.D, last24Hours time.Time, limit int, followingIDs []primitive.ObjectID, userCollection mongo.Collection) ([]userdomain.GetUser, error) {

	// Pipeline para obtener usuarios que son seguidos por los usuarios en followingIDs
	userRelevancePipeline := bson.A{
		// Filtrar usuarios activos en las últimas 24 horas
		bson.D{{Key: "$match", Value: bson.M{
			"Online": true,
		}}},
		bson.D{{Key: "$match", Value: bson.M{
			"_id": bson.M{"$nin": followingIDs},
		}}},

		bson.D{{Key: "$match", Value: excludeFilter}},
		// Filtrar para obtener los usuarios que son seguidos por los usuarios de followingIDs
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "isFollowedByFollowingIDs", Value: bson.D{
				{Key: "$in", Value: bson.A{"$_id", followingIDs}},
			}},
		}}},
		// Dejar solo los usuarios que son seguidos por al menos uno de los usuarios de followingIDs
		bson.D{{Key: "$match", Value: bson.M{"isFollowedByFollowingIDs": true}}},
		// Calcular la cantidad de seguidores y suscriptores
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "followersCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Followers", bson.A{}}}}}}},
			{Key: "subscriptionsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Subscriptions", bson.A{}}}}}}},
		}}},
		// Ordenar por la cantidad de seguidores o cualquier otra métrica de relevancia
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "followersCount", Value: -1},
		}}},
		bson.D{{Key: "$limit", Value: limit}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "FullName", Value: 1},
			{Key: "Avatar", Value: 1},
			{Key: "NameUser", Value: 1},
			{Key: "followersCount", Value: 1},
			{Key: "subscriptionsCount", Value: 1},
			{Key: "Online", Value: 1},
		}}},
	}

	cursor, err := userCollection.Aggregate(ctx, userRelevancePipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var relevantUsers []userdomain.GetUser
	for cursor.Next(ctx) {
		var user userdomain.GetUser
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		relevantUsers = append(relevantUsers, user)
	}

	return relevantUsers, nil
}
func (u *UserRepository) getRandomUsers(ctx context.Context, idT primitive.ObjectID, collUsers *mongo.Collection, excludeFilter bson.D, limit int, followingIDs []primitive.ObjectID) ([]userdomain.GetUser, error) {
	randomUserPipeline := bson.A{

		// Filtrar usuarios excluidos
		bson.D{{Key: "$match", Value: excludeFilter}},
		bson.D{{Key: "$match", Value: bson.M{
			"_id": bson.M{"$nin": followingIDs},
		}}},

		// Ordenar aleatoriamente
		bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit}}}},
		// Seleccionar campos relevantes
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "FullName", Value: 1},
			{Key: "Avatar", Value: 1},
			{Key: "NameUser", Value: 1},
			{Key: "followersCount", Value: 1},
			{Key: "subscriptionsCount", Value: 1},
			{Key: "Online", Value: 1},
		}}},
	}

	cursor, err := collUsers.Aggregate(ctx, randomUserPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var randomUsers []userdomain.GetUser
	for cursor.Next(ctx) {
		var user userdomain.GetUser
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		randomUsers = append(randomUsers, user)
	}

	return randomUsers, nil
}
func (u *UserRepository) getUserAndCheckFollow(filter bson.D, id primitive.ObjectID) (*userdomain.GetUser, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	currentTime := time.Now()

	pipeline := mongo.Pipeline{
		// Filtra el usuario basado en el filtro proporcionado
		bson.D{{Key: "$match", Value: filter}},
		// Agrega campos adicionales como FollowersCount, FollowingCount, SubscribersCount
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "FollowersCount", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$ifNull", Value: bson.A{
						bson.D{{Key: "$objectToArray", Value: "$Followers"}},
						bson.A{},
					}},
				}},
			}},
			{Key: "FollowingCount", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$ifNull", Value: bson.A{
						bson.D{{Key: "$objectToArray", Value: "$Following"}},
						bson.A{},
					}},
				}},
			}},
			{Key: "SubscribersCount", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$ifNull", Value: bson.A{"$Subscribers", bson.A{}}},
				}},
			}},
		}}},
		// Verifica si el 'id' está en las claves de 'Followers'
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "isFollowedByUser", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$in", Value: bson.A{id.Hex(), bson.D{
							{Key: "$map", Value: bson.D{
								{Key: "input", Value: bson.D{{Key: "$objectToArray", Value: "$Followers"}}},
								{Key: "as", Value: "follower"},
								{Key: "in", Value: "$$follower.k"}, // La clave es el ObjectID
							}},
						}}},
					}},
					{Key: "then", Value: true},
					{Key: "else", Value: false},
				}},
			}},
		}}},
		// Realiza un lookup en la colección de suscripciones
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Subscriptions"},
			{Key: "let", Value: bson.D{{Key: "userID", Value: "$_id"}}}, // Pasa el ID del usuario actual
			{Key: "pipeline", Value: mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{
						{Key: "$and", Value: bson.A{
							bson.D{{Key: "$eq", Value: bson.A{"$destinationUserID", "$$userID"}}}, // Coincide el userID con el destinationUserID
							bson.D{{Key: "$gt", Value: bson.A{"$SubscriptionEnd", currentTime}}},  // Verifica que la suscripción esté activa
						}}}}}}},
				bson.D{{Key: "$count", Value: "activeSubscriptionsCount"}},
			}},
			{Key: "as", Value: "SubscriptionData"},
		}}},
		// Agrega el SubscriptionCount desde SubscriptionData
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "SubscriptionCount", Value: bson.D{
				{Key: "$ifNull", Value: bson.A{
					bson.D{{Key: "$arrayElemAt", Value: bson.A{"$SubscriptionData.activeSubscriptionsCount", 0}}},
					0,
				}},
			}},
		}}},
		// Proyección para excluir campos innecesarios
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Followers", Value: 0},
			{Key: "Subscribers", Value: 0},
			{Key: "SubscriptionData", Value: 0}, // Excluir los datos de lookup
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

func (r *UserRepository) GetStreamByNameUser(nameUser string) (*streamdomain.Stream, error) {
	ctx := context.Background()
	cacheKey := "stream:" + nameUser

	// Intentar buscar en Redis primero
	cachedStream, err := r.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Si encontramos el stream en la caché, lo deserializamos y lo devolvemos
		var stream streamdomain.Stream
		if err := json.Unmarshal([]byte(cachedStream), &stream); err == nil {
			return &stream, nil
		}
	}

	// Si no está en la caché o hay un error, consultamos en la base de datos
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	FindStreamInDb := bson.D{
		{Key: "Streamer", Value: nameUser},
	}
	var FindStreamsByStreamer streamdomain.Stream
	errCollStreams := GoMongoDBCollStreams.FindOne(ctx, FindStreamInDb).Decode(&FindStreamsByStreamer)
	if errCollStreams != nil {
		return nil, errCollStreams
	}

	// Serializamos el resultado y lo almacenamos en Redis
	streamData, err := json.Marshal(FindStreamsByStreamer)
	if err == nil {
		err = r.redisClient.Set(ctx, cacheKey, streamData, 5*time.Minute).Err()
		if err != nil {
			return &FindStreamsByStreamer, errors.New("user not found")
		}
	}

	return &FindStreamsByStreamer, nil
}

func (r *UserRepository) GetStreamAndUserData(nameUser string, id primitive.ObjectID, nameUserToken string) (*streamdomain.Stream, *userdomain.GetUser, *domain.UserInfo, error) {
	stream, err := r.GetStreamByNameUser(nameUser)
	if err != nil {
		return nil, nil, nil, err
	}
	filter := bson.D{
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "NameUser", Value: nameUser}},
		}},
	}
	user, err := r.getUserAndCheckFollow(filter, id)
	if err != nil {
		return nil, nil, nil, err
	}
	UserInfo, err := r.GetInfoUserInRoom(nameUserToken, stream.ID)
	return stream, user, UserInfo, err
}

func (r *UserRepository) GetInfoUserInRoom(nameUser string, getInfoUserInRoom primitive.ObjectID) (*domain.UserInfo, error) {
	database := r.mongoClient.Database("PINKKER-BACKEND")
	var room *domain.UserInfo
	filter := bson.D{
		{Key: "NameUser", Value: nameUser},
		{Key: "Rooms.Room", Value: getInfoUserInRoom},
	}

	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: filter},
		},
		{
			{Key: "$unwind", Value: "$Rooms"},
		},
		{
			{Key: "$match", Value: bson.D{{Key: "Rooms.Room", Value: getInfoUserInRoom}}},
		},
		{
			{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$Rooms"}}},
		},
	}

	// Obtener el cursor
	cursor, err := database.Collection("UserInformationInAllRooms").Aggregate(context.Background(), pipeline)
	if err != nil {
		return room, err
	}
	defer cursor.Close(context.Background())

	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&room); err != nil {
			return room, err
		}
	}
	if err := cursor.Err(); err != nil {
		return room, err
	}

	return room, nil
}

func (r *UserRepository) GetAllPendingNameUserAds(page int, userId primitive.ObjectID) ([]advertisements.Advertisements, error) {
	// Base de datos y colecciones
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Users := db.Collection("Users")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	// Buscar el usuario por userId
	var user struct {
		OwnerCommunities []primitive.ObjectID `bson:"OwnerCommunities"`
	}

	err := Users.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// Validar que el usuario tenga comunidades
	if len(user.OwnerCommunities) == 0 {
		return nil, errors.New("el usuario no tiene comunidades")
	}

	// Paginación
	limit := int64(10)
	skip := int64((page - 1) * 10)

	// Filtro para anuncios pendientes con CommunityId en OwnerCommunities
	filter := bson.M{
		"State": "pending",
		"CommunityId": bson.M{
			"$in": user.OwnerCommunities,
		},
	}

	// Opciones de consulta: paginación
	options := options.Find()
	options.SetLimit(limit)
	options.SetSkip(skip)

	// Ejecutar la consulta en Advertisements
	cursor, err := Advertisements.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decodificar los resultados
	var pendingAds []advertisements.Advertisements
	for cursor.Next(ctx) {
		var ad advertisements.Advertisements
		if err := cursor.Decode(&ad); err != nil {
			return nil, err
		}
		pendingAds = append(pendingAds, ad)
	}

	// Verificar errores del cursor
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return pendingAds, nil
}
func (r *UserRepository) AcceptOrDeleteAdvertisement(userId primitive.ObjectID, advertisementID primitive.ObjectID, action bool) error {
	// Base de datos y colecciones
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Users := db.Collection("Users")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	// Paso 1: Buscar al usuario por UserId para obtener OwnerCommunities
	var user struct {
		OwnerCommunities []primitive.ObjectID `bson:"OwnerCommunities"`
	}

	err := Users.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return errors.New("user not found")
	}

	// Paso 2: Buscar el anuncio por su ID
	var advertisement struct {
		CommunityId          primitive.ObjectID `bson:"CommunityId"`
		PricePTotalCommunity float64            `bson:"PricePTotalCommunity"` // Asegúrate que este campo es float64
		IdUser               primitive.ObjectID `bson:"IdUser"`
		State                string             `bson:"State"` // Agregamos el campo State para verificar su estado

	}

	err = Advertisements.FindOne(ctx, bson.M{"_id": advertisementID}).Decode(&advertisement)
	if err != nil {
		return errors.New("advertisement not found")
	}

	// Si el anuncio ya está aceptado, no hacer nada
	if advertisement.State == "accepted" {
		return nil // Anuncio ya aceptado, no se hace nada
	}

	// Paso 3: Verificar si el anuncio pertenece a una de las comunidades del usuario
	hasPermission := false
	for _, communityID := range user.OwnerCommunities {
		if communityID == advertisement.CommunityId {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return errors.New("user does not have permission to modify this advertisement")
	}

	// Acción cuando el anuncio es aceptado
	if action {
		// Aceptar el anuncio
		_, err := Advertisements.UpdateOne(ctx, bson.M{"_id": advertisementID}, bson.M{
			"$set": bson.M{
				"State": "accepted",
			},
		})
		if err != nil {
			return err
		}

		// Calcular el monto que se le suma al usuario (90% del total)
		amountToAdd := advertisement.PricePTotalCommunity * 0.90
		// Actualizar el saldo del usuario con el nuevo valor
		_, err = Users.UpdateOne(ctx, bson.M{"_id": userId}, bson.M{
			"$inc": bson.M{
				"Pixeles": amountToAdd, // Aumenta el saldo del usuario
			},
		})
		if err != nil {
			return err
		}
		CommissionsCommunity := advertisement.PricePTotalCommunity * 0.10
		err = r.updatePinkkerProfitPerMonthAAdCommunity(ctx, CommissionsCommunity)
		if err != nil {
			return err
		}

	} else {
		// Acción cuando el anuncio es rechazado
		// Eliminar el anuncio
		_, err := Advertisements.DeleteOne(ctx, bson.M{"_id": advertisementID})
		if err != nil {
			return err
		}

		// Devolver el dinero al creador del anuncio
		_, err = Users.UpdateOne(ctx, bson.M{"_id": advertisement.IdUser}, bson.M{
			"$inc": bson.M{
				"Pixeles": advertisement.PricePTotalCommunity, // Devolver el total al creador
			},
		})
		if err != nil {
			return err
		}

	}

	return nil
}

func (r *UserRepository) GetAllAcceptedNameUserAds(page int, userId primitive.ObjectID) ([]advertisements.Advertisements, error) {
	// Base de datos y colecciones
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Users := db.Collection("Users")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	// Buscar el usuario por userId
	var user struct {
		OwnerCommunities []primitive.ObjectID `bson:"OwnerCommunities"`
	}

	err := Users.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// Validar que el usuario tenga comunidades
	if len(user.OwnerCommunities) == 0 {
		return nil, errors.New("el usuario no tiene comunidades")
	}

	// Paginación
	limit := int64(10)
	skip := int64((page - 1) * 10)

	// Filtro para anuncios aceptados
	filter := bson.M{
		"State": "accepted",
		"CommunityId": bson.M{
			"$in": user.OwnerCommunities,
		},
	}

	// Opciones de consulta: paginación
	options := options.Find()
	options.SetLimit(limit)
	options.SetSkip(skip)

	// Ejecutar la consulta en Advertisements
	cursor, err := Advertisements.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decodificar los resultados
	var acceptedAds []advertisements.Advertisements
	for cursor.Next(ctx) {
		var ad advertisements.Advertisements
		if err := cursor.Decode(&ad); err != nil {
			return nil, err
		}
		acceptedAds = append(acceptedAds, ad)
	}

	// Verificar errores del cursor
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return acceptedAds, nil
}
func (r *UserRepository) GetActiveAdsByEndAdCommunity(page int, userId primitive.ObjectID) ([]advertisements.Advertisements, error) {
	// Base de datos y colecciones
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Users := db.Collection("Users")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	// Buscar el usuario por userId
	var user struct {
		OwnerCommunities []primitive.ObjectID `bson:"OwnerCommunities"`
	}

	err := Users.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// Validar que el usuario tenga comunidades
	if len(user.OwnerCommunities) == 0 {
		return nil, errors.New("el usuario no tiene comunidades")
	}

	// Paginación
	limit := int64(10)
	skip := int64((page - 1) * 10)

	// Filtro para anuncios con EndAdCommunity superior a la fecha actual
	currentTime := time.Now()
	filter := bson.M{
		"State": "accepted",
		"CommunityId": bson.M{
			"$in": user.OwnerCommunities,
		},
		"EndAdCommunity": bson.M{
			"$gt": currentTime, // Activos si la fecha es mayor a la actual
		},
	}

	// Opciones de consulta: paginación
	options := options.Find()
	options.SetLimit(limit)
	options.SetSkip(skip)

	// Ejecutar la consulta en Advertisements
	cursor, err := Advertisements.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decodificar los resultados
	var activeAds []advertisements.Advertisements
	for cursor.Next(ctx) {
		var ad advertisements.Advertisements
		if err := cursor.Decode(&ad); err != nil {
			return nil, err
		}
		activeAds = append(activeAds, ad)
	}

	// Verificar errores del cursor
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return activeAds, nil
}
func (r *UserRepository) GetAdsByNameUser(page int, userId primitive.ObjectID, nameUser string) ([]advertisements.Advertisements, error) {
	// Base de datos y colecciones
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Users := db.Collection("Users")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	// Buscar el usuario por userId
	var user struct {
		OwnerCommunities []primitive.ObjectID `bson:"OwnerCommunities"`
	}

	err := Users.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// Validar que el usuario tenga comunidades
	if len(user.OwnerCommunities) == 0 {
		return nil, errors.New("el usuario no tiene comunidades")
	}

	// Paginación
	limit := int64(10)
	skip := int64((page - 1) * 10)

	// Filtro para anuncios por NameUser
	filter := bson.M{
		"NameUser": nameUser,
		"CommunityId": bson.M{
			"$in": user.OwnerCommunities,
		},
	}

	// Opciones de consulta: paginación
	options := options.Find()
	options.SetLimit(limit)
	options.SetSkip(skip)

	// Ejecutar la consulta en Advertisements
	cursor, err := Advertisements.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decodificar los resultados
	var adsByName []advertisements.Advertisements
	for cursor.Next(ctx) {
		var ad advertisements.Advertisements
		if err := cursor.Decode(&ad); err != nil {
			return nil, err
		}
		adsByName = append(adsByName, ad)
	}

	// Verificar errores del cursor
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return adsByName, nil
}
func (r *UserRepository) updatePinkkerProfitPerMonthAAdCommunity(ctx context.Context, CostToCreateCommunity float64) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollMonthly := GoMongoDB.Collection("PinkkerProfitPerMonth")

	currentTime := time.Now()
	currentMonth := int(currentTime.Month())
	currentYear := currentTime.Year()

	// Usamos la función para obtener el día en formato "YYYY-MM-DD"
	currentDay := helpers.GetDayOfMonth(currentTime)

	startOfMonth := time.Date(currentYear, time.Month(currentMonth), 1, 0, 0, 0, 0, time.UTC)
	startOfNextMonth := time.Date(currentYear, time.Month(currentMonth+1), 1, 0, 0, 0, 0, time.UTC)

	// Filtro para el documento del mes actual
	monthlyFilter := bson.M{
		"timestamp": bson.M{
			"$gte": startOfMonth,
			"$lt":  startOfNextMonth,
		},
	}

	// Paso 1: Inserta el documento si no existe con la estructura básica para el día actual
	_, err := GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, bson.M{
		"$setOnInsert": bson.M{
			"timestamp":          currentTime,
			"days." + currentDay: PinkkerProfitPerMonthdomain.NewDefaultDay(),
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	// Actualización diaria del valor total y el campo CommissionsCommunity para el día actual
	monthlyUpdate := bson.M{
		"$inc": bson.M{
			"total": CostToCreateCommunity,
			"days." + currentDay + ".CommissionsCommunity": CostToCreateCommunity,
		},
	}

	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) SaveNotification(userID primitive.ObjectID, notification notificationsdomain.Notification) error {
	// Colección de notificaciones
	ctx := context.Background()
	db := r.mongoClient.Database("PINKKER-BACKEND")

	notificationsCollection := db.Collection("Notifications")

	// Asegurar que la notificación tenga una marca de tiempo
	if notification.Timestamp.IsZero() {
		notification.Timestamp = time.Now()
	}

	// Filtro para buscar el documento del usuario
	filter := bson.M{"userId": userID}

	// Actualización para agregar la notificación y crear el documento si no existe
	update := bson.M{
		"$push":        bson.M{"notifications": notification}, // Agrega la notificación al array
		"$setOnInsert": bson.M{"userId": userID},              // Crea el documento si no existe
	}

	// Realizar la operación con upsert
	_, err := notificationsCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("error al guardar la notificación: %v", err)
	}

	return nil
}
func (r *UserRepository) UpdatePinkkerProfitPerMonthRegisterLinkReferent(source string) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollMonthly := GoMongoDB.Collection("PinkkerProfitPerMonth")
	ctx := context.Background()
	currentTime := time.Now()
	currentMonth := int(currentTime.Month())
	currentYear := currentTime.Year()
	currentDay := helpers.GetDayOfMonth(currentTime) // Por ejemplo, "15"
	startOfMonth := time.Date(currentYear, time.Month(currentMonth), 1, 0, 0, 0, 0, time.UTC)
	startOfNextMonth := time.Date(currentYear, time.Month(currentMonth+1), 1, 0, 0, 0, 0, time.UTC)

	monthlyFilter := bson.M{
		"timestamp": bson.M{
			"$gte": startOfMonth,
			"$lt":  startOfNextMonth,
		},
	}

	// Paso 1: Inserta el documento si no existe, inicializando valores básicos
	_, err := GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, bson.M{
		"$setOnInsert": bson.M{
			"timestamp":          currentTime,
			"days." + currentDay: PinkkerProfitPerMonthdomain.NewDefaultDay(),
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	// Paso 2: Inicializa 'days.currentDay.userRegistrations' si no existe
	monthlyUpdateEnsureDay := bson.M{
		"$setOnInsert": bson.M{
			"days." + currentDay + ".UserRegistrations": bson.M{}, // Solo lo inicializa si no existe
		},
	}

	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdateEnsureDay)
	if err != nil {
		return err
	}

	// Paso 3: Incrementa el conteo de registros por fuente
	monthlyUpdate := bson.M{
		"$inc": bson.M{
			"days." + currentDay + ".UserRegistrations." + source: 1, // Incrementa el registro por la fuente dada
		},
	}
	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdate)
	if err != nil {
		return err
	}

	return nil
}
