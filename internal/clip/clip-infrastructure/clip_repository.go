// clip-infrastructure/clip_repository.go
package clipinfrastructure

import (
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClipRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewClipRepository(redisClient *redis.Client, mongoClient *mongo.Client) *ClipRepository {
	return &ClipRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}

func (c *ClipRepository) SaveClip(clip *clipdomain.Clip) (*clipdomain.Clip, error) {
	database := c.mongoClient.Database("pinkker")
	clipCollection := database.Collection("Clips")
	userCollection := database.Collection("Users")

	result, err := clipCollection.InsertOne(context.Background(), clip)
	if err != nil {
		return nil, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("no se pudo obtener el ID insertado")
	}
	clip.ID = insertedID

	filterUser := bson.M{"_id": clip.UserID}
	update := bson.M{"$push": bson.M{"Clips": insertedID}}

	opts := options.Update().SetUpsert(false)

	resultuserCollection, err := userCollection.UpdateOne(context.Background(), filterUser, update, opts)
	if err != nil {
		return clip, err
	}

	if resultuserCollection.ModifiedCount == 0 {
		return clip, errors.New("No se encontraron documentos para actualizar.")
	}

	return clip, err
}
func (c *ClipRepository) UpdateClip(clipID primitive.ObjectID, newURL string) {
	clipCollection := c.mongoClient.Database("pinkker").Collection("Clips")

	filter := bson.M{"_id": clipID}

	update := bson.M{"$set": bson.M{"url": newURL}}

	opts := options.Update().SetUpsert(false)

	clipCollection.UpdateOne(context.Background(), filter, update, opts)

}
func (c *ClipRepository) UpdateClipPreviouImage(clipID primitive.ObjectID, newURL string) {
	clipCollection := c.mongoClient.Database("pinkker").Collection("Clips")

	filter := bson.M{"_id": clipID}

	update := bson.M{"$set": bson.M{"PreviouImage": newURL}}

	opts := options.Update().SetUpsert(false)

	clipCollection.UpdateOne(context.Background(), filter, update, opts)

}
func (c *ClipRepository) FindrClipId(IdClip primitive.ObjectID) (*clipdomain.Clip, error) {
	GoMongoDBColl := c.mongoClient.Database("pinkker").Collection("Clips")
	FindClipInDb := bson.D{
		{Key: "_id", Value: IdClip},
	}

	var findClipInDbExist *clipdomain.Clip
	err := GoMongoDBColl.FindOne(context.Background(), FindClipInDb).Decode(&findClipInDbExist)
	if err != nil {
		return nil, err
	}

	update := bson.D{
		{Key: "$inc", Value: bson.D{{Key: "views", Value: 1}}},
	}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), FindClipInDb, update)
	if err != nil {
		return nil, err
	}

	err = GoMongoDBColl.FindOne(context.Background(), FindClipInDb).Decode(&findClipInDbExist)
	if err != nil {
		return nil, err
	}

	return findClipInDbExist, nil
}

func (c *ClipRepository) FindUser(totalKey string) (*userdomain.User, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("pinkker").Collection("Users")
	var FindUserInDb primitive.D
	FindUserInDb = bson.D{
		{Key: "KeyTransmission", Value: "live" + totalKey},
	}
	var findUserInDbExist *userdomain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&findUserInDbExist)
	return findUserInDbExist, errCollUsers
}
func (c *ClipRepository) FindUserId(FindUserId string) (*userdomain.User, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("pinkker").Collection("Users")
	var FindUserInDb primitive.D
	FindUserInDb = bson.D{
		{Key: "NameUser", Value: FindUserId},
	}
	var findUserInDbExist *userdomain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&findUserInDbExist)
	return findUserInDbExist, errCollUsers
}
func (c *ClipRepository) FindCategorieStream(StreamerID primitive.ObjectID) (*streamdomain.Stream, error) {
	GoMongoDBColl := c.mongoClient.Database("pinkker").Collection("Streams")
	var FindInDb primitive.D
	FindInDb = bson.D{
		{Key: "StreamerID", Value: StreamerID},
	}
	var findStream *streamdomain.Stream
	errCollUsers := GoMongoDBColl.FindOne(context.Background(), FindInDb).Decode(&findStream)
	return findStream, errCollUsers
}
func (c *ClipRepository) GetClipsNameUser(page int, GetClipsNameUser string) ([]clipdomain.Clip, error) {
	GoMongoDBColl := c.mongoClient.Database("pinkker").Collection("Clips")

	options := options.Find()
	options.SetSort(bson.D{{"TimeStamp", -1}})
	options.SetSkip(int64((page - 1) * 10))
	options.SetLimit(10)
	filter := bson.D{{"NameUserCreator", GetClipsNameUser}}

	cursor, err := GoMongoDBColl.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var clips []clipdomain.Clip
	if err := cursor.All(context.Background(), &clips); err != nil {
		return nil, err
	}

	return clips, nil
}
func (c *ClipRepository) GetClipsCategory(page int, Category string, lastClipID primitive.ObjectID) ([]clipdomain.Clip, error) {
	GoMongoDBColl := c.mongoClient.Database("pinkker").Collection("Clips")

	options := options.Find()
	options.SetSort(bson.D{{"_id", -1}}) // Ordenar por _id en orden descendente
	options.SetSkip(int64((page - 1) * 10))
	options.SetLimit(10)

	var filter bson.D
	if Category != "" {
		filter = bson.D{{"Category", Category}}
	}

	// Si lastClipID no está vacío, añade un filtro para traer clips después de este _id
	if !lastClipID.IsZero() {
		filter = append(filter, bson.E{"_id", bson.M{"$lt": lastClipID}})
	}

	cursor, err := GoMongoDBColl.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var clips []clipdomain.Clip
	if err := cursor.All(context.Background(), &clips); err != nil {
		return nil, err
	}

	return clips, nil
}
func (c *ClipRepository) GetClipsMostViewed(page int) ([]clipdomain.Clip, error) {
	GoMongoDBColl := c.mongoClient.Database("pinkker").Collection("Clips")

	options := options.Find()
	options.SetSort(bson.D{{"Views", -1}})
	options.SetSkip(int64((page - 1) * 10))
	options.SetLimit(10)

	filter := bson.D{}

	cursor, err := GoMongoDBColl.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var clips []clipdomain.Clip
	if err := cursor.All(context.Background(), &clips); err != nil {
		return nil, err
	}

	return clips, nil
}

func (c *ClipRepository) LikeClip(ClipId, idValueToken primitive.ObjectID) error {
	GoMongoDB := c.mongoClient.Database("pinkker")
	GoMongoDBColl := GoMongoDB.Collection("Clips")

	count, err := GoMongoDBColl.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: ClipId}})
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("el ClipId no existe")
	}
	filter := bson.D{{Key: "_id", Value: ClipId}}
	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "Likes", Value: idValueToken}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	GoMongoDBColl = GoMongoDB.Collection("Users")

	filter = bson.D{{Key: "_id", Value: idValueToken}}
	update = bson.D{{Key: "$addToSet", Value: bson.D{{Key: "ClipsLikes", Value: ClipId}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}
func (c *ClipRepository) ClipDislike(ClipId, idValueToken primitive.ObjectID) error {
	GoMongoDB := c.mongoClient.Database("pinkker")
	GoMongoDBColl := GoMongoDB.Collection("Clips")

	count, err := GoMongoDBColl.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: ClipId}})
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("el ClipId no existe")
	}
	filter := bson.D{{Key: "_id", Value: ClipId}}
	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "Likes", Value: idValueToken}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	GoMongoDBColl = GoMongoDB.Collection("Users")

	filter = bson.D{{Key: "_id", Value: idValueToken}}
	update = bson.D{{Key: "$pull", Value: bson.D{{Key: "ClipsLikes", Value: ClipId}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil

}
func (c *ClipRepository) MoreViewOfTheClip(ClipId primitive.ObjectID) error {

	GoMongoDB := c.mongoClient.Database("pinkker")
	GoMongoDBColl := GoMongoDB.Collection("Clips")

	filter := bson.D{{Key: "_id", Value: ClipId}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "views", Value: 1}}}}

	_, err := GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	return err
}
