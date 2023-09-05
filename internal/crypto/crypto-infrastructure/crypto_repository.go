package cryptoinfrastructure

import (
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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
func (r *CryptoRepository) TransferBNB(client *ethclient.Client, sourceAddress, destinationAddress string, amount *big.Int) (string, error) {
	// Convierte las direcciones en tipos comunes
	source := common.HexToAddress(sourceAddress)
	destination := common.HexToAddress(destinationAddress)

	// Obtiene el nonce del origen (número de transacción)
	nonce, err := client.PendingNonceAt(context.Background(), source)
	if err != nil {
		return "", err
	}
	// Establece la gas price y la cantidad máxima de gas
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	// Establece el límite de gas
	gasLimit := uint64(21000) // Gas limit para transferencias

	// Crea la transacción
	tx := types.NewTransaction(nonce, destination, amount, gasLimit, gasPrice, nil)

	// Firma la transacción
	chainID := big.NewInt(56) // Para Binance Smart Chain (BSC)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), nil)
	if err != nil {
		return "", err
	}

	// Envía la transacción
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	// Devuelve el hash de la transacción
	return signedTx.Hash().Hex(), nil
}
func (r *CryptoRepository) UpdateSubscriptionState(SourceAddress string, DestinationAddress string) error {
	// Crear una variable para el contexto con un temporizador de un mes
	ctx, cancel := context.WithTimeout(context.Background(), 30*24*time.Hour)
	defer cancel()

	// Buscar los dos usuarios en la colección de MongoDB
	sourceUser, destUser, err := r.findUsersByWallets(ctx, SourceAddress, DestinationAddress)
	if err != nil {
		return err
	}

	// Verificar si el usuario que recibe ya está suscrito al usuario que da
	alreadySubscribed := false
	subscriptionIndex := -1

	for i, subscription := range sourceUser.Subscriptions {
		if subscription.SubscriberID == destUser.ID {
			alreadySubscribed = true
			subscriptionIndex = i
			break
		}
	}

	// Si el usuario que recibe no está suscrito, agregarlo como suscriptor
	if !alreadySubscribed {
		r.addSubscription(sourceUser, destUser)
	} else {
		r.updateSubscription(sourceUser, subscriptionIndex)
	}

	// Actualizar el usuario que da en MongoDB
	if err := r.updateUser(ctx, sourceUser); err != nil {
		return err
	}

	// Agregar al usuario que da como suscriptor en la colección de subscriptores del usuario que recibe
	r.addSubscriber(destUser, sourceUser)

	return nil
}

// Encuentra dos usuarios por sus direcciones de wallet
func (r *CryptoRepository) findUsersByWallets(ctx context.Context, sourceWallet, destWallet string) (*userdomain.User, *userdomain.User, error) {
	usersCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("tu_coleccion_de_usuarios")

	filter := bson.M{
		"wallet": bson.M{"$in": []string{sourceWallet, destWallet}},
	}

	cur, err := usersCollection.Find(ctx, filter)
	if err != nil {
		return nil, nil, err
	}
	defer cur.Close(ctx)

	var sourceUser userdomain.User
	var destUser userdomain.User

	for cur.Next(ctx) {
		var user userdomain.User
		if err := cur.Decode(&user); err != nil {
			return nil, nil, err
		}

		if user.Wallet == sourceWallet {
			sourceUser = user
		} else if user.Wallet == destWallet {
			destUser = user
		}
	}

	if err := cur.Err(); err != nil {
		return nil, nil, err
	}

	return &sourceUser, &destUser, nil
}

// Agrega un nuevo suscriptor al usuario que da
func (r *CryptoRepository) addSubscription(sourceUser *userdomain.User, destUser *userdomain.User) {
	subscription := userdomain.Subscription{
		SubscriberID:      destUser.ID,
		SubscriptionStart: time.Now(),
		SubscriptionEnd:   time.Now().Add(30 * 24 * time.Hour),
	}
	sourceUser.Subscriptions = append(sourceUser.Subscriptions, subscription)
}

// Actualiza una suscripción existente
func (r *CryptoRepository) updateSubscription(sourceUser *userdomain.User, index int) {
	sourceUser.Subscriptions[index].SubscriptionEnd = time.Now().Add(30 * 24 * time.Hour)
	sourceUser.Subscriptions[index].MonthsSubscribed = int(sourceUser.Subscriptions[index].SubscriptionEnd.Sub(sourceUser.Subscriptions[index].SubscriptionStart).Hours() / 24 / 30)
}

// Actualiza el usuario que da en MongoDB
func (r *CryptoRepository) updateUser(ctx context.Context, user *userdomain.User) error {
	usersCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("tu_coleccion_de_usuarios")

	filter := bson.M{"wallet": user.Wallet}
	update := bson.M{"$set": bson.M{"subscriptions": user.Subscriptions}}

	_, err := usersCollection.UpdateOne(ctx, filter, update)
	return err
}

// Agrega al usuario que da como suscriptor en la colección de subscriptores del usuario que recibe
func (r *CryptoRepository) addSubscriber(destUser *userdomain.User, sourceUser *userdomain.User) {
	// Crear una estructura para el suscriptor
	subscriber := userdomain.Subscriber{
		SubscriberID:    sourceUser.ID,
		SubscriptionEnd: sourceUser.Subscriptions[0].SubscriptionEnd, // Usar la fecha de finalización de la suscripción del usuario que da
	}

	// Agregar al usuario que da como suscriptor en la colección de subscriptores del usuario que recibe
	destUser.Subscribers = append(destUser.Subscribers, subscriber)
}
