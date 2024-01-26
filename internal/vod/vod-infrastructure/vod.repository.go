package vodinfrastructure

import (
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	voddomain "PINKKER-BACKEND/internal/vod/vod-domain"
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VodRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewVodRepository(redisClient *redis.Client, mongoClient *mongo.Client) *VodRepository {
	return &VodRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (v *VodRepository) GetVodsByStreamer(streamer string, limit string, sort string) ([]*voddomain.Vod, error) {
	GoMongoDBCollVod := v.mongoClient.Database("PINKKER-BACKEND").Collection("Vod")
	var FindvodInDb primitive.D
	FindvodInDb = bson.D{
		{Key: "Streamer", Value: streamer},
	}

	findOptions := options.Find()
	findOptions.SetProjection(bson.D{
		{Key: "__v", Value: 0},
		{Key: "StreamerId", Value: 0},
	})

	if limitInt, err := strconv.ParseInt(limit, 10, 64); err == nil {
		findOptions.SetLimit(limitInt)
	}

	cur, err := GoMongoDBCollVod.Find(context.Background(), FindvodInDb, findOptions)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	var vods []*voddomain.Vod

	for cur.Next(context.Background()) {
		var vod *voddomain.Vod
		if err := cur.Decode(&vod); err != nil {
			return nil, err
		}
		vods = append(vods, vod)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return vods, nil
}

func (v *VodRepository) GetVodWithId(vodId string) (*voddomain.Vod, error) {
	GoMongoDBCollVod := v.mongoClient.Database("PINKKER-BACKEND").Collection("Vod")
	var FindvodInDb primitive.D
	FindvodInDb = bson.D{
		{Key: "_id", Value: vodId},
	}

	var findVodInDbExist *voddomain.Vod
	errCollUsers := GoMongoDBCollVod.FindOne(context.Background(), FindvodInDb).Decode(&findVodInDbExist)
	return findVodInDbExist, errCollUsers
}
func (v *VodRepository) GetUserByStreamKey(streamKey string) (*userdomain.User, error) {
	GoMongoDBCollUsers := v.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var FindUserInDb primitive.D
	FindUserInDb = bson.D{
		{Key: "KeyTransmission", Value: "live" + streamKey},
	}

	var user *userdomain.User
	err := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&user)
	return user, err
}

func (v *VodRepository) GetStreamByStreamer(streamer string) (*streamdomain.Stream, error) {
	GoMongoDBCollStreams := v.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	var FindStreamInDb primitive.D
	FindStreamInDb = bson.D{
		{Key: "Streamer", Value: streamer},
	}

	var stream *streamdomain.Stream
	err := GoMongoDBCollStreams.FindOne(context.Background(), FindStreamInDb).Decode(&stream)
	return stream, err
}

func (v *VodRepository) CreateVod(vod *voddomain.Vod) error {
	GoMongoDBCollVod := v.mongoClient.Database("PINKKER-BACKEND").Collection("Vod")
	_, err := GoMongoDBCollVod.InsertOne(context.Background(), vod)
	return err
}
