package withdrawalstinfrastructure

import (
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	withdrawalsdomain "PINKKER-BACKEND/internal/withdraw/withdraw"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type WithdrawalsRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewwithdrawalsRepository(redisClient *redis.Client, mongoClient *mongo.Client) *WithdrawalsRepository {
	return &WithdrawalsRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (r *WithdrawalsRepository) WithdrawalRequest(id primitive.ObjectID, data withdrawalsdomain.WithdrawalRequestReq) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := db.Collection("Users")
	GoMongoDBCollWithdrawals := db.Collection("WithdrawalRequests")

	amount, err := strconv.ParseFloat(data.Amount, 64)
	if err != nil {
		return err
	}

	var user userdomain.User
	err = GoMongoDBCollUsers.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return err
	}

	if user.Pixeles < amount || user.Pixeles > 1000 {
		return errors.New("insufficient pixels")
	}

	CreateWithdrawalRequest := withdrawalsdomain.WithdrawalRequests{
		ID:          primitive.NewObjectID(),
		Destination: data.Cbu,
		AcceptedBy:  primitive.NilObjectID,
		RequestedBy: id,
		Amount:      amount,
		TimeStamp:   time.Now(),
		Notified:    false,
		State:       "Pending",
		TextReturn:  "",
	}

	_, err = GoMongoDBCollWithdrawals.InsertOne(context.Background(), CreateWithdrawalRequest)
	if err != nil {
		return err
	}

	return nil
}
