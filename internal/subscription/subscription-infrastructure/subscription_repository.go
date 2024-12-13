package subscriptioninfrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"PINKKER-BACKEND/config"
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	notificationsdomain "PINKKER-BACKEND/internal/notifications/notifications"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"PINKKER-BACKEND/pkg/utils"
)

type SubscriptionRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewSubscriptionRepository(redisClient *redis.Client, mongoClient *mongo.Client) *SubscriptionRepository {
	return &SubscriptionRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (r *SubscriptionRepository) GetTOTPSecret(ctx context.Context, userID primitive.ObjectID) (string, error) {
	usersCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var result struct {
		TOTPSecret string `bson:"TOTPSecret"`
	}

	err := usersCollection.FindOne(
		ctx,
		bson.M{"_id": userID},
		options.FindOne().SetProjection(bson.M{"TOTPSecret": 1}),
	).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.TOTPSecret, nil

}

func (r *SubscriptionRepository) GetWebSocketClientsInRoom(roomID string) ([]*websocket.Conn, error) {
	clients, err := utils.NewChatService().GetWebSocketClientsInRoom(roomID)

	return clients, err
}
func (r *SubscriptionRepository) Subscription(Source, Destination primitive.ObjectID, text string) (string, string, error) {
	usersCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	ctx, cancel := context.WithTimeout(context.Background(), 30*24*time.Hour)
	defer cancel()

	sourceUser, destUser, err := r.findUsersBy_ids(ctx, Source, Destination, usersCollection)
	if err != nil {
		return sourceUser.NameUser, sourceUser.Avatar, err
	}
	if sourceUser.ID == destUser.ID {
		return sourceUser.NameUser, sourceUser.Avatar, errors.New("you can't subscribe to yourself")
	}

	SubsPayPerPrint := config.SubsPayPerPrint()
	moneySubs, err := strconv.Atoi(SubsPayPerPrint)
	if err != nil {
		fmt.Printf("el valor SubsPayPerPrint no es un número válido, usando 1000 como valor predeterminado")
		moneySubs = 1000
	}
	percent := (moneySubs * 2) / 100
	totalSub := float64(moneySubs + percent)
	if sourceUser.Pixeles < totalSub {
		return sourceUser.NameUser, sourceUser.Avatar, errors.New("pixeles insufficient")
	}
	existingSubscription, err := r.getSubscriptionByUserIDs(sourceUser.ID, destUser.ID)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return sourceUser.NameUser, sourceUser.Avatar, err
		}
	}

	subscriptionStart := time.Now()
	subscriptionEnd := subscriptionStart.Add(30 * 24 * time.Hour)

	var subscriptionID primitive.ObjectID
	var subscriberID primitive.ObjectID

	if existingSubscription.ID == primitive.NilObjectID {
		subscriptionID, err = r.addSubscription(sourceUser, destUser, subscriptionStart, subscriptionEnd, text)
		if err != nil {
			return sourceUser.NameUser, sourceUser.Avatar, err
		}
		subscriberID, err = r.addSubscriber(destUser, sourceUser, subscriptionEnd, text)
		if err != nil {
			return sourceUser.NameUser, sourceUser.Avatar, err
		}
	} else {
		err = r.updateSubscription(existingSubscription.ID, subscriptionStart, subscriptionEnd, text)
		if err != nil {
			return sourceUser.NameUser, sourceUser.Avatar, err
		}
		subscriberIDgetSubscriber, err := r.getSubscribersByUserIDs(sourceUser.ID, destUser.ID)
		subscriberID = subscriberIDgetSubscriber.ID

		if err != nil {
			if err != mongo.ErrNoDocuments {
				return sourceUser.NameUser, sourceUser.Avatar, err
			}
		}
		err = r.updateSubscriber(subscriberIDgetSubscriber.ID.Hex(), subscriptionEnd, text)
		if err != nil {
			return sourceUser.NameUser, sourceUser.Avatar, err
		}
		subscriptionID = existingSubscription.ID
	}

	if err := r.updateUserSource(ctx, sourceUser, usersCollection, Destination, subscriptionID, totalSub); err != nil {
		return sourceUser.NameUser, sourceUser.Avatar, err
	}
	err = r.updateUserDest(ctx, destUser, usersCollection, subscriberID, moneySubs)
	if err != nil {
		return sourceUser.NameUser, sourceUser.Avatar, err
	}
	CommissionsSub := totalSub * 0.02
	err = r.updatePinkkerProfitPerMonth(ctx, CommissionsSub)
	return sourceUser.NameUser, sourceUser.Avatar, err
}

func (r *SubscriptionRepository) findUsersBy_ids(ctx context.Context, source_id, dest_id primitive.ObjectID, usersCollection *mongo.Collection) (*userdomain.User, *userdomain.User, error) {
	var sourceUser struct {
		ID       primitive.ObjectID `bson:"_id"`
		NameUser string             `bson:"NameUser"`
		Avatar   string             `bson:"Avatar"`
		Pixeles  float64            `bson:"Pixeles"`
	}
	var destUser struct {
		ID       primitive.ObjectID `bson:"_id"`
		NameUser string             `bson:"NameUser"`
	}

	// Fetch source user with necessary fields
	err := usersCollection.FindOne(
		ctx,
		bson.M{"_id": source_id},
		options.FindOne().SetProjection(bson.M{"_id": 1, "NameUser": 1, "Pixeles": 1, "Avatar": 1}),
	).Decode(&sourceUser)
	if err != nil {
		return nil, nil, err
	}

	// Fetch destination user with necessary fields
	err = usersCollection.FindOne(
		ctx,
		bson.M{"_id": dest_id},
		options.FindOne().SetProjection(bson.M{"_id": 1, "NameUser": 1}),
	).Decode(&destUser)
	if err != nil {
		return nil, nil, err
	}

	// Convert simplified structs to full user domain objects
	sourceUserDomain := &userdomain.User{
		ID:       sourceUser.ID,
		NameUser: sourceUser.NameUser,
		Pixeles:  sourceUser.Pixeles,
		Avatar:   sourceUser.Avatar,
	}
	destUserDomain := &userdomain.User{
		ID:       destUser.ID,
		NameUser: destUser.NameUser,
	}

	return sourceUserDomain, destUserDomain, nil
}

func (r *SubscriptionRepository) addSubscription(sourceUser *userdomain.User, destUser *userdomain.User, subscriptionStart, subscriptionEnd time.Time, text string) (primitive.ObjectID, error) {

	var monthsSubscribed int
	if subscriptionEnd.After(time.Now()) {
		monthsSubscribed = 0
	}

	subscription := subscriptiondomain.Subscription{
		SubscriptionNameUser: destUser.NameUser,
		SourceUserID:         sourceUser.ID,
		DestinationUserID:    destUser.ID,
		SubscriptionStart:    subscriptionStart,
		SubscriptionEnd:      subscriptionEnd,
		MonthsSubscribed:     monthsSubscribed,
		Notified:             false,
		Text:                 text,
		TimeStamp:            time.Now(),
	}

	subscriptionID, err := r.insertSubscription(subscription)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return subscriptionID, nil

}

func (r *SubscriptionRepository) insertSubscription(subscription subscriptiondomain.Subscription) (primitive.ObjectID, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscriptions")

	result, err := collection.InsertOne(context.Background(), subscription)
	if err != nil {
		return primitive.NilObjectID, err
	}
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, errors.New("could not convert InsertedID to ObjectID")
	}

	return insertedID, nil
}

func (r *SubscriptionRepository) getSubscriptionByUserIDs(sourceUserID, destUserID primitive.ObjectID) (subscriptiondomain.Subscription, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscriptions")

	var subscription subscriptiondomain.Subscription
	filter := bson.M{
		"sourceUserID":      sourceUserID,
		"destinationUserID": destUserID,
	}

	err := collection.FindOne(context.Background(), filter).Decode(&subscription)
	if err != nil {
		return subscription, err
	}

	return subscription, nil
}
func (r *SubscriptionRepository) getSubscribersByUserIDs(sourceUserID, destUserID primitive.ObjectID) (subscriptiondomain.Subscriber, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")

	var subscription subscriptiondomain.Subscriber
	filter := bson.M{
		"sourceUserID":      sourceUserID,
		"destinationUserID": destUserID,
	}

	err := collection.FindOne(context.Background(), filter).Decode(&subscription)
	if err != nil {
		return subscription, err
	}

	return subscription, nil
}

func (r *SubscriptionRepository) updateSubscription(subscriptionID primitive.ObjectID, subscriptionStart, subscriptionEnd time.Time, text string) error {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscriptions")

	update := bson.M{
		"$set": bson.M{
			"SubscriptionStart": subscriptionStart,
			"SubscriptionEnd":   subscriptionEnd,
			"Text":              text,
		},
	}

	_, err := collection.UpdateOne(context.Background(), bson.M{"_id": subscriptionID}, update)
	return err
}

func (r *SubscriptionRepository) updateUserSource(ctx context.Context, user *userdomain.User, usersCollection *mongo.Collection, Destination primitive.ObjectID, subscriptionID primitive.ObjectID, totalSub float64) error {
	filter := bson.M{"_id": user.ID}

	update := bson.M{
		"$addToSet": bson.M{
			"Subscriptions": subscriptionID,
		},
		"$inc": bson.M{
			"Pixeles": -totalSub,
		},
	}
	_, err := usersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	database := r.mongoClient.Database("PINKKER-BACKEND")

	streamFilter := bson.M{"StreamerID": Destination}
	streamSub := streamdomain.Stream{}
	err = database.Collection("Streams").FindOne(ctx, streamFilter).Decode(&streamSub)
	if err != nil {
		return err
	}

	updateRoom := bson.M{
		"$set": bson.M{
			"Rooms.$.Subscription":  subscriptionID,
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

func (r *SubscriptionRepository) updateUserDest(ctx context.Context, user *userdomain.User, usersCollection *mongo.Collection, subscriptionID primitive.ObjectID, moneySubs int) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{

		"$addToSet": bson.M{
			"Subscribers": subscriptionID,
		},
		"$inc": bson.M{
			"Pixeles": moneySubs,
		},
	}

	_, err := usersCollection.UpdateOne(ctx, filter, update)
	return err
}

func (r *SubscriptionRepository) addSubscriber(destUser *userdomain.User, sourceUser *userdomain.User, subscriptionEnd time.Time, text string) (primitive.ObjectID, error) {

	subscriber := subscriptiondomain.Subscriber{
		SubscriberNameUser: sourceUser.NameUser,
		SourceUserID:       sourceUser.ID,
		DestinationUserID:  destUser.ID,
		SubscriptionStart:  time.Now(),
		SubscriptionEnd:    subscriptionEnd,
		Notified:           false,
		Text:               text,
		TimeStamp:          time.Now(),
	}

	subscriberID, err := r.insertSubscriber(subscriber)
	if err != nil {
		return subscriberID, err
	}

	return subscriberID, nil
}
func (r *SubscriptionRepository) updateSubscriber(subscriberID string, subscriptionEnd time.Time, text string) error {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")

	objectID, err := primitive.ObjectIDFromHex(subscriberID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"SubscriptionEnd": subscriptionEnd,
			"TimeStamp":       time.Now(),
			"Text":            text,
		},
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	return err
}

func (r *SubscriptionRepository) insertSubscriber(subscriber subscriptiondomain.Subscriber) (primitive.ObjectID, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")

	result, err := collection.InsertOne(context.Background(), subscriber)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

// func (r *SubscriptionRepository) getSubscriberByID(subscriberID primitive.ObjectID) subscriptiondomain.Subscription {
// 	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")

// 	var subscriber subscriptiondomain.Subscription
// 	collection.FindOne(context.Background(), bson.M{"_id": subscriberID}).Decode(&subscriber)

//		return subscriber
//	}
func (r *SubscriptionRepository) GetSubsChat(id primitive.ObjectID) ([]subscriptiondomain.ResSubscriber, error) {
	GoMongoDBCollDonations := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "destinationUserID", Value: id}}}},
		{{Key: "$sort", Value: bson.D{{Key: "SubscriptionStart", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "sourceUserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "FromUserInfo"},
		}}},
		{{Key: "$unwind", Value: "$FromUserInfo"}},
		// Proyecta los campos necesarios
		{{Key: "$project", Value: bson.D{
			{Key: "SubscriberNameUser", Value: 1},
			{Key: "FromUserInfo.Avatar", Value: 1},
			{Key: "FromUserInfo.NameUser", Value: 1},
			{Key: "SubscriptionStart", Value: 1},
			{Key: "SubscriptionEnd", Value: 1},
			{Key: "Notified", Value: 1},
			{Key: "Text", Value: 1},
			{Key: "id", Value: "$_id"},
		}}},
	}

	cursor, err := GoMongoDBCollDonations.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var Subscriber []subscriptiondomain.ResSubscriber

	for cursor.Next(context.Background()) {
		var donation subscriptiondomain.ResSubscriber
		if err := cursor.Decode(&donation); err != nil {
			return nil, err
		}
		Subscriber = append(Subscriber, donation)
	}
	if len(Subscriber) == 0 {
		return nil, errors.New("no documents found")

	}
	return Subscriber, nil
}
func (r *SubscriptionRepository) UpdataSubsChat(id, ToUserID primitive.ObjectID) error {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	var subscriber streamdomain.Stream
	collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&subscriber)
	var ToUser streamdomain.Stream
	collection.FindOne(context.Background(), bson.M{"_id": ToUserID}).Decode(&ToUser)

	userHashKey := "userInformation:" + subscriber.Streamer + ":inTheRoom:" + ToUser.ID.Hex()
	err := r.redisClient.Del(context.Background(), userHashKey)

	return err.Err()
}

func (r *SubscriptionRepository) GetSubsAct(Source, Destination primitive.ObjectID) (subscriptiondomain.Subscription, error) {
	subs, err := r.getSubscriptionByUserIDs(Source, Destination)
	return subs, err
}

func (u *SubscriptionRepository) DeleteRedisUserChatInOneRoom(userToDelete, IdRoom primitive.ObjectID) error {
	db := u.mongoClient.Database("PINKKER-BACKEND")
	collectionStreams := db.Collection("Streams")
	filter := bson.M{"StreamerID": IdRoom}
	var stream streamdomain.Stream

	// Obtener el stream usando el ID de la habitación
	err := collectionStreams.FindOne(context.Background(), filter).Decode(&stream)
	if err != nil {
		return err
	}

	collectionUsers := db.Collection("Users")

	userFilter := bson.M{"_id": userToDelete}
	var userFolloer struct {
		NameUser string `bson:"NameUser"`
	}

	err = collectionUsers.FindOne(
		context.Background(),
		userFilter,
		options.FindOne().SetProjection(bson.M{"NameUser": 1}),
	).Decode(&userFolloer)
	if err != nil {
		return err
	}

	userHashKey := "userInformation:" + userFolloer.NameUser + ":inTheRoom:" + stream.ID.Hex()
	_, err = u.redisClient.Del(context.Background(), userHashKey).Result()
	return err
}

func (u *SubscriptionRepository) GetStreamByStreamerID(user primitive.ObjectID) (streamdomain.Stream, error) {
	GoMongoDBColl := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreams := GoMongoDBColl.Collection("Streams")
	filter := bson.M{"StreamerID": user}
	var stream streamdomain.Stream
	err := GoMongoDBCollStreams.FindOne(context.Background(), filter).Decode(&stream)
	return stream, err
}
func (u *SubscriptionRepository) PublishNotification(roomID string, noty map[string]interface{}) error {

	chatMessageJSON, err := json.Marshal(noty)
	if err != nil {
		return err
	}
	err = u.redisClient.Publish(
		context.Background(),
		roomID+"action",
		string(chatMessageJSON),
	).Err()
	if err != nil {
		return err
	}

	return err
}
func (u *SubscriptionRepository) StateTheUserInChat(Donado primitive.ObjectID, Donante primitive.ObjectID) (bool, error) {
	ctx := context.Background()
	db := u.mongoClient.Database("PINKKER-BACKEND")

	stream, err := u.GetStreamByStreamerID(Donado)
	if err != nil {
		return true, err
	}
	userDonante, err := u.GetUserID(ctx, db, Donante)
	if err != nil {
		return true, err
	}

	infoUser, err := u.GetInfoUserInRoom(userDonante, stream.ID)
	return infoUser.Baneado, err
}
func (r *SubscriptionRepository) GetInfoUserInRoom(nameUser string, getInfoUserInRoom primitive.ObjectID) (*userdomain.UserInfo, error) {
	database := r.mongoClient.Database("PINKKER-BACKEND")
	var room *userdomain.UserInfo
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
	} else {
		return nil, fmt.Errorf("no room found for user %s in room %s", nameUser, getInfoUserInRoom.Hex())
	}

	if err := cursor.Err(); err != nil {
		return room, err
	}

	return room, nil
}

func (u *SubscriptionRepository) GetUserID(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) (string, error) {
	usersCollection := db.Collection("Users")
	filter := bson.M{"_id": userID}
	var user struct {
		NameUser string `bson:"NameUser"`
	}

	err := usersCollection.FindOne(
		ctx,
		filter,
		options.FindOne().SetProjection(bson.M{"NameUser": 1}),
	).Decode(&user)
	if err != nil {
		return "", err
	}
	return user.NameUser, nil
}
func (r *SubscriptionRepository) updatePinkkerProfitPerMonth(ctx context.Context, CommissionsSub float64) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollMonthly := GoMongoDB.Collection("PinkkerProfitPerMonth")

	currentTime := time.Now()

	// Obtenemos el día en formato "YYYY-MM-DD"
	currentDay := helpers.GetDayOfMonth(currentTime)

	// Definimos el mes actual para filtrar los documentos
	currentMonth := int(currentTime.Month())
	currentYear := currentTime.Year()

	startOfMonth := time.Date(currentYear, time.Month(currentMonth), 1, 0, 0, 0, 0, time.UTC)
	startOfNextMonth := time.Date(currentYear, time.Month(currentMonth+1), 1, 0, 0, 0, 0, time.UTC)

	// Filtro para el documento del mes actual
	monthlyFilter := bson.M{
		"timestamp": bson.M{
			"$gte": startOfMonth,
			"$lt":  startOfNextMonth,
		},
	}

	// Inserta el documento si no existe con la estructura básica para el día actual
	_, err := GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, bson.M{
		"$setOnInsert": bson.M{
			"timestamp":          currentTime,
			"days." + currentDay: PinkkerProfitPerMonthdomain.NewDefaultDay(),
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	// Actualización diaria del valor total y el campo CommissionsSuscripcion para el día actual
	monthlyUpdate := bson.M{
		"$inc": bson.M{
			"total": CommissionsSub,
			"days." + currentDay + ".CommissionsSuscripcion": CommissionsSub,
		},
	}

	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (r *SubscriptionRepository) SaveNotification(userID primitive.ObjectID, notification notificationsdomain.Notification) error {
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

func (u *SubscriptionRepository) IsFollowing(IdUserTokenP, followedUserID primitive.ObjectID) (bool, error) {
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
