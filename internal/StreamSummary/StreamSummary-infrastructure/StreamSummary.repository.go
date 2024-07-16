package StreamSummaryinfrastructure

import (
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-domain"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StreamSummaryRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewStreamSummaryRepository(redisClient *redis.Client, mongoClient *mongo.Client) *StreamSummaryRepository {
	return &StreamSummaryRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (r *StreamSummaryRepository) GetStreamSummaryByID(id primitive.ObjectID) (*StreamSummarydomain.StreamSummary, error) {
	ctx := context.Background()

	// Obtener la base de datos y la colección
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	filter := bson.M{"_id": id}

	opts := options.FindOne()

	var streamSummary StreamSummarydomain.StreamSummary
	err := collection.FindOne(ctx, filter, opts).Decode(&streamSummary)
	if err != nil {
		return nil, err
	}

	return &streamSummary, nil
}
func (r *StreamSummaryRepository) UpdateStreamSummary(StreamerID primitive.ObjectID, data StreamSummarydomain.UpdateStreamSummary) error {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")

	// Definir el filtro para encontrar el resumen del stream por StreamerID
	filter := bson.M{"StreamerID": StreamerID}

	// Definir la actualización con los campos proporcionados en data
	update := bson.M{
		"$set": bson.M{
			"AverageViewers":   data.AverageViewers,
			"MaxViewers":       data.MaxViewers,
			"NewFollowers":     data.NewFollowers,
			"NewSubscriptions": data.NewSubscriptions,
			"Advertisements":   data.Advertisements,
		},
	}

	// Ejecutar la actualización en la colección de StreamSummary
	_, err := GoMongoDBCollStreamSummary.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
func (r *StreamSummaryRepository) AddAds(idValueObj primitive.ObjectID, AddAds StreamSummarydomain.AddAds) error {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")
	GoMongoDBCollAdvertisements := GoMongoDB.Collection("Advertisements")

	filter := bson.M{"_id": idValueObj}

	result := GoMongoDBCollUsers.FindOne(ctx, filter)
	if result.Err() != nil {
		return result.Err()
	}

	key := "ADS_" + idValueObj.Hex()
	exists, err := r.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		return nil
	}

	filterSummary := bson.M{"StreamerID": AddAds.StreamerID}
	opts := options.FindOne().SetSort(bson.D{{Key: "StartOfStream", Value: -1}})
	var streamSummary StreamSummarydomain.StreamSummary
	err = GoMongoDBCollStreamSummary.FindOne(ctx, filterSummary, opts).Decode(&streamSummary)
	if err != nil {
		return err
	}

	updateStream := bson.M{
		"$inc": bson.M{"Advertisements": 1},
	}
	filterUpdate := bson.M{"_id": streamSummary.ID}
	_, err = GoMongoDBCollStreamSummary.UpdateOne(ctx, filterUpdate, updateStream)
	if err != nil {
		return err
	}

	advertisementFilter := bson.M{"_id": AddAds.AdvertisementsId}
	advertisementUpdate := bson.M{"$inc": bson.M{"Impressions": 1}}
	_, err = GoMongoDBCollAdvertisements.UpdateOne(ctx, advertisementFilter, advertisementUpdate)
	if err != nil {
		return err
	}

	err = r.redisClient.Set(ctx, key, "true", time.Minute).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *StreamSummaryRepository) AverageViewers(StreamerID primitive.ObjectID) error {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")
	GoMongoDBColl := GoMongoDB.Collection("Streams")

	filter := bson.M{"StreamerID": StreamerID}
	Stream := streamdomain.Stream{}
	result := GoMongoDBColl.FindOne(ctx, filter).Decode(&Stream)
	if result != nil {
		return result
	}

	filter = bson.M{"StreamerID": StreamerID}

	opts := options.FindOne().SetSort(bson.D{{Key: "StartOfStream", Value: -1}})

	var streamSummary StreamSummarydomain.StreamSummary
	err := GoMongoDBCollStreamSummary.FindOne(ctx, filter, opts).Decode(&streamSummary)
	if err != nil {
		return err
	}
	currentDateTime := time.Now().Format("2006-01-02 15:04:05")

	update := bson.M{
		"$set": bson.M{
			"AverageViewersByTime." + currentDateTime: Stream.ViewerCount,
		},
	}
	filterUpdate := bson.M{"_id": streamSummary.ID}
	_, err = GoMongoDBCollStreamSummary.UpdateOne(ctx, filterUpdate, update)
	if err != nil {
		return err
	}

	return nil
}
func (r *StreamSummaryRepository) GetLastSixStreamSummariesBeforeDate(StreamerID primitive.ObjectID, date time.Time) ([]StreamSummarydomain.StreamSummary, error) {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")

	filter := bson.M{
		"StreamerID": StreamerID,
		"StartOfStream": bson.M{
			"$lt": date,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "StartOfStream", Value: -1}}).SetLimit(6)

	cursor, err := GoMongoDBCollStreamSummary.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var summaries []StreamSummarydomain.StreamSummary
	for cursor.Next(ctx) {
		var summary StreamSummarydomain.StreamSummary
		if err := cursor.Decode(&summary); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}
