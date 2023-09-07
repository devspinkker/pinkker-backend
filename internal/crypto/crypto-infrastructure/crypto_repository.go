package cryptoinfrastructure

import (
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CryptoRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewCryptoRepository(redisClient *redis.Client, mongoClient *mongo.Client) *CryptoRepository {
	return &CryptoRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (r *CryptoRepository) TransferToken(client *ethclient.Client, signedTx string) (string, error) {
	// Decodificar la transacción firmada desde su representación en cadena
	txBytes, err := hex.DecodeString(signedTx)
	if err != nil {
		return "", err
	}

	// Deserializar la transacción firmada
	var tx types.Transaction
	rlp.DecodeBytes(txBytes, &tx)

	// Establece el límite de gas y precio de gas
	// gasLimit := uint64(21000)
	// gasPrice, err := client.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	return "", err
	// }

	// Enviar la transacción firmada al cliente de BSC
	err = client.SendTransaction(context.Background(), &tx)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}
func (r *CryptoRepository) UpdateSubscriptionState(SourceAddress string, DestinationAddress string) error {
	usersCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	ctx, cancel := context.WithTimeout(context.Background(), 30*24*time.Hour)
	defer cancel()

	sourceUser, destUser, err := r.findUsersByWallets(ctx, SourceAddress, DestinationAddress, usersCollection)
	if err != nil {
		return err
	}

	// Verificar si el usuario que recibe ya está suscrito
	var existingSubscription *userdomain.Subscription
	for _, subscription := range sourceUser.Subscriptions {
		if subscription.SubscriptionNameUser == destUser.NameUser {
			existingSubscription = &subscription
			break
		}
	}

	subscriptionStart := time.Now()
	subscriptionEnd := subscriptionStart.Add(30 * 24 * time.Hour)

	if existingSubscription == nil {
		// Si el usuario que recibe no está suscrito, agregarlo como suscriptor
		r.addSubscription(sourceUser, destUser, subscriptionStart, subscriptionEnd)
		r.addSubscriber(destUser, sourceUser, subscriptionEnd)
	} else {
		// Si el usuario que recibe ya está suscrito, actualizar la suscripción
		r.updateSubscription(existingSubscription, subscriptionStart, subscriptionEnd)
	}

	if err := r.updateUserSource(ctx, sourceUser, usersCollection); err != nil {
		return err
	}

	err = r.updateUserDest(ctx, destUser, usersCollection)
	return err
}

// Encuentra dos usuarios por sus direcciones de wallet
func (r *CryptoRepository) findUsersByWallets(ctx context.Context, sourceWallet, destWallet string, usersCollection *mongo.Collection) (*userdomain.User, *userdomain.User, error) {
	var sourceUser userdomain.User
	filtersourceWallet := bson.M{
		"Wallet": sourceWallet,
	}
	err := usersCollection.FindOne(ctx, filtersourceWallet).Decode(&sourceUser)
	if err != nil {
		return nil, nil, err
	}

	var destUser userdomain.User
	filterdestUserWallet := bson.M{
		"Wallet": destWallet,
	}

	err = usersCollection.FindOne(ctx, filterdestUserWallet).Decode(&destUser)
	if err != nil {
		return nil, nil, err
	}

	return &sourceUser, &destUser, nil
}

// Agrega un nuevo suscriptor al usuario que da
func (r *CryptoRepository) addSubscription(sourceUser *userdomain.User, destUser *userdomain.User, subscriptionStart, subscriptionEnd time.Time) {
	subscription := userdomain.Subscription{
		SubscriptionNameUser: destUser.NameUser,
		SubscriptionStart:    subscriptionStart,
		SubscriptionEnd:      subscriptionEnd,
		MonthsSubscribed:     1, // Comienza en 1 mes
	}
	sourceUser.Subscriptions = append(sourceUser.Subscriptions, subscription)
}

// Actualiza una suscripción existente
func (r *CryptoRepository) updateSubscription(subscription *userdomain.Subscription, subscriptionStart, subscriptionEnd time.Time) {
	subscription.SubscriptionStart = subscriptionStart
	subscription.SubscriptionEnd = subscriptionEnd
	// No es necesario actualizar MonthsSubscribed, ya que comenzamos desde 1
}

// Actualiza el usuario que da en MongoDB
func (r *CryptoRepository) updateUserSource(ctx context.Context, user *userdomain.User, usersCollection *mongo.Collection) error {
	filter := bson.M{"Wallet": user.Wallet}
	update := bson.M{"$set": bson.M{"Subscriptions": user.Subscriptions}}
	valor, err := usersCollection.UpdateOne(ctx, filter, update)
	fmt.Println(valor)
	return err
}

// Actualiza el usuario que destino en MongoDB
func (r *CryptoRepository) updateUserDest(ctx context.Context, user *userdomain.User, usersCollection *mongo.Collection) error {
	filter := bson.M{"Wallet": user.Wallet}
	update := bson.M{"$set": bson.M{"Subscribers": user.Subscribers}}
	valor, err := usersCollection.UpdateOne(ctx, filter, update)
	fmt.Println(valor)
	return err
}

// Agrega al usuario que da como suscriptor en la colección de subscriptores del usuario que recibe
func (r *CryptoRepository) addSubscriber(destUser *userdomain.User, sourceUser *userdomain.User, subscriptionEnd time.Time) {
	// Verificar si el usuario ya es un suscriptor
	existingSubscriber := false
	for _, subscriber := range destUser.Subscribers {
		if subscriber.SubscriberNameUser == sourceUser.NameUser {
			existingSubscriber = true
			break
		}
	}

	if !existingSubscriber {
		subscriber := userdomain.Subscriber{
			SubscriberNameUser: sourceUser.NameUser,
			SubscriptionEnd:    subscriptionEnd,
		}
		destUser.Subscribers = append(destUser.Subscribers, subscriber)
	}
}
