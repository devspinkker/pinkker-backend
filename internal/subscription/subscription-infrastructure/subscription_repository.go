package subscriptioninfrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
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

func (r *SubscriptionRepository) Subscription(Source, Destination primitive.ObjectID, text string) error {
	usersCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	ctx, cancel := context.WithTimeout(context.Background(), 30*24*time.Hour)
	defer cancel()

	sourceUser, destUser, err := r.findUsersBy_ids(ctx, Source, Destination, usersCollection)
	if err != nil {
		return err
	}
	if sourceUser.ID == destUser.ID {
		return errors.New("You can't subscribe to yourself")
	}
	if sourceUser.Pixeles < 1000 {
		return errors.New("pixeles insufficient")
	}

	// Verificar si el usuario fuente ya tiene una suscripción existente al usuario destino
	existingSubscription, err := r.getSubscriptionByUserIDs(sourceUser.ID, destUser.ID)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return err
		}
	}

	subscriptionStart := time.Now()
	subscriptionEnd := subscriptionStart.Add(30 * 24 * time.Hour)

	var subscriptionID primitive.ObjectID
	if existingSubscription.ID == primitive.NilObjectID {
		subscriptionID, err = r.addSubscription(sourceUser, destUser, subscriptionStart, subscriptionEnd, text)
		if err != nil {
			return err
		}
		err = r.addSubscriber(destUser, sourceUser, subscriptionEnd, text)
		if err != nil {
			return err
		}

	} else {
		err = r.updateSubscription(existingSubscription.ID, subscriptionStart, subscriptionEnd, text)
		if err != nil {
			return err
		}
		Subscriber, err := r.getSubscribersByUserIDs(sourceUser.ID, destUser.ID)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return err
			}
		}
		err = r.updateSubscriber(Subscriber.ID.Hex(), subscriptionEnd, text)
		if err != nil {
			return err
		}
		subscriptionID = existingSubscription.ID
	}

	if err := r.updateUserSource(ctx, sourceUser, usersCollection, Destination, subscriptionID); err != nil {
		return err
	}

	err = r.updateUserDest(ctx, destUser, usersCollection)
	return err
}

func (r *SubscriptionRepository) findUsersBy_ids(ctx context.Context, source_id, dest_id primitive.ObjectID, usersCollection *mongo.Collection) (*userdomain.User, *userdomain.User, error) {
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

func (r *SubscriptionRepository) addSubscription(sourceUser *userdomain.User, destUser *userdomain.User, subscriptionStart, subscriptionEnd time.Time, text string) (primitive.ObjectID, error) {
	ctx := context.TODO()
	usersCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	existingSubscription, _ := r.getSubscriptionByUserIDs(sourceUser.ID, destUser.ID)

	if existingSubscription.ID != primitive.NilObjectID && existingSubscription.SubscriptionEnd.After(time.Now()) {
		existingSubscription.MonthsSubscribed += 1
		existingSubscription.SubscriptionStart = subscriptionStart
		existingSubscription.SubscriptionEnd = subscriptionEnd
		existingSubscription.Text = text
		existingSubscription.Notified = false

		sourceUser.Pixeles -= 1000
		destUser.Pixeles += 1000

		err := r.updateSubscription(existingSubscription.ID, subscriptionStart, subscriptionEnd, text)
		if err != nil {
			return primitive.NilObjectID, err
		}

		destUser.Subscribers = append(destUser.Subscribers, existingSubscription.ID)
		err = r.updateUserDest(ctx, destUser, usersCollection)
		if err != nil {
			return primitive.NilObjectID, err
		}

		return existingSubscription.ID, nil
	}

	if sourceUser.Pixeles >= 1000 {
		sourceUser.Pixeles -= 1000
		destUser.Pixeles += 1000

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
		}

		subscriptionID, err := r.insertSubscription(subscription)
		if err != nil {
			return primitive.NilObjectID, err
		}

		sourceUser.Subscriptions = append(sourceUser.Subscriptions, subscriptionID)

		err = r.updateUserDest(ctx, destUser, usersCollection)
		if err != nil {
			return primitive.NilObjectID, err
		}

		return subscriptionID, nil
	}

	return primitive.NilObjectID, errors.New("pixeles insufficient")
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

func (r *SubscriptionRepository) updateUserSource(ctx context.Context, user *userdomain.User, usersCollection *mongo.Collection, Destination primitive.ObjectID, subscriptionID primitive.ObjectID) error {
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

func (r *SubscriptionRepository) updateUserDest(ctx context.Context, user *userdomain.User, usersCollection *mongo.Collection) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{
			"Subscribers": user.Subscribers,
			"Pixeles":     user.Pixeles,
		},
	}

	_, err := usersCollection.UpdateOne(ctx, filter, update)
	return err
}

func (r *SubscriptionRepository) addSubscriber(destUser *userdomain.User, sourceUser *userdomain.User, subscriptionEnd time.Time, text string) error {

	existingSubscriberID := ""
	for _, subscriberID := range destUser.Subscribers {
		subscriber := r.getSubscriberByID(subscriberID)

		if subscriber.SubscriptionNameUser == sourceUser.NameUser {
			existingSubscriberID = subscriberID.Hex()
			break
		}
	}
	if existingSubscriberID != "" {
		// Actualizar la suscripción existente
		err := r.updateSubscriber(existingSubscriberID, subscriptionEnd, text)
		if err != nil {
			return err
		}
		destUser.Pixeles += 1000
		return nil
	}

	// Si no hay una suscripción existente, agregar una nueva
	subscriber := subscriptiondomain.Subscription{
		SubscriptionNameUser: sourceUser.NameUser,
		SourceUserID:         sourceUser.ID,
		DestinationUserID:    destUser.ID,
		SubscriptionStart:    time.Now(),
		SubscriptionEnd:      subscriptionEnd,
		Notified:             false,
		MonthsSubscribed:     0,
		Text:                 text,
	}

	subscriberID, err := r.insertSubscriber(subscriber)
	if err != nil {
		return err
	}

	destUser.Pixeles += 1000
	destUser.Subscribers = append(destUser.Subscribers, subscriberID)

	return nil
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
			"Text":            text,
		},
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
	return err
}

func (r *SubscriptionRepository) insertSubscriber(subscriber subscriptiondomain.Subscription) (primitive.ObjectID, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")

	result, err := collection.InsertOne(context.Background(), subscriber)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *SubscriptionRepository) getSubscriberByID(subscriberID primitive.ObjectID) subscriptiondomain.Subscription {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")

	var subscriber subscriptiondomain.Subscription
	collection.FindOne(context.Background(), bson.M{"_id": subscriberID}).Decode(&subscriber)

	return subscriber
}
func (r *SubscriptionRepository) GetSubsChat(id primitive.ObjectID) ([]subscriptiondomain.ResSubscriber, error) {
	GoMongoDBCollDonations := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscribers")
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "destinationUserID", Value: id}}}},
		{{Key: "$sort", Value: bson.D{{Key: "SubscriptionEnd", Value: -1}}}},
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

func (u *SubscriptionRepository) DeleteRedisUserChatInOneRoom(userToDelete primitive.ObjectID, IdRoom primitive.ObjectID) error {
	GoMongoDBColl := u.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreams := GoMongoDBColl.Collection("Streams")
	filter := bson.M{"StreamerID": IdRoom}
	var stream streamdomain.Stream
	err := GoMongoDBCollStreams.FindOne(context.Background(), filter).Decode(&stream)
	if err != nil {
		return err
	}
	userHashKey := "userInformation:" + userToDelete.Hex() + ":inTheRoom:" + stream.ID.Hex()
	_, err = u.redisClient.Del(context.Background(), userHashKey).Result()
	if err != nil {
		return err
	}
	return err
}
