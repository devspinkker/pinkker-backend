package advertisementsinfrastructure

import (
	"PINKKER-BACKEND/internal/advertisements/advertisements"
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
func (r *AdvertisementsRepository) IdOfTheUsersWhoClicked(IdU primitive.ObjectID, idAdvertisements primitive.ObjectID) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Advertisements")
	ctx := context.TODO()

	// Obtener la fecha actual en formato de cadena (YYYY-MM-DD)
	currentDate := time.Now().Format("2006-01-02")

	// Filtro para encontrar el documento relevante
	filter := bson.M{
		"_id": idAdvertisements,
		// "IdOfTheUsersWhoClicked": bson.M{"$ne": IdU},
	}

	// Actualización para añadir el ID y manejar la longitud de la lista, y agregar clics por día
	update := bson.M{
		"$push": bson.M{
			"IdOfTheUsersWhoClicked": bson.M{
				"$each":     []primitive.ObjectID{IdU},
				"$position": -1,
				"$slice":    -50,
			},
		},
		"$inc": bson.M{
			"Clicks": 1,
		},
		"$setOnInsert": bson.M{
			"": []bson.M{
				{"Date": currentDate, "Clicks": 1},
			},
		},
	}

	// Opciones para definir el arrayFilters
	options := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"element.Date": currentDate},
		},
	}).SetUpsert(true) // Upsert para insertar si no existe

	// Ejecutar la actualización
	result, err := collection.UpdateOne(ctx, filter, update, options)
	if err != nil {
		return err
	}

	// Si no se actualizó ningún documento, significa que no existe un registro para la fecha actual
	if result.ModifiedCount == 0 {
		// Insertar un nuevo objeto para la fecha actual
		newDateUpdate := bson.M{
			"$push": bson.M{
				"ClicksPerDay": bson.M{
					"Date":   currentDate,
					"Clicks": 1,
				},
			},
			"$inc": bson.M{
				"Clicks": 1,
			},
		}

		_, err = collection.UpdateOne(ctx, filter, newDateUpdate)
		if err != nil {
			return err
		}
	}

	return nil
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
	documento.ClicksPerDay = []advertisements.ClicksPerDay{}

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

func (r *AdvertisementsRepository) DeleteAdvertisement(id primitive.ObjectID) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Advertisements")

	filter := bson.M{"_id": id}
	_, err := collection.DeleteOne(context.Background(), filter)
	return err
}
