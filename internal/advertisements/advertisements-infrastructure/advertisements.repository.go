package advertisementsinfrastructure

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
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (r *WithdrawalsRepository) AcceptWithdrawal(id primitive.ObjectID, data withdrawalsdomain.AcceptWithdrawal) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollWithdrawals := db.Collection("WithdrawalRequests")
	ctx := context.TODO()

	filter := bson.M{
		"_id":   data.WithdrawalRequestsId,
		"State": "Pending",
	}
	var withdrawals withdrawalsdomain.WithdrawalRequests
	err := GoMongoDBCollWithdrawals.FindOne(context.Background(), filter).Decode(&withdrawals)
	if err != nil {
		return err
	}

	collectionUsers := db.Collection("Users")
	var UserRequest userdomain.User

	err = collectionUsers.FindOne(context.Background(), bson.M{"_id": withdrawals.RequestedBy}).Decode(&UserRequest)
	if err != nil {
		return err
	}

	if UserRequest.Pixeles < withdrawals.Amount {
		updateState := bson.M{
			"$set": bson.M{
				"State":      "rejected",
				"AcceptedBy": id,
				"TimeStamp":  time.Now(),
				"TextReturn": "falta de fondos, retiro rechazado",
			},
		}

		_, err = GoMongoDBCollWithdrawals.UpdateOne(ctx, filter, updateState)
		if err != nil {
			return err

		}
		return errors.New("falta de fondos, retiro rechazado")
	}

	updateWithdrawals := bson.D{
		{Key: "$inc", Value: bson.D{
			{Key: "Pixeles", Value: -(withdrawals.Amount)},
		}},
	}

	_, err = collectionUsers.UpdateOne(ctx, bson.M{"_id": withdrawals.RequestedBy}, updateWithdrawals)
	if err != nil {
		return err
	}

	updateState := bson.M{
		"$set": bson.M{
			"State":      "Accepted",
			"AcceptedBy": id,
			"TimeStamp":  time.Now(),
			"TextReturn": "Retiro aceptado",
		},
	}

	_, err = GoMongoDBCollWithdrawals.UpdateOne(ctx, filter, updateState)
	if err != nil {
		return err
	}

	return nil
}
func (r *WithdrawalsRepository) RejectWithdrawal(id primitive.ObjectID, data withdrawalsdomain.RejectWithdrawal) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollWithdrawals := db.Collection("WithdrawalRequests")
	ctx := context.TODO()

	filter := bson.M{
		"_id":   data.WithdrawalRequestsId,
		"State": "Pending",
	}
	var withdrawals withdrawalsdomain.WithdrawalRequests
	err := GoMongoDBCollWithdrawals.FindOne(ctx, filter).Decode(&withdrawals)
	if err != nil {
		return err
	}

	updateState := bson.M{
		"$set": bson.M{
			"State":      "rejected",
			"AcceptedBy": id,
			"TimeStamp":  time.Now(),
			"TextReturn": data.TextReturn,
		},
	}

	_, err = GoMongoDBCollWithdrawals.UpdateOne(ctx, filter, updateState)
	if err != nil {
		return err
	}

	return nil
}

func (r *WithdrawalsRepository) GetWithdrawalToken(id primitive.ObjectID) ([]withdrawalsdomain.WithdrawalRequests, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollWithdrawals := db.Collection("WithdrawalRequests")
	ctx := context.TODO()

	filter := bson.M{
		"RequestedBy": id,
	}

	findOptions := options.Find()
	sort := bson.D{{Key: "TimeStamp", Value: -1}}
	findOptions.SetSort(sort)

	cursor, err := GoMongoDBCollWithdrawals.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var withdrawals []withdrawalsdomain.WithdrawalRequests
	for cursor.Next(ctx) {
		var withdrawal withdrawalsdomain.WithdrawalRequests
		if err := cursor.Decode(&withdrawal); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}
