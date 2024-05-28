package streaminfrastructure

import (
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-domain"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"context"
	"errors"
	"fmt"
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

	return nil
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
func (r *StreamRepository) GetStreamsByCategorie(Categorie string, page int) ([]streamdomain.Stream, error) {
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
	findOptions.SetSort(bson.D{{"ViewerCount", -1}})
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

	var user userdomain.User
	if err := GoMongoDBCollUsers.FindOne(context.Background(), bson.D{{Key: "_id", Value: idValueObj}}).Decode(&user); err != nil {
		return nil, err
	}
	var followingIDs []primitive.ObjectID

	for ids := range user.Following {
		followingIDs = append(followingIDs, ids)
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
func (r *StreamRepository) UpdateOnline(Key string, state bool) error {
	session, err := r.mongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	err = session.StartTransaction()
	if err != nil {
		return err
	}

	ctx := mongo.NewSessionContext(context.Background(), session)

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")
	GoMongoDBCollStreams := GoMongoDB.Collection("Streams")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")

	filterUsers := bson.D{
		{Key: "KeyTransmission", Value: Key},
	}

	var userFind userdomain.User
	err = GoMongoDBCollUsers.FindOne(ctx, filterUsers).Decode(&userFind)
	if err != nil {
		session.AbortTransaction(ctx)
		return err
	}

	filterStreams := bson.D{
		{Key: "StreamerID", Value: userFind.ID},
	}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	updateStreams := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Online", Value: state},
			{Key: "StartDate", Value: time.Now()},
		}},
	}
	var StreamFind streamdomain.Stream

	err = GoMongoDBCollStreams.FindOneAndUpdate(ctx, filterStreams, updateStreams, options).Decode(&StreamFind)
	if err != nil {
		session.AbortTransaction(ctx)
		return err
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
		return err
	}
	notifyOnlineStreamer := []string{}
	if state {
		for _, followInfo := range userFind.Followers {
			if followInfo.Notifications {
				notifyOnlineStreamer = append(notifyOnlineStreamer, followInfo.Email)
			}
		}
		_ = helpers.ResendNotificationStreamerOnline(userFind.NameUser, notifyOnlineStreamer)

		// modo de chat cuandos se prende
		exist, err := r.redisClient.Exists(context.Background(), StreamFind.ID.Hex()).Result()
		if err != nil {
			return err
		}
		if exist == 1 {
			err := r.redisClient.Set(context.Background(), StreamFind.ID.Hex(), StreamFind.ModChat, 0).Err()
			if err != nil {
				return err
			}
		} else {
			err := r.redisClient.Set(context.Background(), StreamFind.ID.Hex(), StreamFind.ModChat, 0).Err()
			if err != nil {
				return err
			}
		}
		startFollowersCount := len(userFind.Followers)
		startSubsCount := len(userFind.Subscriptions)

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
			StartSubsCount:       startSubsCount,
		}

		_, err = GoMongoDBCollStreamSummary.InsertOne(ctx, summary)
		if err != nil {
			return err
		}
	} else {
		_, err := r.redisClient.Del(context.Background(), StreamFind.ID.Hex()).Result()
		if err != nil {
			fmt.Println(err)
		}

		latestSummary, err := r.FindLatestStreamSummaryByStreamerID(userFind.ID)
		if err != nil {
			session.AbortTransaction(ctx)
			return err
		}

		newFollowersCount := len(userFind.Followers) - latestSummary.StartFollowersCount
		newSubsCount := len(userFind.Subscriptions) - latestSummary.StartSubsCount
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

		// Actualizar el resumen del stream
		updateSummary := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "EndOfStream", Value: time.Now()},
				{Key: "NewFollowers", Value: newFollowersCount},
				{Key: "NewSubscriptions", Value: newSubsCount},
				{Key: "AverageViewers", Value: AverageViewers},
				{Key: "MaxViewers", Value: maxViewers},
			}},
		}

		_, err = GoMongoDBCollStreamSummary.UpdateOne(ctx, bson.M{"_id": latestSummary.ID}, updateSummary)
		if err != nil {
			return err
		}
	}

	err = session.CommitTransaction(ctx)
	if err != nil {
		return err
	}

	return nil
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

	var userCmt *userdomain.User
	filterUser := bson.D{
		{Key: "Cmt", Value: cmt},
	}
	err := GoMongoDBCollUser.FindOne(context.Background(), filterUser).Decode(&userCmt)
	if err != nil {
		return err
	}
	filter := bson.D{
		{Key: "StreamerID", Value: userCmt.ID},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "StreamThumbnail", Value: image},
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
	var user userdomain.User
	if err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users").FindOne(context.Background(), userFilter).Decode(&user); err != nil {
		return err
	}
	streamerName := user.NameUser

	var previousStream streamdomain.Stream
	if err := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams").FindOne(context.Background(), bson.M{"Streamer": streamerName}).Decode(&previousStream); err != nil {
		return err
	}
	categoriaImgUpdateFilter := bson.M{"Name": updateInfo.Category}
	var categoriaImgUpdate streamdomain.Categoria
	dbColeccionCategorias := r.mongoClient.Database("PINKKER-BACKEND").Collection("Categorias")
	err := dbColeccionCategorias.FindOne(context.Background(), categoriaImgUpdateFilter).Decode(&categoriaImgUpdate)
	if err != nil {
		return err
	}
	streamFilter := bson.M{"Streamer": streamerName}
	update := bson.M{
		"$set": bson.M{
			"StreamTitle":        updateInfo.Title,
			"StreamNotification": updateInfo.Notification,
			"StreamCategory":     updateInfo.Category,
			"StreamTag":          updateInfo.Tag,
			"StreamIdiom":        updateInfo.Idiom,
			"StartDate":          updateInfo.Date,
			"ImageCategorie":     categoriaImgUpdate.Img,
		},
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

	return nil
}
func (r *StreamRepository) UpdateModChat(updateInfo streamdomain.UpdateModChat, id primitive.ObjectID) error {
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
