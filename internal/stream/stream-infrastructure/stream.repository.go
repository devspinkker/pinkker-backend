package streaminfrastructure

import (
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"errors"
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
	skip := (page - 1*15)
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
func (r *StreamRepository) GetAllStreamsOnlineThatUserFollows(idValueObj primitive.ObjectID) ([]streamdomain.Stream, error) {
	GoMongoDBCollUsers := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	var user userdomain.User
	if err := GoMongoDBCollUsers.FindOne(context.Background(), bson.D{{Key: "_id", Value: idValueObj}}).Decode(&user); err != nil {
		return nil, err
	}

	cursor, err := GoMongoDBCollStreams.Find(
		context.Background(),
		bson.D{
			{Key: "Online", Value: true},
			{Key: "StreamerID", Value: bson.D{{Key: "$in", Value: user.Following}}},
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

	updateStreams := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Online", Value: state},
			{Key: "StartDate", Value: time.Now()},
		}},
	}

	_, err = GoMongoDBCollStreams.UpdateOne(ctx, filterStreams, updateStreams)
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

	err = session.CommitTransaction(ctx)
	if err != nil {
		return err
	}

	return nil
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
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")

	GoMongoDBCollUser := GoMongoDB.Collection("Users")
	var userCmt *userdomain.User
	filterUser := bson.D{
		{Key: "_id", Value: id},
	}
	err := GoMongoDBCollUser.FindOne(context.Background(), filterUser).Decode(&userCmt)
	if err != nil {
		return err
	}

	GoMongoDBCollStreams := GoMongoDB.Collection("Streams")
	update := bson.M{
		"$set": bson.M{
			"StreamTitle":        updateInfo.Title,
			"StreamNotification": updateInfo.Notification,
			"StreamCategory":     updateInfo.Category,
			"StreamTag":          updateInfo.Tag,
			"StreamIdiom":        updateInfo.Idiom,
			"StartDate":          updateInfo.Date,
		},
	}

	updata, err := GoMongoDBCollStreams.UpdateOne(context.Background(), bson.M{"Streamer": userCmt.NameUser}, update)
	if err != nil {
		return err
	}

	if updata.MatchedCount == 0 {
		return errors.New("Not Found")
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

func (r *StreamRepository) GetCategories() (error, []streamdomain.Categoria) {
	GoMongoDBCollCategorias := r.mongoClient.Database("PINKKER-BACKEND").Collection("Categorias")
	FindCategoriasInDb := bson.D{}

	cursor, err := GoMongoDBCollCategorias.Find(context.Background(), FindCategoriasInDb)
	if err != nil {
		return err, []streamdomain.Categoria{}
	}
	defer cursor.Close(context.Background())

	var Categorias []streamdomain.Categoria
	for cursor.Next(context.Background()) {
		var caregorie streamdomain.Categoria
		if err := cursor.Decode(&caregorie); err != nil {
			return err, []streamdomain.Categoria{}
		}
		Categorias = append(Categorias, caregorie)
	}

	return err, Categorias

}
