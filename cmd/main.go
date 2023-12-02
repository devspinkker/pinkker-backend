package main

import (
	"PINKKER-BACKEND/config"
	cryptoroutes "PINKKER-BACKEND/internal/crypto/crypto-routes"
	donationroutes "PINKKER-BACKEND/internal/donation/donation-routes"
	streamroutes "PINKKER-BACKEND/internal/stream/stream-routes"
	tweetroutes "PINKKER-BACKEND/internal/tweet/tweet-routes"
	userroutes "PINKKER-BACKEND/internal/user/user-routes"
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	redisClient := setupRedisClient()
	newMongoDB := setupMongoDB()
	// defer redisClient.Close()
	defer newMongoDB.Disconnect(context.Background())

	app := fiber.New()
	app.Use(cors.New())

	// users
	userroutes.UserRoutes(app, redisClient, newMongoDB)
	// tweet
	tweetroutes.TweetdRoutes(app, redisClient, newMongoDB)
	// donation
	donationroutes.DonatioRoutes(app, redisClient, newMongoDB)
	// streams
	streamroutes.StreamsRoutes(app, redisClient, newMongoDB)
	// crypto
	cryptoroutes.CryptoRoutes(app, redisClient, newMongoDB)
	PORT := config.PORT()
	if PORT == "" {
		PORT = "8081"
	}
	log.Fatal(app.Listen(":" + PORT))
}

func setupRedisClient() *redis.Client {
	PasswordRedis := config.PASSWORDREDIS()
	ADDRREDIS := config.ADDRREDIS()
	client := redis.NewClient(&redis.Options{
		Addr:     ADDRREDIS,
		Password: PasswordRedis,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Error al conectar con Redis: %s", err.Error())
	}
	fmt.Println("Redis connect")
	return client
}
func setupMongoDB() *mongo.Client {
	URI := config.MONGODB_URI()
	if URI == "" {
		log.Fatal("MONGODB_URI FATAL")
	}

	clientOptions := options.Client().ApplyURI(URI)

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal("MONGODB ERROR", err.Error())
	}

	if err = client.Connect(context.Background()); err != nil {
		log.Fatal("MONGODB ERROR", err.Error())
	}

	if err = client.Ping(context.Background(), nil); err != nil {
		log.Fatal("MONGODB ERROR", err.Error())
	}
	return client
}
