package streaminfrastructure

import (
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
func (r *StreamRepository) GetStreamsByCategorie(Categorie string) ([]streamdomain.Stream, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	FindStreamsInDb := bson.D{
		{Key: "StreamCategory", Value: Categorie},
		{Key: "Online", Value: true},
	}

	cursor, err := GoMongoDBCollStreams.Find(context.Background(), FindStreamsInDb)
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
func (r *StreamRepository) GetAllsStreamsOnline() ([]streamdomain.Stream, error) {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	cursor, err := GoMongoDBCollStreams.Find(context.Background(), bson.D{{Key: "Online", Value: true}})
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
func (r *StreamRepository) Update_online(Key string, state bool) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBColluser := GoMongoDB.Collection("Users")

	filter := bson.D{
		{Key: "KeyTransmission", Value: Key},
	}
	var userFind userdomain.User
	GoMongoDBColluser.FindOne(context.Background(), filter).Decode(&userFind)
	fmt.Println(userFind.ID)

	GoMongoDBCollStreams := GoMongoDB.Collection("Streams")

	filter = bson.D{
		{Key: "StreamerID", Value: userFind.ID},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Online", Value: true},
		}},
	}

	_, err := GoMongoDBCollStreams.UpdateOne(context.Background(), filter, update)

	return err
}
func (r *StreamRepository) CloseStream(idUser primitive.ObjectID) error {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	filter := bson.D{
		{Key: "StreamerID", Value: idUser},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Online", Value: false},
		}},
	}

	_, err := GoMongoDBCollStreams.UpdateOne(context.Background(), filter, update)

	return err
}
func (r *StreamRepository) Update_thumbnail(idUser primitive.ObjectID, image string) error {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")

	filter := bson.D{
		{Key: "StreamerID", Value: idUser},
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "StreamThumbnail", Value: image},
		}},
	}

	_, err := GoMongoDBCollStreams.UpdateOne(context.Background(), filter, update)

	return err
}
func (r *StreamRepository) Update_start_date(idUser primitive.ObjectID, date int) error {
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
func (r *StreamRepository) UpdateStreamInfo(streamID primitive.ObjectID, updateInfo streamdomain.UpdateStreamInfo) error {
	GoMongoDBCollStreams := r.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	update := bson.M{
		"$set": bson.M{
			"StreamTitle":        updateInfo.Title,
			"StreamNotification": updateInfo.Notification,
			"StreamCategory":     updateInfo.Category,
			"StreamTag":          updateInfo.Tag,
			"StreamIdiom":        updateInfo.Idiom,
			// "StartDate":updateInfo.Date,
		},
	}
	updata, err := GoMongoDBCollStreams.UpdateOne(context.Background(), bson.M{"StreamerID": streamID}, update)
	if err != nil {
		return err
	}
	if updata.MatchedCount == 0 {
		return errors.New("Not Found")
	}
	fmt.Println(updata.MatchedCount)
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

// beta

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
