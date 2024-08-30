package advertisementsinfrastructure

import (
	"PINKKER-BACKEND/config"
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"fmt"
	"log"
	"strconv"
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

	currentDate := time.Now().Format("2006-01-02")

	filter := bson.M{
		"_id":                    idAdvertisements,
		"IdOfTheUsersWhoClicked": bson.M{"$ne": IdU},
	}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}
	// si el count es 0 signnifica que ese anuncio ya lo vio ese usuario por lo que hay que retornar
	if count == 0 {
		return nil
	}
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
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	err = r.updatePinkkerProfitPerMonth(ctx)
	if err != nil {
		return err
	}
	updatePerDay := bson.M{
		"$inc": bson.M{
			"ClicksPerDay.$[elem].Clicks": 1,
		},
	}

	// Filtro del array para encontrar la fecha actual
	arrayFilter := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.Date": currentDate},
		},
	}

	// Intentar actualizar el conteo de clics para la fecha actual
	result, err := collection.UpdateOne(ctx, bson.M{
		"_id": idAdvertisements,
	}, updatePerDay, options.Update().SetArrayFilters(arrayFilter))
	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		newDateUpdate := bson.M{
			"$addToSet": bson.M{
				"ClicksPerDay": bson.M{
					"Date":   currentDate,
					"Clicks": 1,
				},
			},
		}

		_, err = collection.UpdateOne(ctx, bson.M{"_id": idAdvertisements}, newDateUpdate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *AdvertisementsRepository) updatePinkkerProfitPerMonth(ctx context.Context) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollMonthly := GoMongoDB.Collection("PinkkerProfitPerMonth")
	AdvertisementsPayPerPrint := config.AdvertisementsPayClicks()
	AdvertisementsClicks, err := strconv.ParseFloat(AdvertisementsPayPerPrint, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor")
	}
	currentTime := time.Now()
	currentMonth := int(currentTime.Month())
	currentYear := currentTime.Year()
	currentWeek := getWeekOfMonth(currentTime)

	startOfMonth := time.Date(currentYear, time.Month(currentMonth), 1, 0, 0, 0, 0, time.UTC)
	startOfNextMonth := time.Date(currentYear, time.Month(currentMonth+1), 1, 0, 0, 0, 0, time.UTC)

	monthlyFilter := bson.M{
		"timestamp": bson.M{
			"$gte": startOfMonth,
			"$lt":  startOfNextMonth,
		},
	}

	defaultWeek := PinkkerProfitPerMonthdomain.Week{
		Impressions: 0,
		Clicks:      0,
		Pixels:      0.0,
	}

	monthlyUpdate := bson.M{
		"$inc": bson.M{
			"total":                            AdvertisementsClicks,
			"weeks." + currentWeek + ".clicks": AdvertisementsClicks,
		},
		"$setOnInsert": bson.M{
			"timestamp": currentTime,
			"weeks":     map[string]PinkkerProfitPerMonthdomain.Week{currentWeek: defaultWeek},
		},
	}

	// Set the option to upsert, creating a new document if one doesn't exist
	monthlyOpts := options.Update().SetUpsert(true)
	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdate, monthlyOpts)
	if err != nil {
		return err
	}

	return nil
}

func getWeekOfMonth(t time.Time) string {
	startOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	dayOfMonth := t.Day()
	dayOfWeek := int(startOfMonth.Weekday())
	weekNumber := (dayOfMonth+dayOfWeek-1)/7 + 1
	return "week_" + strconv.Itoa(weekNumber)
}

func (r *AdvertisementsRepository) AdvertisementsGet(page int64) ([]advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	findOptions := options.Find()
	findOptions.SetLimit(6)
	findOptions.SetSkip((page - 1) * 6)

	cursor, err := Advertisements.Find(ctx, bson.M{}, findOptions)
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

func (r *AdvertisementsRepository) GetAdsUser(NameUser string) ([]advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	filter := bson.M{"NameUser": NameUser}

	cursor, err := Advertisements.Find(ctx, filter)
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
	AdvertisementsPayPerPrint := config.AdvertisementsPayPerPrint()
	floatValue, err := strconv.ParseFloat(AdvertisementsPayPerPrint, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor")
	}
	var documento advertisements.Advertisements
	documento.Name = ad.Name
	documento.NameUser = ad.NameUser
	documento.Destination = ad.Destination
	documento.Categorie = ad.Categorie
	documento.UrlVideo = ad.UrlVideo
	documento.ReferenceLink = ad.ReferenceLink
	documento.ImpressionsMax = ad.ImpressionsMax
	documento.PayPerPrint = floatValue
	documento.Impressions = 0
	documento.ClicksMax = ad.ClicksMax
	documento.DocumentToBeAnnounced = ad.DocumentToBeAnnounced
	documento.IdOfTheUsersWhoClicked = []primitive.ObjectID{}
	documento.ClicksPerDay = []advertisements.ClicksPerDay{}
	documento.ImpressionsPerDay = []advertisements.ImpressionsPerDay{}
	documento.Timestamp = time.Now()

	_, err = collection.InsertOne(context.Background(), documento)
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
