package PinkkerProfitPerMonthinfrastructure

import (
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
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

func (u *PinkkerProfitPerMonthRepository) AutCode(id primitive.ObjectID, code string) error {
	db := u.mongoClient.Database("PINKKER-BACKEND")
	collectionUsers := db.Collection("Users")
	var User userdomain.User

	err := collectionUsers.FindOne(context.Background(), bson.M{"_id": id}).Decode(&User)
	if err != nil {
		return err
	}

	if User.PanelAdminPinkker.Level != 1 || !User.PanelAdminPinkker.Asset || User.PanelAdminPinkker.Code != code {
		return fmt.Errorf("usuario no autorizado")
	}
	return nil
}
