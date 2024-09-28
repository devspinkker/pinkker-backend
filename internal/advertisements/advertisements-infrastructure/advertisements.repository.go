package advertisementsinfrastructure

import (
	"PINKKER-BACKEND/config"
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	"context"
	"errors"
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
		log.Fatalf("error al convertir el valor: %v", err)
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

	// Paso 1: Inserta el documento si no existe con la estructura básica
	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, bson.M{
		"$setOnInsert": bson.M{
			"timestamp": currentTime,
			"weeks." + currentWeek: PinkkerProfitPerMonthdomain.Week{
				Impressions: 0,
				Clicks:      0,
				Pixels:      0.0,
			},
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	monthlyUpdate := bson.M{
		"$inc": bson.M{
			"total":                            AdvertisementsClicks,
			"weeks." + currentWeek + ".clicks": AdvertisementsClicks,
		},
	}

	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdate)
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

	if weekNumber > 4 {
		weekNumber = 4
	}

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

	// Definir la proyección para obtener solo los campos necesarios
	projection := bson.M{
		"PanelAdminPinkker.Level": 1,
		"PanelAdminPinkker.Asset": 1,
		"PanelAdminPinkker.Code":  1,
	}

	var user struct {
		PanelAdminPinkker struct {
			Level int    `bson:"Level"`
			Asset bool   `bson:"Asset"`
			Code  string `bson:"Code"`
		} `bson:"PanelAdminPinkker"`
	}

	err := collectionUsers.FindOne(context.Background(), bson.M{"_id": id}, options.FindOne().SetProjection(projection)).Decode(&user)
	if err != nil {
		return err
	}

	// Comprobar las condiciones de autorización
	if user.PanelAdminPinkker.Level != 1 || !user.PanelAdminPinkker.Asset || user.PanelAdminPinkker.Code != code {
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

func (r *AdvertisementsRepository) BuyadCreate(ad advertisements.UpdateAdvertisement, idUser primitive.ObjectID) (advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Advertisements")
	collectionUsers := db.Collection("Users")

	ctx := context.Background()
	AdvertisementsPayPerPrint := config.AdvertisementsPayPerPrint()
	floatValuePayPerPrint, err := strconv.ParseFloat(AdvertisementsPayPerPrint, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor")
	}
	floatValuePayPerPrint *= 2
	AdvertisementsPayClicks := config.AdvertisementsPayClicks()
	floatValuePayClicks, err := strconv.ParseFloat(AdvertisementsPayClicks, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor")
	}

	// Buscar solo los Pixeles del usuario
	userPixeles, err := r.findPixelesByUserId(ctx, idUser, collectionUsers)
	if err != nil {
		log.Fatalf("error al encontrar los pixeles del usuario")
		return advertisements.Advertisements{}, err
	}

	// Calcular los pixeles necesarios
	var PixelesUserNeed float64
	if ad.Destination == "Muro" {
		PixelesUserNeed = floatValuePayClicks * float64(ad.ClicksMax)
	} else if ad.Destination == "Streams" {
		PixelesUserNeed = floatValuePayPerPrint * float64(ad.ImpressionsMax)
	} else {
		return advertisements.Advertisements{}, errors.New("destination undefined")
	}

	// Verificar si el usuario tiene suficientes pixeles
	if PixelesUserNeed > userPixeles || PixelesUserNeed <= 50000 {
		return advertisements.Advertisements{}, errors.New("insufficient Pixeles")
	}

	// Restar los pixeles del usuario
	userPixeles -= PixelesUserNeed

	// Actualizar la cantidad de pixeles del usuario en la base de datos
	update := bson.M{
		"$inc": bson.M{
			"Pixeles": -PixelesUserNeed,
		},
	}
	_, err = collectionUsers.UpdateOne(ctx, bson.M{"_id": idUser}, update)
	if err != nil {
		return advertisements.Advertisements{}, err
	}

	// Crear el documento del anuncio
	var documento advertisements.Advertisements
	documento.Name = ad.Name
	documento.NameUser = ad.NameUser
	documento.Destination = ad.Destination
	documento.Categorie = ad.Categorie
	documento.UrlVideo = ad.UrlVideo
	documento.ReferenceLink = ad.ReferenceLink
	documento.ImpressionsMax = ad.ImpressionsMax
	documento.PayPerPrint = floatValuePayPerPrint
	documento.Impressions = 0
	documento.ClicksMax = ad.ClicksMax
	documento.DocumentToBeAnnounced = ad.DocumentToBeAnnounced
	documento.IdOfTheUsersWhoClicked = []primitive.ObjectID{}
	documento.ClicksPerDay = []advertisements.ClicksPerDay{}
	documento.ImpressionsPerDay = []advertisements.ImpressionsPerDay{}
	documento.Timestamp = time.Now()
	documento.State = "pending"

	// Insertar el anuncio en la base de datos
	_, err = collection.InsertOne(ctx, documento)
	if err != nil {
		return advertisements.Advertisements{}, err
	}

	// Devolver el anuncio creado
	return documento, nil
}
func (r *AdvertisementsRepository) findPixelesByUserId(ctx context.Context, source_id primitive.ObjectID, usersCollection *mongo.Collection) (float64, error) {
	// Proyección para solo obtener el campo 'Pixeles'
	projection := bson.M{
		"Pixeles": 1, // Solo queremos el campo 'Pixeles'
	}

	var result struct {
		Pixeles float64 `bson:"Pixeles"` // Definimos solo el campo que queremos
	}

	// Filtro para buscar por ID
	filter := bson.M{
		"_id": source_id,
	}

	// Ejecutar la consulta con la proyección
	err := usersCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		return 0, err
	}

	// Retornar solo los Pixeles
	return result.Pixeles, nil
}
func (u *AdvertisementsRepository) GetTOTPSecret(ctx context.Context, userID primitive.ObjectID) (string, error) {
	usersCollection := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	filter := bson.M{"_id": userID}

	projection := bson.M{"TOTPSecret": 1, "_id": 0}

	var result struct {
		TOTPSecret string `bson:"TOTPSecret"`
	}

	err := usersCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.TOTPSecret, nil
}

func (r *AdvertisementsRepository) AcceptPendingAdByID(adID string) (int64, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	// Convertir el ID de string a ObjectID
	id, err := primitive.ObjectIDFromHex(adID)
	if err != nil {
		return 0, err
	}

	filter := bson.M{
		"_id":   id,
		"State": "pending",
	}

	update := bson.M{
		"$set": bson.M{
			"State": "accepted",
		},
	}

	// Ejecutar la actualización y obtener el número de documentos modificados
	result, err := Advertisements.UpdateOne(ctx, filter, update)
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}
func (r *AdvertisementsRepository) AcceptPendingAds(NameUser string) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	// Filtro para seleccionar anuncios pendientes para el usuario especificado
	filter := bson.M{
		"NameUser": NameUser,
		"State":    "pending",
	}

	// Actualización para cambiar el estado a aceptado
	update := bson.M{
		"$set": bson.M{
			"State": "accepted",
		},
	}

	// Ejecutar la actualización y obtener el número de documentos modificados
	_, err := Advertisements.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (r *AdvertisementsRepository) RemovePendingAds(NameUser string) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Advertisements := db.Collection("Advertisements")
	Users := db.Collection("Users")

	ctx := context.TODO()

	// Obtener los anuncios pendientes para el usuario
	filter := bson.M{
		"NameUser": NameUser,
		"State":    "pending",
	}

	cursor, err := Advertisements.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var ads []advertisements.Advertisements
	for cursor.Next(ctx) {
		var ad advertisements.Advertisements
		if err := cursor.Decode(&ad); err != nil {
			return err
		}
		ads = append(ads, ad)
	}

	if err := cursor.Err(); err != nil {
		return err
	}
	AdvertisementsPayPerPrint := config.AdvertisementsPayPerPrint()
	floatValuePayPerPrint, err := strconv.ParseFloat(AdvertisementsPayPerPrint, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor")
	}
	floatValuePayPerPrint *= 2
	AdvertisementsPayClicks := config.AdvertisementsPayClicks()
	floatValuePayClicks, err := strconv.ParseFloat(AdvertisementsPayClicks, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor")
	}

	var totalPixeles float64
	for _, ad := range ads {
		var pixelesUserNeed float64
		if ad.Destination == "Muro" || ad.Destination == "ClipAds" {
			pixelesUserNeed = float64(ad.ClicksMax) * float64(floatValuePayClicks)
		} else if ad.Destination == "Streams" {
			pixelesUserNeed = float64(ad.ImpressionsMax) * float64(floatValuePayPerPrint)
		} else {
			return errors.New("destination undefined")
		}

		totalPixeles += pixelesUserNeed
	}

	// Eliminar los anuncios pendientes
	_, err = Advertisements.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	update := bson.M{
		"$inc": bson.M{
			"Pixeles": totalPixeles,
		},
	}
	_, err = Users.UpdateOne(ctx, bson.M{"NameUser": NameUser}, update)
	if err != nil {
		return err
	}

	return nil
}
func (r *AdvertisementsRepository) GetAllPendingAds(page int64) ([]advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	limit := int64(10)
	skip := int64((page - 1) * 10)

	filter := bson.M{
		"State": "pending",
	}

	options := options.Find()
	options.SetLimit(limit)
	options.SetSkip(skip)

	cursor, err := Advertisements.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pendingAds []advertisements.Advertisements
	for cursor.Next(ctx) {
		var ad advertisements.Advertisements
		if err := cursor.Decode(&ad); err != nil {
			return nil, err
		}
		pendingAds = append(pendingAds, ad)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return pendingAds, nil
}
func (r *AdvertisementsRepository) GetAllPendingNameUserAds(page int64, nameuser string) ([]advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	Advertisements := db.Collection("Advertisements")

	ctx := context.TODO()

	limit := int64(10)
	skip := int64((page - 1) * 10)

	filter := bson.M{
		"State":    "pending",
		"NameUser": nameuser,
	}

	options := options.Find()
	options.SetLimit(limit)
	options.SetSkip(skip)

	cursor, err := Advertisements.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pendingAds []advertisements.Advertisements
	for cursor.Next(ctx) {
		var ad advertisements.Advertisements
		if err := cursor.Decode(&ad); err != nil {
			return nil, err
		}
		pendingAds = append(pendingAds, ad)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return pendingAds, nil
}

func (r *AdvertisementsRepository) CreateAdsAdvertisement(data advertisements.ClipAdsCreate) (advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Advertisements")
	AdvertisementsPayPerPrint := config.AdvertisementsPayPerPrint()
	floatValue, err := strconv.ParseFloat(AdvertisementsPayPerPrint, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor")
	}
	var documento advertisements.Advertisements
	documento.Name = data.Name
	documento.NameUser = data.NameUser
	documento.Destination = data.Destination
	documento.Categorie = data.Categorie
	documento.UrlVideo = data.UrlVideo
	documento.ReferenceLink = data.ReferenceLink
	documento.ImpressionsMax = data.ImpressionsMax
	documento.PayPerPrint = floatValue
	documento.Impressions = 0
	documento.ClicksMax = data.ClicksMax
	documento.DocumentToBeAnnounced = data.DocumentToBeAnnounced
	documento.IdOfTheUsersWhoClicked = []primitive.ObjectID{}
	documento.ClicksPerDay = []advertisements.ClicksPerDay{}
	documento.ImpressionsPerDay = []advertisements.ImpressionsPerDay{}
	documento.Timestamp = time.Now()
	documento.ClipId = primitive.ObjectID{}

	_, err = collection.InsertOne(context.Background(), documento)
	if err != nil {
		return advertisements.Advertisements{}, err
	}

	return documento, nil
}
func (c *AdvertisementsRepository) FindUser(id primitive.ObjectID) (string, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	FindUserInDb := bson.D{
		{Key: "_id", Value: id},
	}

	projection := bson.M{
		"Avatar": 1,
	}

	var result struct {
		Avatar string `bson:"Avatar"`
	}

	// Realizar la consulta con proyección
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb, options.FindOne().SetProjection(projection)).Decode(&result)
	if errCollUsers != nil {
		return "", errCollUsers
	}

	return result.Avatar, nil
}

func (c *AdvertisementsRepository) SaveClip(clip *clipdomain.Clip) (primitive.ObjectID, error) {
	database := c.mongoClient.Database("PINKKER-BACKEND")
	clipCollection := database.Collection("Clips")

	result, err := clipCollection.InsertOne(context.Background(), clip)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return primitive.ObjectID{}, errors.New("no se pudo obtener el ID insertado")
	}
	return insertedID, err
}
func (c *AdvertisementsRepository) UpdateClip(clipID primitive.ObjectID, newURL string, ad primitive.ObjectID) {
	clipCollection := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	filter := bson.M{"_id": clipID}

	update := bson.M{"$set": bson.M{"url": newURL, "AdId": ad}}

	opts := options.Update().SetUpsert(false)

	clipCollection.UpdateOne(context.Background(), filter, update, opts)
}

func (r *AdvertisementsRepository) BuyadClipCreate(ad advertisements.ClipAdsCreate, idUser, clipid primitive.ObjectID) (advertisements.Advertisements, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Advertisements")
	collectionUsers := db.Collection("Users")

	ctx := context.Background()
	AdvertisementsPayPerPrint := config.AdvertisementsPayPerPrint()
	floatValuePayPerPrint, err := strconv.ParseFloat(AdvertisementsPayPerPrint, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor")
	}
	floatValuePayPerPrint *= 2
	AdvertisementsPayClicks := config.AdvertisementsPayClicks()
	floatValuePayClicks, err := strconv.ParseFloat(AdvertisementsPayClicks, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor")
	}

	// Buscar solo los Pixeles del usuario
	userPixeles, err := r.findPixelesByUserId(ctx, idUser, collectionUsers)
	if err != nil {
		log.Fatalf("error al encontrar los pixeles del usuario")
		return advertisements.Advertisements{}, err
	}

	// Calcular los pixeles necesarios
	var PixelesUserNeed float64
	if ad.Destination == "ClipAds" {
		PixelesUserNeed = floatValuePayClicks * float64(ad.ClicksMax)
	} else {
		return advertisements.Advertisements{}, errors.New("destination undefined")
	}

	// Verificar si el usuario tiene suficientes pixeles
	if PixelesUserNeed > userPixeles || PixelesUserNeed <= 50000 {
		return advertisements.Advertisements{}, errors.New("insufficient Pixeles")
	}

	// Restar los pixeles del usuario
	userPixeles -= PixelesUserNeed

	// Actualizar la cantidad de pixeles del usuario en la base de datos
	update := bson.M{
		"$inc": bson.M{
			"Pixeles": -PixelesUserNeed,
		},
	}
	_, err = collectionUsers.UpdateOne(ctx, bson.M{"_id": idUser}, update)
	if err != nil {
		return advertisements.Advertisements{}, err
	}

	// Crear el documento del anuncio
	var documento advertisements.Advertisements
	documento.Name = ad.Name
	documento.NameUser = ad.NameUser
	documento.Destination = ad.Destination
	documento.Categorie = ad.Categorie
	documento.UrlVideo = ad.UrlVideo
	documento.ReferenceLink = ad.ReferenceLink
	documento.ImpressionsMax = ad.ImpressionsMax
	documento.PayPerPrint = floatValuePayPerPrint
	documento.Impressions = 0
	documento.ClicksMax = ad.ClicksMax
	documento.DocumentToBeAnnounced = ad.DocumentToBeAnnounced
	documento.IdOfTheUsersWhoClicked = []primitive.ObjectID{}
	documento.ClicksPerDay = []advertisements.ClicksPerDay{}
	documento.ImpressionsPerDay = []advertisements.ImpressionsPerDay{}
	documento.Timestamp = time.Now()
	documento.State = "pending"
	documento.ClipId = clipid

	// Insertar el anuncio en la base de datos
	_, err = collection.InsertOne(ctx, documento)
	if err != nil {
		return advertisements.Advertisements{}, err
	}

	// Devolver el anuncio creado
	return documento, nil
}
