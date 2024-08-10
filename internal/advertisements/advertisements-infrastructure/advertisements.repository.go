package advertisementsinfrastructure

import (
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AdvertisementsRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewadvertisementsRepository(redisClient *redis.Client, mongoClient *mongo.Client) *AdvertisementsRepository {
	return &AdvertisementsRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (r *AdvertisementsRepository) AdvertisementsGet() ([]advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	findOptions := options.Find()

	cursor, err := Advertisements.Find(ctx, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var advertisementsArray []advertisements.Advertisements
	for cursor.Next(ctx) {
		var advertisement advertisements.Advertisements
		if err := cursor.Decode(&advertisement); err != nil {
			return nil, err
		}
		advertisementsArray = append(advertisementsArray, advertisement)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return advertisementsArray, nil

}
func (r *AdvertisementsRepository) AutCode(id primitive.ObjectID, code string) error {
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
func (r *AdvertisementsRepository) CreateAdvertisement(ad advertisements.UpdateAdvertisement) (advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Advertisements")

	var documento advertisements.Advertisements
	documento.Name = ad.Name
	documento.Destination = ad.Destination
	documento.Categorie = ad.Categorie
	documento.UrlVideo = ad.UrlVideo
	documento.ReferenceLink = ad.ReferenceLink
	documento.ImpressionsMax = ad.ImpressionsMax
	documento.PayPerPrint = 10
	documento.Impressions = 0
	documento.ClicksMax = ad.ClicksMax
	documento.DocumentToBeAnnounced = ad.DocumentToBeAnnounced
	documento.IdOfTheUsersWhoClicked = []primitive.ObjectID{}

	_, err := collection.InsertOne(context.Background(), documento)
	if err != nil {
		return advertisements.Advertisements{}, err
	}

	return documento, nil
}
func (r *AdvertisementsRepository) UpdateAdvertisement(ad advertisements.UpdateAdvertisement) (advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Advertisements")

	filter := bson.M{"_id": ad.ID}
	update := bson.M{
		"$set": bson.M{
			"Name":                  ad.Name,
			"Destination":           ad.Destination,
			"Categorie":             ad.Categorie,
			"Impressions":           0,
			"UrlVideo":              ad.UrlVideo,
			"ReferenceLink":         ad.ReferenceLink,
			"ClicksMax":             ad.ClicksMax,
			"DocumentToBeAnnounced": ad.DocumentToBeAnnounced,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedAd advertisements.Advertisements
	err := collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedAd)
	if err != nil {
		return advertisements.Advertisements{}, err
	}

	return updatedAd, nil
}

func (r *AdvertisementsRepository) IdOfTheUsersWhoClicked(IdU primitive.ObjectID, idAdvertisements primitive.ObjectID) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Advertisements")
	ctx := context.TODO()

	filter := bson.M{
		"_id":                    idAdvertisements,
		"IdOfTheUsersWhoClicked": bson.M{"$ne": IdU},
	}

	update := bson.M{
		"$addToSet": bson.M{
			"IdOfTheUsersWhoClicked": IdU,
		},
		"$inc": bson.M{
			"Clicks": 1,
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *AdvertisementsRepository) DeleteAdvertisement(id primitive.ObjectID) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Advertisements")

	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(context.Background(), filter)
	return err
}
