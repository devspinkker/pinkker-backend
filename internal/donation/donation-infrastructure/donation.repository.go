package donationtinfrastructure

import (
	donationdomain "PINKKER-BACKEND/internal/donation/donation"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DonationRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewDonationRepository(redisClient *redis.Client, mongoClient *mongo.Client) *DonationRepository {
	return &DonationRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (D *DonationRepository) StateTheUserInChat(Donado primitive.ObjectID, Donante primitive.ObjectID) (bool, string, error) {
	ctx := context.Background()
	db := D.mongoClient.Database("PINKKER-BACKEND")

	stream, err := D.GetStreamByStreamerID(Donado)
	if err != nil {
		return true, "", err
	}
	userDonante, avatar, err := D.GetUserID(ctx, db, Donante)
	if err != nil {
		return true, avatar, err
	}

	userinfo, err := D.GetInfoUserInRoom(userDonante, stream.ID)
	return userinfo.Baneado, avatar, fmt.Errorf("room with ID %s not found for user %s", stream.ID, userDonante)
}

func (r *DonationRepository) GetInfoUserInRoom(nameUser string, getInfoUserInRoom primitive.ObjectID) (*userdomain.UserInfo, error) {
	database := r.mongoClient.Database("PINKKER-BACKEND")
	var room *userdomain.UserInfo
	filter := bson.D{
		{Key: "NameUser", Value: nameUser},
		{Key: "Rooms.Room", Value: getInfoUserInRoom},
	}

	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: filter},
		},
		{
			{Key: "$unwind", Value: "$Rooms"},
		},
		{
			{Key: "$match", Value: bson.D{{Key: "Rooms.Room", Value: getInfoUserInRoom}}},
		},
		{
			{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$Rooms"}}},
		},
	}

	// Obtener el cursor
	cursor, err := database.Collection("UserInformationInAllRooms").Aggregate(context.Background(), pipeline)
	if err != nil {
		return room, err
	}
	defer cursor.Close(context.Background())

	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&room); err != nil {
			return room, err
		}
	}

	if err := cursor.Err(); err != nil {
		return room, err
	}

	return room, nil
}

func (u *DonationRepository) GetTOTPSecret(ctx context.Context, userID primitive.ObjectID) (string, error) {
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

func (d *DonationRepository) UserHasNumberPikels(FromUser primitive.ObjectID, Pixeles float64) error {
	GoMongoDBCollUsers := d.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	PixelesTotals := (Pixeles*2)/100 + Pixeles
	filter := bson.M{"_id": FromUser, "Pixeles": bson.M{"$gte": PixelesTotals}}

	err := GoMongoDBCollUsers.FindOne(context.Background(), filter)

	if err != nil {
		if err.Err() == mongo.ErrNoDocuments {
			return errors.New("insufficient pixels")
		} else {
			return err.Err()
		}
	}
	return nil
}
func (D *DonationRepository) LatestStreamSummaryByUpdateDonations(streamerID primitive.ObjectID, pixeles float64) error {
	ctx := context.Background()

	GoMongoDB := D.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")

	filter := bson.M{
		"StreamerID": streamerID,
	}

	update := bson.M{
		"$inc": bson.M{
			"DonationsMoney": pixeles,
			"TotalMoney":     pixeles,
		},
	}

	opts := options.FindOneAndUpdate().SetSort(bson.D{{Key: "StartOfStream", Value: -1}}).SetReturnDocument(options.After)

	result := GoMongoDBCollStreamSummary.FindOneAndUpdate(ctx, filter, update, opts)
	if err := result.Err(); err != nil {
		return err
	}

	return nil
}

func (D *DonationRepository) GetWebSocketClientsInRoom(roomID string) ([]*websocket.Conn, error) {
	clients, err := utils.NewChatService().GetWebSocketClientsInRoom(roomID)

	return clients, err
}
func (D *DonationRepository) GetStreamByStreamerID(user primitive.ObjectID) (streamdomain.Stream, error) {
	GoMongoDBColl := D.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreams := GoMongoDBColl.Collection("Streams")
	filter := bson.M{"StreamerID": user}
	var stream streamdomain.Stream
	err := GoMongoDBCollStreams.FindOne(context.Background(), filter).Decode(&stream)
	return stream, err
}
func (u *DonationRepository) GetUserID(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) (string, string, error) {
	usersCollection := db.Collection("Users")

	// Filtro para buscar el usuario por ID
	filter := bson.M{"_id": userID}

	// Proyección para solo obtener la propiedad NameUser
	projection := bson.M{"NameUser": 1, "Avatar": 1}

	var result struct {
		NameUser string `bson:"NameUser"`
		Avatar   string `bson:"Avatar"`
	}

	// Consulta con proyección para obtener solo NameUser
	err := usersCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		return "", "", err
	}

	return result.NameUser, result.Avatar, nil
}

func (D *DonationRepository) PublishNotification(roomID string, noty map[string]interface{}) error {

	chatMessageJSON, err := json.Marshal(noty)
	if err != nil {
		return err
	}
	err = D.redisClient.Publish(
		context.Background(),
		roomID+"action",
		string(chatMessageJSON),
	).Err()
	if err != nil {
		return err
	}

	return err
}

// DonatePixels transfiere pixeles de un usuario a otro
func (d *DonationRepository) DonatePixels(FromUser, ToUser primitive.ObjectID, Pixels float64, text string) error {
	// Obtener las colecciones "Users" y "Donations"
	GoMongoDBCollUsers := d.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	GoMongoDBCollDonations := d.mongoClient.Database("PINKKER-BACKEND").Collection("Donations")

	// Iniciar una sesión para realizar las actualizaciones de manera transaccional
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	// Iniciar una transacción
	err = session.StartTransaction()
	if err != nil {
		return err
	}

	// Actualizar el usuario donante (FromUser)
	PixelesTotals := (Pixels*2)/100 + Pixels
	_, err = GoMongoDBCollUsers.UpdateOne(
		context.Background(),
		primitive.D{{Key: "_id", Value: FromUser}},
		primitive.D{{Key: "$inc", Value: primitive.D{{Key: "Pixeles", Value: -PixelesTotals}}}},
	)
	if err != nil {
		session.AbortTransaction(context.Background())
		return err
	}

	// Verificar si el usuario receptor (ToUser) existe
	toUserExists, err := GoMongoDBCollUsers.CountDocuments(
		context.Background(),
		primitive.D{{Key: "_id", Value: ToUser}},
	)
	if err != nil {
		session.AbortTransaction(context.Background())
		return err
	}
	if toUserExists == 0 {
		session.AbortTransaction(context.Background())
		return errors.New("el usuario receptor no existe")
	}

	// Actualizar el usuario receptor (ToUser)
	_, err = GoMongoDBCollUsers.UpdateOne(
		context.Background(),
		primitive.D{{Key: "_id", Value: ToUser}},
		primitive.D{{Key: "$inc", Value: primitive.D{{Key: "Pixeles", Value: Pixels}}}},
	)
	if err != nil {
		session.AbortTransaction(context.Background())
		return err
	}

	// Crear el documento de donación
	donation := donationdomain.Donation{
		FromUser:  FromUser,
		ToUser:    ToUser,
		Pixeles:   Pixels,
		Text:      text,
		TimeStamp: time.Now(),
	}

	// Insertar el documento de donación
	_, err = GoMongoDBCollDonations.InsertOne(context.Background(), donation)
	if err != nil {
		session.AbortTransaction(context.Background())
		return err
	}

	// Finalizar la transacción
	err = session.CommitTransaction(context.Background())
	if err != nil {
		return err
	}

	return err
}

// donadores de pixeles con Notified en false
func (d *DonationRepository) MyPixelesdonors(id primitive.ObjectID) ([]donationdomain.ResDonation, error) {
	GoMongoDBCollDonations := d.mongoClient.Database("PINKKER-BACKEND").Collection("Donations")
	filter := bson.D{
		{Key: "ToUser", Value: id},
		{Key: "Notified", Value: false},
	}
	pipeline := []bson.D{
		// Filtra las donations que cumplan con el filtro
		{{Key: "$match", Value: filter}},
		// Une las donations con la información del usuario donante (FromUser)
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "FromUser"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "FromUserInfo"},
		}}},
		{{Key: "$unwind", Value: "$FromUserInfo"}},
		// Proyecta los campos necesarios
		{{Key: "$project", Value: bson.D{
			{Key: "FromUser", Value: "$FromUser"},
			{Key: "FromUserInfo.Avatar", Value: "$FromUserInfo.Avatar"},
			{Key: "FromUserInfo.NameUser", Value: "$FromUserInfo.NameUser"},
			{Key: "Pixeles", Value: 1},
			{Key: "Text", Value: 1},
			{Key: "TimeStamp", Value: 1},
			{Key: "Notified", Value: 1},
			{Key: "ToUser", Value: 1},
			{Key: "id", Value: "$_id"},
		}}},
		{{Key: "$limit", Value: 20}},
	}
	cursor, err := GoMongoDBCollDonations.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var Donations []donationdomain.ResDonation
	for cursor.Next(context.Background()) {
		var donation donationdomain.ResDonation
		if err := cursor.Decode(&donation); err != nil {
			return nil, err
		}
		Donations = append(Donations, donation)
	}
	if len(Donations) == 0 {
		return nil, errors.New("no documents found")

	}
	return Donations, nil
}

// todos los donantes de Pixeles de user token
func (d *DonationRepository) AllMyPixelesDonors(id primitive.ObjectID) ([]donationdomain.ResDonation, error) {
	GoMongoDBCollDonations := d.mongoClient.Database("PINKKER-BACKEND").Collection("Donations")
	filter := bson.D{
		{Key: "ToUser", Value: id},
	}
	pipeline := []bson.D{
		// Filtra las donations que cumplan con el filtro
		{{Key: "$match", Value: filter}},
		// Une las donations con la información del usuario donante (FromUser)
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "FromUser"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "FromUserInfo"},
		}}},
		{{Key: "$unwind", Value: "$FromUserInfo"}},
		// Proyecta los campos necesarios
		{{Key: "$project", Value: bson.D{
			{Key: "FromUser", Value: "$FromUser"},
			{Key: "FromUserInfo.Avatar", Value: "$FromUserInfo.Avatar"},
			{Key: "FromUserInfo.NameUser", Value: "$FromUserInfo.NameUser"},
			{Key: "Pixeles", Value: 1},
			{Key: "Text", Value: 1},
			{Key: "TimeStamp", Value: 1},
			{Key: "Notified", Value: 1},
			{Key: "ToUser", Value: 1},
			{Key: "id", Value: "$_id"},
		}}},
		{{Key: "$limit", Value: 20}},
	}
	cursor, err := GoMongoDBCollDonations.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var Donations []donationdomain.ResDonation

	for cursor.Next(context.Background()) {
		var donation donationdomain.ResDonation
		if err := cursor.Decode(&donation); err != nil {
			return nil, err
		}
		Donations = append(Donations, donation)
	}
	if len(Donations) == 0 {
		return nil, errors.New("no documents found")

	}
	return Donations, nil
}
func (d *DonationRepository) GetPixelesDonationsChat(id primitive.ObjectID) ([]donationdomain.ResDonation, error) {
	GoMongoDBCollDonations := d.mongoClient.Database("PINKKER-BACKEND").Collection("Donations")
	filter := bson.D{
		{Key: "ToUser", Value: id},
	}
	pipeline := []bson.D{
		// Filtra las donations que cumplan con el filtro
		{{Key: "$match", Value: filter}},
		{{Key: "$sort", Value: bson.D{{Key: "TimeStamp", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "FromUser"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "FromUserInfo"},
		}}},
		{{Key: "$unwind", Value: "$FromUserInfo"}},
		// Proyecta los campos necesarios
		{{Key: "$project", Value: bson.D{
			{Key: "FromUser", Value: "$FromUser"},
			{Key: "FromUserInfo.Avatar", Value: "$FromUserInfo.Avatar"},
			{Key: "FromUserInfo.NameUser", Value: "$FromUserInfo.NameUser"},
			{Key: "Pixeles", Value: 1},
			{Key: "Text", Value: 1},
			{Key: "TimeStamp", Value: 1},
			{Key: "Notified", Value: 1},
			{Key: "ToUser", Value: 1},
			{Key: "id", Value: "$_id"},
		}}},
	}
	cursor, err := GoMongoDBCollDonations.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var Donations []donationdomain.ResDonation

	for cursor.Next(context.Background()) {
		var donation donationdomain.ResDonation
		if err := cursor.Decode(&donation); err != nil {
			return nil, err
		}
		Donations = append(Donations, donation)
	}
	if len(Donations) == 0 {
		return nil, errors.New("no documents found")

	}
	return Donations, nil
}

// actualzaa el Notified
func (d *DonationRepository) UpdateDonationsNotifiedStatus(donations []donationdomain.ResDonation) error {
	GoMongoDBCollDonations := d.mongoClient.Database("PINKKER-BACKEND").Collection("Donations")
	var donationsID []primitive.ObjectID
	for _, value := range donations {
		donationsID = append(donationsID, value.ID)
	}

	filter := bson.D{
		{Key: "_id", Value: bson.D{{Key: "$in", Value: donationsID}}},
	}

	// Construir la actualización
	updateDoc := bson.D{
		{Key: "$set", Value: bson.D{{Key: "Notified", Value: true}}},
	}
	_, err := GoMongoDBCollDonations.UpdateMany(context.Background(), filter, updateDoc)
	if err != nil {
		return err
	}

	return nil

}
