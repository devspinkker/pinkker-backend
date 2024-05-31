package withdrawalstinfrastructure

import (
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	withdrawalsdomain "PINKKER-BACKEND/internal/withdraw/withdraw"
	"context"
	"errors"
	"fmt"
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
func (r *WithdrawalsRepository) WithdrawalRequest(id primitive.ObjectID, nameUser string, data withdrawalsdomain.WithdrawalRequestReq) error {
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

	if user.Pixeles < amount || user.Pixeles < 1000 {
		return errors.New("insufficient pixels")
	}
	var existingRequest withdrawalsdomain.WithdrawalRequests
	err = GoMongoDBCollWithdrawals.FindOne(context.Background(), bson.M{
		"RequestedBy": id,
		"State":       "Pending",
	}).Decode(&existingRequest)
	if err == nil {
		return errors.New("there is already a pending withdrawal request")
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	CreateWithdrawalRequest := withdrawalsdomain.WithdrawalRequests{
		ID:               primitive.NewObjectID(),
		Destination:      data.Cbu,
		AcceptedBy:       primitive.NilObjectID,
		RequestedBy:      id,
		RequesteNameUser: nameUser,
		Amount:           amount,
		TimeStamp:        time.Now(),
		Notified:         false,
		State:            "Pending",
		TextReturn:       "",
	}

	_, err = GoMongoDBCollWithdrawals.InsertOne(context.Background(), CreateWithdrawalRequest)
	if err != nil {
		return err
	}

	return nil
}
func (r *WithdrawalsRepository) GetPendingUnnotifiedWithdrawals(data withdrawalsdomain.WithdrawalRequestGet) ([]withdrawalsdomain.WithdrawalRequests, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollWithdrawals := db.Collection("WithdrawalRequests")

	filter := bson.M{
		"State":    "Pending",
		"Notified": false,
	}

	cursor, err := GoMongoDBCollWithdrawals.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var withdrawalRequests []withdrawalsdomain.WithdrawalRequests
	for cursor.Next(context.Background()) {
		var request withdrawalsdomain.WithdrawalRequests
		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}
		withdrawalRequests = append(withdrawalRequests, request)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return withdrawalRequests, nil
}
func (r *WithdrawalsRepository) AutCode(id primitive.ObjectID, code string) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
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
