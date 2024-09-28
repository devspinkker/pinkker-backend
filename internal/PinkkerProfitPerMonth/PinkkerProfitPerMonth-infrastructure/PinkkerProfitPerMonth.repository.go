package PinkkerProfitPerMonthinfrastructure

import (
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PinkkerProfitPerMonthRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewPinkkerProfitPerMonthRepository(redisClient *redis.Client, mongoClient *mongo.Client) *PinkkerProfitPerMonthRepository {
	return &PinkkerProfitPerMonthRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (r *PinkkerProfitPerMonthRepository) GetProfitByMonth(date time.Time) (PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("PinkkerProfitPerMonth")
	ctx := context.Background()

	startOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfNextMonth := time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": startOfMonth,
			"$lt":  startOfNextMonth,
		},
	}

	opts := options.Find()
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth{}, err
	}
	defer cursor.Close(ctx)

	var results PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth
	if err := cursor.All(ctx, &results); err != nil {
		return PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth{}, err
	}

	return results, nil
}

func (r *PinkkerProfitPerMonthRepository) GetProfitByMonthRange(startDate time.Time, endDate time.Time) ([]PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("PinkkerProfitPerMonth")
	ctx := context.Background()

	startOfRange := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfNextMonthAfterRange := time.Date(endDate.Year(), endDate.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": startOfRange,
			"$lt":  startOfNextMonthAfterRange,
		},
	}

	opts := options.Find()
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
func (r *PinkkerProfitPerMonthRepository) AutCode(id primitive.ObjectID, code string) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collectionUsers := db.Collection("Users")

	// Definir la proyección para obtener solo los campos necesarios
	projection := bson.M{
		"PanelAdminPinkker.Level": 1,
		"PanelAdminPinkker.Asset": 1,
		"PanelAdminPinkker.Code":  1,
	}

	var user struct {
		PanelAdminPinkker struct {
			Level int    `bson:"Level"`
			Asset bool   `bson:"Asset"`
			Code  string `bson:"Code"`
		} `bson:"PanelAdminPinkker"`
	}

	err := collectionUsers.FindOne(context.Background(), bson.M{"_id": id}, options.FindOne().SetProjection(projection)).Decode(&user)
	if err != nil {
		return err
	}

	// Comprobar las condiciones de autorización
	if user.PanelAdminPinkker.Level != 1 || !user.PanelAdminPinkker.Asset || user.PanelAdminPinkker.Code != code {
		return fmt.Errorf("usuario no autorizado")
	}

	return nil
}
