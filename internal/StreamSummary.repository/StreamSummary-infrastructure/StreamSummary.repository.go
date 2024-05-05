package StreamSummaryinfrastructure

import (
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary.repository/StreamSummary-domain"
	"context"

	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
