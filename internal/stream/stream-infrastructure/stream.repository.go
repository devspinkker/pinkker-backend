package streaminfrastructure

import (
	"PINKKER-BACKEND/config"
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-domain"
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StreamRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewStreamRepository(redisClient *redis.Client, mongoClient *mongo.Client) *StreamRepository {
	return &StreamRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (r *StreamRepository) RecommendStreams(limit int, page int) ([]streamdomain.Stream, error) {
	ctx := context.Background()

	// Obtener la colección Stream
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreams := GoMongoDB.Collection("Streams")

	skip := (page - 1) * limit

	pipeline := bson.A{
		bson.M{"$sort": bson.M{"RecommendationScore": -1}},
		bson.M{"$skip": skip},
		bson.M{"$limit": limit},
	}
	cursor, err := GoMongoDBCollStreams.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var recommendedStreams []streamdomain.Stream
	if err = cursor.All(ctx, &recommendedStreams); err != nil {
		return nil, err
	}

	return recommendedStreams, nil
}

func (r *StreamRepository) UpdateOnline(Key string, state bool) (primitive.ObjectID, error) {

	LastStreamSummary := primitive.ObjectID{}

	session, err := r.mongoClient.StartSession()
	if err != nil {
		return LastStreamSummary, err
	}
	defer session.EndSession(context.Background())

	err = session.StartTransaction()
	if err != nil {
		return LastStreamSummary, err
	}

	ctx := mongo.NewSessionContext(context.Background(), session)

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")
	GoMongoDBCollStreams := GoMongoDB.Collection("Streams")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")
	filterUsers := bson.D{
		{Key: "KeyTransmission", Value: Key},
	}

	userFind, err := r.getUser(filterUsers)
	if err != nil {
		session.AbortTransaction(ctx)
		return LastStreamSummary, err
	}
	filterStreams := bson.D{
		{Key: "StreamerID", Value: userFind.ID},
	}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	updateStreams := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Online", Value: state},
			{Key: "StartDate", Value: time.Now()},
			{Key: "RecommendationScore", Value: 0},
		}},
	}
	var StreamFind streamdomain.Stream

	err = GoMongoDBCollStreams.FindOneAndUpdate(ctx, filterStreams, updateStreams, options).Decode(&StreamFind)
	if err != nil {
		session.AbortTransaction(ctx)
		return LastStreamSummary, err
	}

	filterUsers = bson.D{
		{Key: "_id", Value: userFind.ID},
	}

	updateUsers := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Online", Value: state},
		}},
	}

	_, err = GoMongoDBCollUsers.UpdateOne(ctx, filterUsers, updateUsers)
	if err != nil {
		session.AbortTransaction(ctx)
		return LastStreamSummary, err
	}
	if state {
		// notifyOnlineStreamer := []string{}
		// for _, followInfo := range userFind.Followers {
		// 	if followInfo.Notifications {
		// 		notifyOnlineStreamer = append(notifyOnlineStreamer, followInfo.Email)
		// 	}
		// }
		// _ = helpers.ResendNotificationStreamerOnline(userFind.NameUser, notifyOnlineStreamer)
		err = r.publishNotification("stream_on", userFind.NameUser, userFind.ID.Hex(), StreamFind.StreamTitle, userFind.Avatar)
		if err != nil {
			fmt.Println(err)
		}
		// modo de chat cuandos se prende
		exist, err := r.redisClient.Exists(context.Background(), StreamFind.ID.Hex()).Result()
		if err != nil {
			return LastStreamSummary, err
		}
		if exist == 1 {
			err := r.redisClient.Set(context.Background(), StreamFind.ID.Hex(), StreamFind.ModChat, 0).Err()
			if err != nil {
				return LastStreamSummary, err
			}
		} else {
			err := r.redisClient.Set(context.Background(), StreamFind.ID.Hex(), StreamFind.ModChat, 0).Err()
			if err != nil {
				return LastStreamSummary, err
			}
		}
		startFollowersCount := userFind.FollowersCount
		startSubsCount := userFind.SubscribersCount

		// aqui quiero crear el resumen del Stream con valores predeterminnados
		summary := StreamSummarydomain.StreamSummary{
			EndOfStream:          time.Now(),
			AverageViewers:       0,
			AverageViewersByTime: make(map[string]int),
			MaxViewers:           0,
			NewFollowers:         0,
			NewSubscriptions:     0,
			Advertisements:       0,
			StartOfStream:        time.Now(),
			StreamerID:           userFind.ID,
			StartFollowersCount:  startFollowersCount,
			StreamDocumentID:     StreamFind.ID,
			StartSubsCount:       startSubsCount,
			Title:                StreamFind.StreamTitle,
			StreamThumbnail:      StreamFind.StreamThumbnail,
			StreamCategory:       StreamFind.StreamCategory,
			Admoney:              0,
			SubscriptionsMoney:   0,
			DonationsMoney:       0,
			TotalMoney:           0,
			Online:               true,
			RecommendationScore:  0,
			Available:            true,
		}

		StreamSummaryRes, err := GoMongoDBCollStreamSummary.InsertOne(ctx, summary)
		StreamSummaryResId := StreamSummaryRes.InsertedID
		streamSummaryID, ok := StreamSummaryResId.(primitive.ObjectID)
		cacheKey := "stream:" + userFind.NameUser
		r.RedisDeleteKey(cacheKey)
		if !ok {
			return primitive.NilObjectID, fmt.Errorf("failed to convert InsertedID to ObjectID")
		}

		if err != nil {
			return streamSummaryID, err
		}
		LastStreamSummary = streamSummaryID
	} else {

		err := r.RedisCacheDeleteRoomMessagesAndUserInfo(StreamFind.ID, StreamFind.Streamer)
		if err != nil {
			fmt.Println("RedisCacheDeleteRoomMessagesAndUserInfo")

			fmt.Println(err)
		}
		_, err = r.redisClient.Del(context.Background(), StreamFind.ID.Hex()).Result()
		if err != nil {
			fmt.Println(err)
		}

		latestSummary, err := r.FindLatestStreamSummaryByStreamerID(userFind.ID)
		if err != nil {
			session.AbortTransaction(ctx)
			return LastStreamSummary, err
		}
		newFollowersCount := userFind.FollowersCount - latestSummary.StartFollowersCount
		newSubsCount := userFind.SubscribersCount - latestSummary.StartSubsCount
		AverageViewers := 0
		maxViewers := 0
		totalCount := 0

		for _, viewers := range latestSummary.AverageViewersByTime {
			AverageViewers += viewers
			totalCount++
			if viewers > maxViewers {
				maxViewers = viewers
			}
		}

		if totalCount > 0 {
			AverageViewers = AverageViewers / totalCount
		}

		// AverageAdPaymentInStreams, err := r.AverageAdPaymentInStreams(ctx, latestSummary.Advertisements)
		// if err != nil {
		// 	return LastStreamSummary, err
		// }
		AdvertisementsPayPerPrint := config.AdvertisementsPayPerPrint()
		intValue, err := strconv.Atoi(AdvertisementsPayPerPrint)
		if err != nil {
			fmt.Printf("el valor  no es un número válido, usando 10 como valor predeterminado")
			intValue = 10
		}
		AverageAdPaymentInStreams := latestSummary.Advertisements * intValue

		err = r.PayUserForStreamsAd(ctx, AverageAdPaymentInStreams, userFind.ID, GoMongoDBCollUsers)

		if err != nil {
			return LastStreamSummary, err
		}
		SubsPayPerPrint := config.SubsPayPerPrint()
		moneySubs, err := strconv.Atoi(SubsPayPerPrint)
		if err != nil {
			fmt.Printf("el valor SubsPayPerPrint no es un número válido, usando 1000 como valor predeterminado")
			intValue = 1000
		}
		TotalSubsSummaryMoney := newSubsCount * moneySubs
		incTotalMoney := AverageAdPaymentInStreams + TotalSubsSummaryMoney
		updateSummary := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "EndOfStream", Value: time.Now()},
				{Key: "NewFollowers", Value: newFollowersCount},
				{Key: "NewSubscriptions", Value: newSubsCount},
				{Key: "AverageViewers", Value: AverageViewers},
				{Key: "MaxViewers", Value: maxViewers},
				{Key: "Title", Value: StreamFind.StreamTitle},
				{Key: "StreamThumbnail", Value: StreamFind.StreamThumbnail},
				{Key: "StreamCategory", Value: StreamFind.StreamCategory},
				{Key: "Admoney", Value: AverageAdPaymentInStreams},
				{Key: "SubscriptionsMoney", Value: TotalSubsSummaryMoney},
				{Key: "Online", Value: false},
			}},
			{Key: "$inc", Value: bson.D{
				{Key: "TotalMoney", Value: incTotalMoney},
			}},
		}

		_, err = GoMongoDBCollStreamSummary.UpdateOne(ctx, bson.M{"_id": latestSummary.ID}, updateSummary)
		if err != nil {
			return LastStreamSummary, err
		}
		LastStreamSummary = latestSummary.ID
		streamDuration := time.Since(latestSummary.StartOfStream).Hours()
		totalTimeOnline := StreamFind.TotalTimeOnline
		totalTimeOnline += streamDuration

		updateStream := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "TotalTimeOnline", Value: totalTimeOnline},
			}},
		}

		// Actualizar el documento de Stream
		_, err = GoMongoDBCollStreams.UpdateOne(ctx, bson.M{"_id": StreamFind.ID}, updateStream)
		if err != nil {
			return LastStreamSummary, err
		}
		notification := map[string]interface{}{
			"action": "Online",
			"Online": false,
		}
		r.PublishAction(StreamFind.ID.Hex()+"action", notification)

	}

	err = session.CommitTransaction(ctx)
	if err != nil {
		return LastStreamSummary, err
	}

	return LastStreamSummary, nil
}

func (r *StreamRepository) RedisCacheDeleteRoomMessagesAndUserInfo(Room primitive.ObjectID, NameUser string) error {
	// Inicia una transacción en Redis usando Pipeline
	pipeline := r.redisClient.Pipeline()

	// Borra la lista de mensajes en la sala
	pipeline.Del(context.Background(), Room.Hex()+"LastMessages")

	// Borra la información del usuario en la sala
	userHashKey := "userInformation:" + NameUser + ":inTheRoom:" + Room.Hex()
	pipeline.Del(context.Background(), userHashKey)

	// Borra la lista de interacciones únicas en la sala
	uniqueInteractionsKey := Room.Hex() + ":uniqueinteractions"
	pipeline.Del(context.Background(), uniqueInteractionsKey)
	// Borra el historial de mensajesdel vod
	MensajesParaElVod := "MensajesParaElVod:" + Room.Hex()
	pipeline.Del(context.Background(), MensajesParaElVod)
	// elimina el modo del chat
	pipeline.Del(context.Background(), Room.Hex())

	cacheKey := "stream:" + NameUser
	pipeline.Del(context.Background(), cacheKey)

	// Ejecuta la transacción
	_, err := pipeline.Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (r *StreamRepository) getUser(filter bson.D) (*userdomain.GetUser, error) {
	GoMongoDBCollUsers := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "FollowersCount", Value: bson.D{
				{Key: "$size", Value: bson.D{
					{Key: "$ifNull", Value: bson.A{
						bson.D{{Key: "$objectToArray", Value: "$Followers"}},
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
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Followers", Value: 0},
			{Key: "Subscribers", Value: 0},
		}}},
	}

	cursor, err := GoMongoDBCollUsers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var user userdomain.GetUser
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
	}

	return &user, nil
}
func (r *StreamRepository) AverageAdPaymentInStreams(ctx context.Context, Advertisements int) (float64, error) {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollAdvertisements := GoMongoDB.Collection("Advertisements")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "Destination", Value: "Streams"},
			{Key: "Impressions", Value: bson.D{{Key: "$lte", Value: "$ImpressionsMax"}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "averagePayPerPrint", Value: bson.D{{Key: "$avg", Value: "$PayPerPrint"}}},
		}}},
	}

	cursor, err := GoMongoDBCollAdvertisements.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return 0, err
	}

	if len(result) > 0 {
		if average, ok := result[0]["averagePayPerPrint"].(float64); ok {
			total := average * float64(Advertisements)
			percentage := total * 0.03
			rounded := math.Round(percentage*100) / 100
			return rounded, nil
		}
	}

	return 0, nil
}

func (r *StreamRepository) PayUserForStreamsAd(ctx context.Context, averageAdPayment int, idUser primitive.ObjectID, coll *mongo.Collection) error {
	filter := bson.D{{Key: "_id", Value: idUser}}

	update := bson.D{
		{Key: "$inc", Value: bson.D{
			{Key: "Pixeles", Value: averageAdPayment},
		}},
	}

	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *StreamRepository) publishNotification(Type, streamerName, id, Title, Avatar string) error {
	message := map[string]interface{}{
		"Type":     Type,
		"Nameuser": streamerName,
		"ID":       id,
		"Title":    Title,
		"Avatar":   Avatar,
	}

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = r.redisClient.Publish(context.Background(), "pinker_notifications", string(jsonMessage)).Err()
	if err != nil {
		return err
	}
	return nil
}
func (r *StreamRepository) CommercialInStreamSelectAdvertisements(StreamCategory string, ViewerCount int) (advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollAdvertisements := db.Collection("Advertisements")
	ctx := context.TODO()

	// Pipeline para buscar coincidencias específicas
	pipelineMatch := bson.A{
		bson.M{"$match": bson.M{
			"Category":    StreamCategory,
			"Destination": "Streams",
			"State":       "accepted",
			"$expr":       bson.M{"$lte": bson.A{bson.M{"$add": bson.A{"$Impressions", ViewerCount}}, "$ImpressionsMax"}},
		}},
		bson.M{"$sample": bson.M{"size": 1}},
	}

	// Pipeline para obtener cualquier documento aleatorio que cumpla con la condición
	pipelineRandom := bson.A{
		bson.M{"$match": bson.M{
			"Destination": "Streams",
			"State":       "accepted",
			"$expr":       bson.M{"$lte": bson.A{bson.M{"$add": bson.A{"$Impressions", ViewerCount}}, "$ImpressionsMax"}},
		}},
		bson.M{"$sample": bson.M{"size": 1}},
		bson.M{"$project": bson.M{
			"IdOfTheUsersWhoClicked": 0,
		}},
	}

	var advertisement advertisements.Advertisements

	// Buscar coincidencia específica
	cursor, err := GoMongoDBCollAdvertisements.Aggregate(ctx, pipelineMatch)
	if err != nil {
		return advertisements.Advertisements{}, err
	}
	defer cursor.Close(ctx)

	// Decodificar el resultado si se encuentra
	if cursor.Next(ctx) {
		if err := cursor.Decode(&advertisement); err != nil {
			return advertisements.Advertisements{}, err
		}
		return advertisement, nil
	}

	// Si no hay coincidencia específica, obtener cualquier documento aleatorio
	cursor, err = GoMongoDBCollAdvertisements.Aggregate(ctx, pipelineRandom)
	if err != nil {
		return advertisements.Advertisements{}, err
	}
	defer cursor.Close(ctx)

	// Decodificar el resultado si se encuentra
	if cursor.Next(ctx) {
		if err := cursor.Decode(&advertisement); err != nil {
			return advertisements.Advertisements{}, err
		}
		return advertisement, nil
	}

	// Si no se encuentra ningún documento, retornar un error
	return advertisements.Advertisements{}, errors.New("no advertisements found")
}

func (r *StreamRepository) CategoriesUpdate(req streamdomain.CategoriesUpdate, idUser primitive.ObjectID) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	fmt.Println(req.Name) // Imprimir el nombre para verificación
	collection := db.Collection("Categorias")
	collectionUsers := db.Collection("Users")
	var User userdomain.User

	err := collectionUsers.FindOne(context.Background(), bson.M{"_id": idUser}).Decode(&User)
	if err != nil {
		return err
	}

	if User.PanelAdminPinkker.Level != 1 || !User.PanelAdminPinkker.Asset || User.PanelAdminPinkker.Code != req.CodeAdmin {
		return fmt.Errorf("usuario no autorizado")
	}

	filter := bson.M{"Name": req.Name}

	if req.Delete {
		_, err := collection.DeleteOne(context.Background(), filter)
		return err
	}

	// Preparar las operaciones de actualización e inserción
	setUpdate := bson.M{
		"Img":      req.Img,
		"TopColor": req.TopColor,
	}

	setOnInsert := bson.M{
		"Name":       req.Name,
		"Spectators": 0,
		"createdAt":  time.Now(),
	}

	// Verificar si la categoría ya existe
	var existingCategory bson.M
	err = collection.FindOne(context.Background(), filter).Decode(&existingCategory)

	if err == mongo.ErrNoDocuments {
		// Si la categoría no existe, la insertamos
		update := bson.M{
			"$set":         setUpdate,
			"$setOnInsert": setOnInsert,
		}

		opts := options.Update().SetUpsert(true)
		_, err = collection.UpdateOne(context.Background(), filter, update, opts)
		if err != nil {
			return err
		}
	} else if err == nil {
		// Si la categoría existe, actualizamos los campos necesarios
		update := bson.M{
			"$set": setUpdate,
		}

		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
	} else {
		// Si hay otro error, lo devolvemos
		return err
	}

	return nil
}

// get stream by id
func (r *StreamRepository) GetStreamById(id primitive.ObjectID) (*streamdomain.Stream, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	FindStreamInDb := bson.D{
		{Key: "StreamerID", Value: id},
	}
	var FindStreamsById *streamdomain.Stream
	errCollStreams := GoMongoDBCollStreams.FindOne(context.Background(), FindStreamInDb).Decode(&FindStreamsById)
	return FindStreamsById, errCollStreams
}

func (r *StreamRepository) GetStreamSummaryById(id primitive.ObjectID) (*StreamSummarydomain.StreamSummary, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("StreamSummary")
	FindStreamInDb := bson.D{
		{Key: "StreamerID", Value: id},
	}
	var FindStreamsById *StreamSummarydomain.StreamSummary
	errCollStreams := GoMongoDBCollStreams.FindOne(context.Background(), FindStreamInDb).Decode(&FindStreamsById)
	return FindStreamsById, errCollStreams
}

func (r *StreamRepository) UpdateModChatSlowMode(updateInfo streamdomain.UpdateModChatSlowMode, id primitive.ObjectID) error {

	userFilter := bson.M{"_id": id}
	var user userdomain.User
	if err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users").FindOne(context.Background(), userFilter).Decode(&user); err != nil {
		return err
	}
	streamerName := user.NameUser

	var previousStream streamdomain.Stream
	if err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams").FindOne(context.Background(), bson.M{"Streamer": streamerName}).Decode(&previousStream); err != nil {
		return err
	}
	err := r.RedisDeleteKey(previousStream.ID.Hex() + "ModSlowMode")
	if err != nil {
		return err
	}
	streamFilter := bson.M{"Streamer": streamerName}
	update := bson.M{
		"$set": bson.M{
			"ModSlowMode": updateInfo.ModSlowMode,
		},
	}
	if _, err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams").UpdateOne(context.Background(), streamFilter, update); err != nil {
		return err
	}

	notification := map[string]interface{}{
		"action":      "ModSlowMode",
		"ModSlowMode": updateInfo.ModSlowMode,
	}
	r.PublishAction(previousStream.ID.Hex()+"action", notification)
	return nil
}

func (r *StreamRepository) PublishAction(roomID string, noty map[string]interface{}) error {

	chatMessageJSON, err := json.Marshal(noty)
	if err != nil {
		return err
	}
	err = r.redisClient.Publish(
		context.Background(),
		roomID,
		string(chatMessageJSON),
	).Err()
	if err != nil {
		return err
	}

	return err
}
func (r *StreamRepository) AddCommercialInStream(CommercialInStream int, id primitive.ObjectID) error {

	userFilter := bson.M{"_id": id}
	var user userdomain.User
	if err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users").FindOne(context.Background(), userFilter).Decode(&user); err != nil {
		return err
	}
	streamerName := user.NameUser

	var previousStream streamdomain.Stream
	if err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams").FindOne(context.Background(), bson.M{"Streamer": streamerName}).Decode(&previousStream); err != nil {
		return err
	}
	err := r.RedisDeleteKey(previousStream.ID.Hex() + "ModSlowMode")
	if err != nil {
		return err
	}
	streamFilter := bson.M{"Streamer": streamerName}
	update := bson.M{
		"$set": bson.M{
			"ModSlowMode": CommercialInStream,
		},
	}
	if _, err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams").UpdateOne(context.Background(), streamFilter, update); err != nil {
		return err
	}

	return nil
}
func (r *StreamRepository) RedisDeleteKey(key string) error {
	err := r.redisClient.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

// get stream by name user
func (r *StreamRepository) GetStreamByNameUser(nameUser string) (*streamdomain.Stream, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	FindStreamInDb := bson.D{
		{Key: "Streamer", Value: nameUser},
	}
	var FindStreamsByStreamer *streamdomain.Stream
	errCollStreams := GoMongoDBCollStreams.FindOne(context.Background(), FindStreamInDb).Decode(&FindStreamsByStreamer)
	return FindStreamsByStreamer, errCollStreams
}

// get streams by Categorie
func (r *StreamRepository) GetStreamsByCategory(Categorie string, page int) ([]streamdomain.Stream, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	skip := (page - 1) * 15

	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(15))

	FindStreamsInDb := bson.D{
		{Key: "StreamCategory", Value: Categorie},
		{Key: "Online", Value: true},
	}
	cursor, err := GoMongoDBCollStreams.Find(context.Background(), FindStreamsInDb, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var streams []streamdomain.Stream
	for cursor.Next(context.Background()) {
		var stream streamdomain.Stream
		if err := cursor.Decode(&stream); err != nil {
			return nil, err
		}
		streams = append(streams, stream)
	}

	if len(streams) == 0 {
		return nil, errors.New("no se encontraron streams")
	}

	return streams, nil
}

func (r *StreamRepository) GetAllsStreamsOnline(page int) ([]streamdomain.Stream, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	skip := (page - 1) * 15
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(15))
	cursor, err := GoMongoDBCollStreams.Find(context.Background(), bson.D{{Key: "Online", Value: true}}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var streams []streamdomain.Stream
	for cursor.Next(context.Background()) {
		var stream streamdomain.Stream
		if err := cursor.Decode(&stream); err != nil {
			return nil, err
		}
		streams = append(streams, stream)
	}

	if len(streams) == 0 {
		return nil, errors.New("no se encontraron streams")
	}

	return streams, nil
}
func (r *StreamRepository) GetStreamsMostViewed(page int) ([]streamdomain.Stream, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "ViewerCount", Value: -1}})
	findOptions.SetSkip(int64((page - 1) * 15))
	findOptions.SetLimit(int64(15))

	filter := bson.D{{Key: "Online", Value: true}}

	cursor, err := GoMongoDBCollStreams.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var streams []streamdomain.Stream
	for cursor.Next(context.Background()) {
		var stream streamdomain.Stream
		if err := cursor.Decode(&stream); err != nil {
			return nil, err
		}
		streams = append(streams, stream)
	}

	if len(streams) == 0 {
		return nil, errors.New("no se encontraron streams")
	}

	return streams, nil
}

func (r *StreamRepository) GetAllStreamsOnlineThatUserFollows(idValueObj primitive.ObjectID) ([]streamdomain.Stream, error) {
	GoMongoDBCollUsers := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	// Estructura para recibir el campo Following
	var user struct {
		Following map[primitive.ObjectID]userdomain.FollowInfo `bson:"Following"`
	}

	// Buscar solo el campo Following
	filter := bson.D{{Key: "_id", Value: idValueObj}}
	projection := bson.D{{Key: "Following", Value: 1}}

	if err := GoMongoDBCollUsers.FindOne(context.Background(), filter, options.FindOne().SetProjection(projection)).Decode(&user); err != nil {
		return nil, err
	}

	var followingIDs []primitive.ObjectID
	for id := range user.Following {
		followingIDs = append(followingIDs, id)
	}

	cursor, err := GoMongoDBCollStreams.Find(
		context.Background(),
		bson.D{
			{Key: "Online", Value: true},
			{Key: "StreamerID", Value: bson.D{{Key: "$in", Value: followingIDs}}},
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var streams []streamdomain.Stream
	for cursor.Next(context.Background()) {
		var stream streamdomain.Stream
		if err := cursor.Decode(&stream); err != nil {
			return nil, err
		}
		streams = append(streams, stream)
	}

	if len(streams) == 0 {
		return nil, errors.New("no se encontraron streams")
	}

	return streams, nil
}

// func (r *StreamRepository) GetAllStreamsOnlineThatUserFollows(idValueObj primitive.ObjectID, limit int64, offset int64) ([]streamdomain.Stream, error) {
// 	GoMongoDBCollUsers := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

// 	pipeline := mongo.Pipeline{
// 		{{Key: "$match", Value: bson.D{{Key: "_id", Value: idValueObj}}}},
// 		{{Key: "$lookup", Value: bson.D{
// 			{Key: "from", Value: "Streams"},
// 			{Key: "let", Value: bson.D{{Key: "following", Value: "$Following"}}},
// 			{Key: "pipeline", Value: mongo.Pipeline{
// 				{{Key: "$match", Value: bson.D{
// 					{Key: "$expr", Value: bson.D{
// 						{Key: "$and", Value: bson.A{
// 							bson.D{{Key: "$eq", Value: bson.A{"$Online", true}}},
// 							bson.D{{Key: "$in", Value: bson.A{"$StreamerID", "$$following"}}},
// 						}},
// 					}},
// 				}}},
// 				{{Key: "$limit", Value: limit}},
// 				{{Key: "$skip", Value: offset}},
// 				{{Key: "$sort", Value: bson.D{{Key: "StartTime", Value: -1}}}}, // Ordenar por StartTime, del más reciente al más antiguo (opcional)
// 			}},
// 			{Key: "as", Value: "streams"},
// 		}}},
// 		{{Key: "$unwind", Value: "$streams"}},
// 		{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$streams"}}}},
// 	}

// 	cursor, err := GoMongoDBCollUsers.Aggregate(context.Background(), pipeline)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(context.Background())

// 	var streams []streamdomain.Stream
// 	for cursor.Next(context.Background()) {
// 		var stream streamdomain.Stream
// 		if err := cursor.Decode(&stream); err != nil {
// 			return nil, err
// 		}
// 		streams = append(streams, stream)
// 	}

// 	if err := cursor.Err(); err != nil {
// 		return nil, err
// 	}

//		return streams, nil
//	}
//
// GetStremesIFollow
func (r *StreamRepository) GetStreamsIdsStreamer(idsUsersF []primitive.ObjectID) ([]streamdomain.Stream, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	filter := bson.D{
		{Key: "StreamerID", Value: bson.D{{Key: "$in", Value: idsUsersF}}},
		{Key: "Online", Value: true},
	}

	cursor, err := GoMongoDBCollStreams.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var streams []streamdomain.Stream
	for cursor.Next(context.Background()) {
		var stream streamdomain.Stream
		if err := cursor.Decode(&stream); err != nil {
			return nil, err
		}
		streams = append(streams, stream)
	}
	if len(streams) == 0 {
		return nil, errors.New("no se encontraron streams")
	}
	return streams, nil
}

func (r *StreamRepository) FindLatestStreamSummaryByStreamerID(streamerID primitive.ObjectID) (*StreamSummarydomain.StreamSummary, error) {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")

	filter := bson.M{"StreamerID": streamerID}
	opts := options.FindOne().SetSort(bson.D{{Key: "StartOfStream", Value: -1}})

	var streamSummary StreamSummarydomain.StreamSummary
	err := GoMongoDBCollStreamSummary.FindOne(ctx, filter, opts).Decode(&streamSummary)
	if err != nil {
		return nil, err
	}

	return &streamSummary, nil
}

func (r *StreamRepository) CloseStream(key string) error {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	filter := bson.D{
		{Key: "KeyTransmission", Value: key},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Online", Value: false},
		}},
	}

	_, err := GoMongoDBCollStreams.UpdateOne(context.Background(), filter, update)

	return err
}
func (r *StreamRepository) Update_thumbnail(cmt, image string) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")

	GoMongoDBCollUser := GoMongoDB.Collection("Users")
	GoMongoDBCollStreams := GoMongoDB.Collection("Streams")

	var userCmt *userdomain.GetUser
	filterUser := bson.D{
		{Key: "Cmt", Value: cmt},
	}
	err := GoMongoDBCollUser.FindOne(context.Background(), filterUser).Decode(&userCmt)
	if err != nil {
		return err
	}
	filter := bson.D{
		{Key: "StreamerID", Value: userCmt.ID},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "StreamThumbnailPermanent", Value: false}},                                  // Existe y es false
			bson.D{{Key: "StreamThumbnailPermanent", Value: bson.D{{Key: "$exists", Value: false}}}}, // No existe
		}},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "StreamThumbnail", Value: image},
			{Key: "StreamThumbnailPermanent", Value: false}, // Agregar si no existe
		}},
	}

	_, err = GoMongoDBCollStreams.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *StreamRepository) Update_start_date(req streamdomain.Update_start_date) error {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	filter := bson.D{
		{Key: "KeyTransmission", Value: req.Key},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "StartDate", Value: req.Date},
		}},
	}

	_, err := GoMongoDBCollStreams.UpdateOne(context.Background(), filter, update)

	return err
}
func (r *StreamRepository) UpdateStreamInfo(updateInfo streamdomain.UpdateStreamInfo, id primitive.ObjectID) error {
	userFilter := bson.M{"_id": id}
	db := r.mongoClient.Database("PINKKER-BACKEND")
	var user userdomain.GetUser
	if err := db.Collection("Users").FindOne(context.Background(), userFilter).Decode(&user); err != nil {
		return err
	}
	streamerName := user.NameUser

	var previousStream streamdomain.Stream
	if err := db.Collection("Streams").FindOne(context.Background(), bson.M{"Streamer": streamerName}).Decode(&previousStream); err != nil {
		return err
	}
	categoriaImgUpdateFilter := bson.M{"Name": updateInfo.Category}
	var categoriaImgUpdate streamdomain.Categoria

	dbColeccionCategorias := db.Collection("Categorias")
	err := dbColeccionCategorias.FindOne(context.Background(), categoriaImgUpdateFilter).Decode(&categoriaImgUpdate)
	if err != nil {
		return err
	}
	streamFilter := bson.M{"Streamer": streamerName}
	update := bson.M{
		"$set": bson.M{
			"StreamTitle":              updateInfo.Title,
			"StreamNotification":       updateInfo.Notification,
			"StreamCategory":           updateInfo.Category,
			"StreamTag":                updateInfo.Tag,
			"StreamIdiom":              updateInfo.Idiom,
			"StartDate":                updateInfo.Date,
			"ImageCategorie":           categoriaImgUpdate.Img,
			"StreamThumbnailPermanent": updateInfo.StreamThumbnailPermanent,
			"AuthorizationToView":      updateInfo.AuthorizationToView,
		},
	}
	if updateInfo.StreamThumbnailPermanent {
		if updateInfo.ThumbnailURL != "" {
			update["$set"].(bson.M)["StreamThumbnail"] = updateInfo.ThumbnailURL
		} else {
			update["$set"].(bson.M)["StreamThumbnailPermanent"] = false
		}
	}

	if _, err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams").UpdateOne(context.Background(), streamFilter, update); err != nil {
		return err
	}

	categoryFilter := bson.M{"Name": previousStream.StreamCategory}
	categoryUpdate := bson.M{"$inc": bson.M{"Spectators": -previousStream.ViewerCount}}
	if _, err := dbColeccionCategorias.UpdateOne(context.Background(), categoryFilter, categoryUpdate); err != nil {
		return err
	}

	categoryUpdate = bson.M{"$inc": bson.M{"Spectators": previousStream.ViewerCount}}
	if _, err := dbColeccionCategorias.UpdateOne(context.Background(), bson.M{"Name": updateInfo.Category}, categoryUpdate); err != nil {
		return err
	}
	cacheKey := "stream:" + streamerName
	_, err = r.redisClient.Del(context.Background(), cacheKey).Result()
	if err != nil {
		fmt.Println(err)
	}
	notification := map[string]interface{}{
		"action":         "updateInfoStream",
		"title":          updateInfo.Title,
		"StreamCategory": updateInfo.Category,
	}
	r.PublishAction(previousStream.ID.Hex()+"action", notification)
	return nil
}

func (r *StreamRepository) UpdateModChat(updateInfo streamdomain.UpdateModChat, id primitive.ObjectID) error {
	userFilter := bson.M{"_id": id}
	var user userdomain.GetUser
	if err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users").FindOne(context.Background(), userFilter).Decode(&user); err != nil {
		return err
	}
	streamerName := user.NameUser

	var previousStream streamdomain.Stream
	if err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams").FindOne(context.Background(), bson.M{"Streamer": streamerName}).Decode(&previousStream); err != nil {
		return err
	}
	streamFilter := bson.M{"Streamer": streamerName}
	update := bson.M{
		"$set": bson.M{
			"ModChat": updateInfo.Mod,
		},
	}
	if _, err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams").UpdateOne(context.Background(), streamFilter, update); err != nil {
		return err
	}
	exist, err := r.redisClient.Exists(context.Background(), previousStream.ID.Hex()).Result()
	if err != nil {
		return err
	}
	if exist == 1 {
		err := r.redisClient.Set(context.Background(), previousStream.ID.Hex(), updateInfo.Mod, 0).Err()
		if err != nil {
			return err
		}
	} else {
		err := r.redisClient.Set(context.Background(), previousStream.ID.Hex(), updateInfo.Mod, 0).Err()
		if err != nil {
			return err
		}
	}
	notification := map[string]interface{}{
		"action":  "ModChat",
		"ModChat": updateInfo.Mod,
	}
	r.PublishAction(previousStream.ID.Hex()+"action", notification)
	return nil
}

func (r *StreamRepository) Update_Emotes(idUser primitive.ObjectID, date int) error {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	filter := bson.D{
		{Key: "StreamerID", Value: idUser},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "StartDate", Value: date},
		}},
	}

	_, err := GoMongoDBCollStreams.UpdateOne(context.Background(), filter, update)

	return err
}

func (r *StreamRepository) Streamings_online() (int, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	FindStreamsInDb := bson.D{
		{Key: "Online", Value: true},
	}

	cursor, err := GoMongoDBCollStreams.Find(context.Background(), FindStreamsInDb)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.Background())

	var streams []streamdomain.Stream
	for cursor.Next(context.Background()) {
		var stream streamdomain.Stream
		if err := cursor.Decode(&stream); err != nil {
			return 0, err
		}
		streams = append(streams, stream)
	}

	return len(streams), nil

}

func (r *StreamRepository) GetCategories() ([]streamdomain.Categoria, error) {
	GoMongoDBCollCategorias := r.mongoClient.Database("PINKKER-BACKEND").Collection("Categorias")
	FindCategoriasInDb := bson.D{}

	cursor, err := GoMongoDBCollCategorias.Find(context.Background(), FindCategoriasInDb)
	if err != nil {
		return []streamdomain.Categoria{}, err
	}
	defer cursor.Close(context.Background())

	var Categorias []streamdomain.Categoria
	for cursor.Next(context.Background()) {
		var caregorie streamdomain.Categoria
		if err := cursor.Decode(&caregorie); err != nil {
			return []streamdomain.Categoria{}, err
		}
		Categorias = append(Categorias, caregorie)
	}

	return Categorias, err

}

func (r *StreamRepository) GetCategia(cate string) (streamdomain.Categoria, error) {
	GoMongoDBCollCategorias := r.mongoClient.Database("PINKKER-BACKEND").Collection("Categorias")

	Find := bson.D{
		{Key: "Name", Value: cate},
	}
	var FindStreamsById streamdomain.Categoria
	errCollStreams := GoMongoDBCollCategorias.FindOne(context.Background(), Find).Decode(&FindStreamsById)
	return FindStreamsById, errCollStreams
}

func (r *StreamRepository) ValidateStreamAccess(idUser, idStreamer primitive.ObjectID) (bool, error) {
	// Obtener la información del streamer
	filterUser := bson.D{
		{Key: "_id", Value: idUser},
	}

	stream, err := r.GetStreamById(idStreamer)
	if err != nil {
		return false, fmt.Errorf("error al obtener el stream del streamer: %v", err)
	}

	// Validar si el stream requiere autorización
	authToView := stream.AuthorizationToView
	if !authToView["pinkker_prime"] && !authToView["subscription"] {
		// Si no requiere autorización, permitir acceso
		return true, nil
	}

	// Obtener al usuario que intenta acceder
	user, err := r.getUser(filterUser)
	if err != nil {
		return false, fmt.Errorf("error al obtener al usuario: %v", err)
	}

	// Validar acceso según las reglas de autorización
	// 1. Validar si el usuario tiene Pinkker Prime activo
	if authToView["pinkker_prime"] {
		if time.Now().After(user.PinkkerPrime.SubscriptionEnd) {
			return false, fmt.Errorf("el usuario no tiene Pinkker Prime activo")
		}
	}

	// 2. Validar si el usuario tiene una suscripción válida
	if authToView["subscription"] {
		subscription, err := r.getSubscriptionByUserIDs(idUser, idStreamer)
		if err != nil {
			return false, fmt.Errorf("error al obtener la suscripción: %v", err)
		}
		if time.Now().After(subscription.SubscriptionEnd) {
			return false, fmt.Errorf("la suscripción del usuario ha expirado")
		}
	}

	// Si cumple con todas las condiciones
	return true, nil
}
func (r *StreamRepository) getSubscriptionByUserIDs(sourceUserID, destUserID primitive.ObjectID) (subscriptiondomain.Subscription, error) {
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
