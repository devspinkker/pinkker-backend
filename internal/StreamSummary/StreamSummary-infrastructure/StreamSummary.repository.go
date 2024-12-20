package StreamSummaryinfrastructure

import (
	"PINKKER-BACKEND/config"
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-domain"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type StreamSummaryRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewStreamSummaryRepository(redisClient *redis.Client, mongoClient *mongo.Client) *StreamSummaryRepository {
	return &StreamSummaryRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (r *StreamSummaryRepository) SetStreamSummaryUnavailableByIDAndStreamerID(id primitive.ObjectID, streamerID primitive.ObjectID, Available bool) error {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	// Filtro para encontrar el documento con el ID y StreamerID especificados
	filter := bson.M{
		"_id":        id,
		"StreamerID": streamerID,
	}

	// Actualización para establecer Available en false
	update := bson.M{
		"$set": bson.M{
			"Available": Available,
		},
	}

	// Ejecutar la actualización
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("no se encontró ningún resumen de stream con el ID o no pertenece al streamer")
	}

	return nil
}

func (r *StreamSummaryRepository) UpdateStreamSummaryByIDAndStreamerID(id primitive.ObjectID, streamerID primitive.ObjectID, newTitle string) error {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	filter := bson.M{
		"_id":        id,
		"StreamerID": streamerID,
	}

	// Definir los campos a actualizar
	update := bson.M{
		"$set": bson.M{
			"Title": newTitle,
		},
	}

	opts := options.Update().SetUpsert(false)
	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("no se encontró ningún resumen de stream con el ID o no pertenece al streamer")
	}

	return nil
}

func (r *StreamSummaryRepository) GetEarningsByRange(streamerID primitive.ObjectID, startDate, endDate time.Time) (StreamSummarydomain.Earnings, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	filter := bson.M{
		"StreamerID": streamerID,
		"StartOfStream": bson.M{
			"$gte": startDate,
			"$lt":  endDate,
		},
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "TotalAdmoney", Value: bson.D{{Key: "$sum", Value: "$Admoney"}}},
			{Key: "TotalSubscriptionsMoney", Value: bson.D{{Key: "$sum", Value: "$SubscriptionsMoney"}}},
			{Key: "TotalDonationsMoney", Value: bson.D{{Key: "$sum", Value: "$DonationsMoney"}}},
		}}},
	}

	opts := options.Aggregate()

	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return StreamSummarydomain.Earnings{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalAdmoney            float64 `bson:"TotalAdmoney"`
		TotalSubscriptionsMoney float64 `bson:"TotalSubscriptionsMoney"`
		TotalDonationsMoney     float64 `bson:"TotalDonationsMoney"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return StreamSummarydomain.Earnings{}, err
	}

	if len(result) == 0 {
		return StreamSummarydomain.Earnings{}, nil
	}

	earnings := StreamSummarydomain.Earnings{
		Admoney:            result[0].TotalAdmoney,
		SubscriptionsMoney: result[0].TotalSubscriptionsMoney,
		DonationsMoney:     result[0].TotalDonationsMoney,
	}

	return earnings, nil
}

func (r *StreamSummaryRepository) GetEarningsByDay(streamerID primitive.ObjectID, day time.Time) (StreamSummarydomain.Earnings, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := bson.M{
		"StreamerID": streamerID,
		"StartOfStream": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "TotalAdmoney", Value: bson.D{{Key: "$sum", Value: "$Admoney"}}},
			{Key: "TotalSubscriptionsMoney", Value: bson.D{{Key: "$sum", Value: "$SubscriptionsMoney"}}},
			{Key: "TotalDonationsMoney", Value: bson.D{{Key: "$sum", Value: "$DonationsMoney"}}},
		}}},
	}

	opts := options.Aggregate()

	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return StreamSummarydomain.Earnings{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalAdmoney            float64 `bson:"TotalAdmoney"`
		TotalSubscriptionsMoney float64 `bson:"TotalSubscriptionsMoney"`
		TotalDonationsMoney     float64 `bson:"TotalDonationsMoney"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return StreamSummarydomain.Earnings{}, err
	}

	if len(result) == 0 {
		return StreamSummarydomain.Earnings{}, nil // or return an empty Earnings struct with zero values
	}

	earnings := StreamSummarydomain.Earnings{
		Admoney:            result[0].TotalAdmoney,
		SubscriptionsMoney: result[0].TotalSubscriptionsMoney,
		DonationsMoney:     result[0].TotalDonationsMoney,
	}

	return earnings, nil
}

func (r *StreamSummaryRepository) GetEarningsByWeek(streamerID primitive.ObjectID, week time.Time) (StreamSummarydomain.Earnings, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	startOfWeek := week.AddDate(0, 0, -int(week.Weekday()))
	endOfWeek := startOfWeek.Add(7 * 24 * time.Hour)

	filter := bson.M{
		"StreamerID": streamerID,
		"StartOfStream": bson.M{
			"$gte": startOfWeek,
			"$lt":  endOfWeek,
		},
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "TotalAdmoney", Value: bson.D{{Key: "$sum", Value: "$Admoney"}}},
			{Key: "TotalSubscriptionsMoney", Value: bson.D{{Key: "$sum", Value: "$SubscriptionsMoney"}}},
			{Key: "TotalDonationsMoney", Value: bson.D{{Key: "$sum", Value: "$DonationsMoney"}}},
		}}},
	}

	opts := options.Aggregate()

	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return StreamSummarydomain.Earnings{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalAdmoney            float64 `bson:"TotalAdmoney"`
		TotalSubscriptionsMoney float64 `bson:"TotalSubscriptionsMoney"`
		TotalDonationsMoney     float64 `bson:"TotalDonationsMoney"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return StreamSummarydomain.Earnings{}, err
	}

	if len(result) == 0 {
		return StreamSummarydomain.Earnings{}, nil // or return an empty Earnings struct with zero values
	}

	earnings := StreamSummarydomain.Earnings{
		Admoney:            result[0].TotalAdmoney,
		SubscriptionsMoney: result[0].TotalSubscriptionsMoney,
		DonationsMoney:     result[0].TotalDonationsMoney,
	}

	return earnings, nil
}

func (r *StreamSummaryRepository) GetEarningsByMonth(streamerID primitive.ObjectID, month time.Time) (StreamSummarydomain.Earnings, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	filter := bson.M{
		"StreamerID": streamerID,
		"StartOfStream": bson.M{
			"$gte": startOfMonth,
			"$lt":  endOfMonth,
		},
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "TotalAdmoney", Value: bson.D{{Key: "$sum", Value: "$Admoney"}}},
			{Key: "TotalSubscriptionsMoney", Value: bson.D{{Key: "$sum", Value: "$SubscriptionsMoney"}}},
			{Key: "TotalDonationsMoney", Value: bson.D{{Key: "$sum", Value: "$DonationsMoney"}}},
		}}},
	}

	opts := options.Aggregate()

	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return StreamSummarydomain.Earnings{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalAdmoney            float64 `bson:"TotalAdmoney"`
		TotalSubscriptionsMoney float64 `bson:"TotalSubscriptionsMoney"`
		TotalDonationsMoney     float64 `bson:"TotalDonationsMoney"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return StreamSummarydomain.Earnings{}, err
	}

	if len(result) == 0 {
		return StreamSummarydomain.Earnings{}, nil // or return an empty Earnings struct with zero values
	}

	earnings := StreamSummarydomain.Earnings{
		Admoney:            result[0].TotalAdmoney,
		SubscriptionsMoney: result[0].TotalSubscriptionsMoney,
		DonationsMoney:     result[0].TotalDonationsMoney,
	}

	return earnings, nil
}
func (r *StreamSummaryRepository) GetDailyEarningsForMonth(streamerID primitive.ObjectID, month time.Time) ([]StreamSummarydomain.EarningsPerDay, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	startOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	filter := bson.M{
		"StreamerID": streamerID,
		"StartOfStream": bson.M{
			"$gte": startOfMonth,
			"$lt":  endOfMonth,
		},
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "year", Value: bson.D{{Key: "$year", Value: "$StartOfStream"}}},
				{Key: "month", Value: bson.D{{Key: "$month", Value: "$StartOfStream"}}},
				{Key: "day", Value: bson.D{{Key: "$dayOfMonth", Value: "$StartOfStream"}}},
			}},
			{Key: "TotalAdmoney", Value: bson.D{{Key: "$sum", Value: "$Admoney"}}},
			{Key: "TotalSubscriptionsMoney", Value: bson.D{{Key: "$sum", Value: "$SubscriptionsMoney"}}},
			{Key: "TotalDonationsMoney", Value: bson.D{{Key: "$sum", Value: "$DonationsMoney"}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "_id.year", Value: 1},
			{Key: "_id.month", Value: 1},
			{Key: "_id.day", Value: 1},
		}}},
	}

	opts := options.Aggregate()

	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID struct {
			Year  int `bson:"year"`
			Month int `bson:"month"`
			Day   int `bson:"day"`
		} `bson:"_id"`
		TotalAdmoney            float64 `bson:"TotalAdmoney"`
		TotalSubscriptionsMoney float64 `bson:"TotalSubscriptionsMoney"`
		TotalDonationsMoney     float64 `bson:"TotalDonationsMoney"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	earningsMap := make(map[int]StreamSummarydomain.Earnings)

	for _, result := range results {
		earningsMap[result.ID.Day] = StreamSummarydomain.Earnings{
			Admoney:            result.TotalAdmoney,
			SubscriptionsMoney: result.TotalSubscriptionsMoney,
			DonationsMoney:     result.TotalDonationsMoney,
		}
	}

	var earningsPerDay []StreamSummarydomain.EarningsPerDay

	for day := 1; day <= endOfMonth.AddDate(0, 0, -1).Day(); day++ {
		date := time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, time.UTC)
		earnings := earningsMap[day] // Recoge las ganancias del mapa
		earningsPerDay = append(earningsPerDay, StreamSummarydomain.EarningsPerDay{
			Date:     date,
			Earnings: earnings,
		})
	}

	return earningsPerDay, nil
}

func (r *StreamSummaryRepository) GetEarningsByYear(streamerID primitive.ObjectID, year time.Time) (StreamSummarydomain.Earnings, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	startOfYear := time.Date(year.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := startOfYear.AddDate(1, 0, 0)

	filter := bson.M{
		"StreamerID": streamerID,
		"StartOfStream": bson.M{
			"$gte": startOfYear,
			"$lt":  endOfYear,
		},
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "TotalAdmoney", Value: bson.D{{Key: "$sum", Value: "$Admoney"}}},
			{Key: "TotalSubscriptionsMoney", Value: bson.D{{Key: "$sum", Value: "$SubscriptionsMoney"}}},
			{Key: "TotalDonationsMoney", Value: bson.D{{Key: "$sum", Value: "$DonationsMoney"}}},
		}}},
	}

	opts := options.Aggregate()

	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return StreamSummarydomain.Earnings{}, err
	}
	defer cursor.Close(ctx)

	var result []struct {
		TotalAdmoney            float64 `bson:"TotalAdmoney"`
		TotalSubscriptionsMoney float64 `bson:"TotalSubscriptionsMoney"`
		TotalDonationsMoney     float64 `bson:"TotalDonationsMoney"`
	}

	if err := cursor.All(ctx, &result); err != nil {
		return StreamSummarydomain.Earnings{}, err
	}

	if len(result) == 0 {
		return StreamSummarydomain.Earnings{}, nil // or return an empty Earnings struct with zero values
	}

	earnings := StreamSummarydomain.Earnings{
		Admoney:            result[0].TotalAdmoney,
		SubscriptionsMoney: result[0].TotalSubscriptionsMoney,
		DonationsMoney:     result[0].TotalDonationsMoney,
	}

	return earnings, nil
}

func (r *StreamSummaryRepository) GetTopVodsLast48Hours() ([]StreamSummarydomain.StreamSummaryGet, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	fortyEightHoursAgo := time.Now().Add(-48 * time.Hour)

	filter := bson.M{
		"StartOfStream": bson.M{
			"$gte": fortyEightHoursAgo,
		},
		"Available": true,
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "StreamerID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "UserInfo.Avatar", Value: "$UserInfo.Avatar"},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "Title", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "EndOfStream", Value: 1},
			{Key: "MaxViewers", Value: 1},
			{Key: "StartOfStream", Value: 1},
			{Key: "StreamerID", Value: 1},
			{Key: "StreamCategory", Value: 1},
			{Key: "UserInfo.Avatar", Value: "$UserInfo.Avatar"},
			{Key: "UserInfo.FullName", Value: "$UserInfo.FullName"},
			{Key: "UserInfo.NameUser", Value: "$UserInfo.NameUser"},
		}}},
		bson.D{{Key: "$limit", Value: 10}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "MaxViewers", Value: -1},
		}}},
	}

	opts := options.Aggregate()

	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var summaries []StreamSummarydomain.StreamSummaryGet
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *StreamSummaryRepository) GetStreamSummaryByTitle(title string) ([]StreamSummarydomain.StreamSummaryGet, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	filter := bson.M{
		"Title":     primitive.Regex{Pattern: title, Options: "i"},
		"Available": true,
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "StreamerID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "Title", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "EndOfStream", Value: 1},
			{Key: "MaxViewers", Value: 1},
			{Key: "StartOfStream", Value: 1},
			{Key: "StreamerID", Value: 1},
			{Key: "StreamCategory", Value: 1},
			{Key: "UserInfo", Value: "$UserInfo"},
		}}},
		bson.D{{Key: "$limit", Value: 10}},
	}

	opts := options.Aggregate()

	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var summaries []StreamSummarydomain.StreamSummaryGet
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *StreamSummaryRepository) GetStreamSummariesByStreamerIDLast30Days(streamerID primitive.ObjectID) ([]StreamSummarydomain.StreamSummaryGet, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	filter := bson.M{
		"StreamerID": streamerID,
		"StartOfStream": bson.M{
			"$gte": thirtyDaysAgo,
		},
		"Available": true,
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "StreamerID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "Title", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "EndOfStream", Value: 1},
			{Key: "MaxViewers", Value: 1},
			{Key: "StartOfStream", Value: 1},
			{Key: "StreamerID", Value: 1},
			{Key: "StreamCategory", Value: 1},
			{Key: "UserInfo", Value: "$UserInfo"},
		}}},
		bson.D{{Key: "$limit", Value: 20}},
	}

	opts := options.Aggregate()

	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var summaries []StreamSummarydomain.StreamSummaryGet
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *StreamSummaryRepository) GetStreamSummaryByID(id primitive.ObjectID) (*StreamSummarydomain.StreamSummaryGet, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	filter := bson.M{
		"_id":       id,
		"Available": true,
	}

	projection := bson.D{
		{Key: "_id", Value: 1},
		{Key: "Title", Value: 1},
		{Key: "StreamThumbnail", Value: 1},
		{Key: "EndOfStream", Value: 1},
		{Key: "MaxViewers", Value: 1},
		{Key: "StartOfStream", Value: 1},
		{Key: "StreamerID", Value: 1},
	}

	opts := options.FindOne().SetProjection(projection)

	var streamSummary StreamSummarydomain.StreamSummaryGet
	err := collection.FindOne(ctx, filter, opts).Decode(&streamSummary)
	if err != nil {
		return nil, err
	}

	return &streamSummary, nil
}

func (r *StreamSummaryRepository) UpdateStreamSummary(StreamerID primitive.ObjectID, data StreamSummarydomain.UpdateStreamSummary) error {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")

	// Definir el filtro para encontrar el resumen del stream por StreamerID
	filter := bson.M{"StreamerID": StreamerID}

	// Definir la actualización con los campos proporcionados en data
	update := bson.M{
		"$set": bson.M{
			"AverageViewers":   data.AverageViewers,
			"MaxViewers":       data.MaxViewers,
			"NewFollowers":     data.NewFollowers,
			"NewSubscriptions": data.NewSubscriptions,
			"Advertisements":   data.Advertisements,
		},
	}

	// Ejecutar la actualización en la colección de StreamSummary
	_, err := GoMongoDBCollStreamSummary.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
func (r *StreamSummaryRepository) AddAds(idValueObj primitive.ObjectID, nameUser string, AddAds StreamSummarydomain.AddAds) error {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")
	GoMongoDBCollAdvertisements := GoMongoDB.Collection("Advertisements")

	filter := bson.M{"_id": idValueObj}
	result := GoMongoDBCollUsers.FindOne(ctx, filter)
	if result.Err() != nil {
		return result.Err()
	}

	key := "ADS_" + idValueObj.Hex()
	exists, err := r.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		return nil
	}

	filterSummary := bson.M{"StreamerID": AddAds.StreamerID}
	opts := options.FindOne().SetSort(bson.D{{Key: "StartOfStream", Value: -1}})
	var streamSummary StreamSummarydomain.StreamSummary
	err = GoMongoDBCollStreamSummary.FindOne(ctx, filterSummary, opts).Decode(&streamSummary)
	if err != nil {
		return err
	}

	connected, err := r.RedisIsUserInRoom(ctx, streamSummary.StreamDocumentID.Hex(), nameUser)
	if err != nil || !connected {
		return errors.New("usuario no activo en la sala")
	}
	// Actualización del conteo de anuncios en el resumen del stream
	updateStream := bson.M{
		"$inc": bson.M{"Advertisements": 1},
	}
	filterUpdate := bson.M{"_id": streamSummary.ID}
	_, err = GoMongoDBCollStreamSummary.UpdateOne(ctx, filterUpdate, updateStream)
	if err != nil {
		return err
	}

	currentDate := time.Now().Format("2006-01-02")

	advertisementFilter := bson.M{"_id": AddAds.AdvertisementsId}

	// Actualización para incrementar el conteo total de impresiones
	updateImpressions := bson.M{
		"$inc": bson.M{
			"Impressions": 1,
		},
	}

	// Ejecutar la actualización de impresiones
	_, err = GoMongoDBCollAdvertisements.UpdateOne(ctx, advertisementFilter, updateImpressions)
	if err != nil {
		return err
	}

	// Actualización para impresiones diarias
	updateImpressionsPerDay := bson.M{
		"$inc": bson.M{
			"ImpressionsPerDay.$[elem].Impressions": 1,
		},
	}

	arrayFilter := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.Date": currentDate},
		},
	}

	updateResult, err := GoMongoDBCollAdvertisements.UpdateOne(ctx, advertisementFilter, updateImpressionsPerDay, options.Update().SetArrayFilters(arrayFilter))
	if err != nil {
		return err
	}

	// Si no se actualizó ningún documento, crear un nuevo registro para la fecha actual

	if updateResult.ModifiedCount == 0 {
		newDateUpdate := bson.M{
			"$addToSet": bson.M{
				"ImpressionsPerDay": bson.M{
					"Date":        currentDate,
					"Impressions": 1,
				},
			},
		}

		_, err = GoMongoDBCollAdvertisements.UpdateOne(ctx, bson.M{"_id": AddAds.AdvertisementsId}, newDateUpdate)
		if err != nil {
			return err
		}
	}
	err = r.updatePinkkerProfitPerMonth(ctx)
	if err != nil {
		return err
	}

	// Establecer el valor en Redis
	err = r.redisClient.Set(ctx, key, "true", time.Minute*10).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *StreamSummaryRepository) RedisIsUserInRoom(ctx context.Context, roomID, nameUser string) (bool, error) {
	userKey := "ActiveUserRooms:" + nameUser

	isActive, err := r.redisClient.SIsMember(ctx, userKey, roomID).Result()
	if err != nil {
		return false, err
	}

	return isActive, nil
}

func (r *StreamSummaryRepository) updatePinkkerProfitPerMonth(ctx context.Context) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollMonthly := GoMongoDB.Collection("PinkkerProfitPerMonth")
	AdvertisementsPayPerPrint := config.AdvertisementsPayPerPrint()

	AdvertisementsPayPerPrintFloat, err := strconv.ParseFloat(AdvertisementsPayPerPrint, 64)
	if err != nil {
		log.Fatalf("error al convertir el valor: %v", err)
	}
	impressions := int(AdvertisementsPayPerPrintFloat)
	currentTime := time.Now()
	currentMonth := int(currentTime.Month())
	currentYear := currentTime.Year()

	// Usamos la función para obtener el día en formato "YYYY-MM-DD"
	currentDay := helpers.GetDayOfMonth(currentTime)

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
			"timestamp":          currentTime,
			"days." + currentDay: PinkkerProfitPerMonthdomain.NewDefaultDay(),
		},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	// Paso 2: Incrementa los valores en 'days.day_x'
	monthlyUpdate := bson.M{
		"$inc": bson.M{
			"total":                               AdvertisementsPayPerPrintFloat,
			"days." + currentDay + ".impressions": impressions,
		},
	}

	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (r *StreamSummaryRepository) UpdateInfoStreamSummary(StreamerID primitive.ObjectID) error {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")
	GoMongoDBCollStreams := GoMongoDB.Collection("Streams")

	// Obtener el Stream
	filter := bson.M{"StreamerID": StreamerID}
	var Stream streamdomain.Stream
	result := GoMongoDBCollStreams.FindOne(ctx, filter).Decode(&Stream)
	if result != nil {
		return result
	}

	// Obtener el último StreamSummary para el Streamer
	filter = bson.M{"StreamerID": StreamerID}
	opts := options.FindOne().SetSort(bson.D{{Key: "StartOfStream", Value: -1}})

	var streamSummary StreamSummarydomain.StreamSummary
	err := GoMongoDBCollStreamSummary.FindOne(ctx, filter, opts).Decode(&streamSummary)
	if err != nil {
		return err
	}

	currentDateTime := time.Now().Format("2006-01-02 15:04:05")
	UniqueUserinRoom, err := r.AddUniqueUserInteractionInStreamSummary(Stream.ID)
	if err != nil {
		return err
	}

	// Calcular la puntuación de recomendación
	topAverageViewers := sumTopViewers(streamSummary.AverageViewersByTime)
	recommendationScore := 0.7*float64(topAverageViewers) + 0.3*float64(UniqueUserinRoom)

	// Actualizar el StreamSummary
	updateSummary := bson.M{
		"$set": bson.M{
			"UniqueInteractions":                      UniqueUserinRoom,
			"AverageViewersByTime." + currentDateTime: Stream.ViewerCount,
			"RecommendationScore":                     recommendationScore,
		},
	}
	filterUpdateSummary := bson.M{"_id": streamSummary.ID}
	_, err = GoMongoDBCollStreamSummary.UpdateOne(ctx, filterUpdateSummary, updateSummary)
	if err != nil {
		return err
	}

	// Actualizar el Stream con el mismo RecommendationScore
	updateStream := bson.M{
		"$set": bson.M{
			"RecommendationScore": recommendationScore,
		},
	}
	notification := map[string]interface{}{
		"action": "views",
		"views":  Stream.ViewerCount,
	}
	r.PublishAction(Stream.ID.Hex()+"action", notification)
	filterUpdateStream := bson.M{"_id": Stream.ID}
	_, err = GoMongoDBCollStreams.UpdateOne(ctx, filterUpdateStream, updateStream)
	if err != nil {
		return err
	}

	return nil
}
func (r *StreamSummaryRepository) PublishAction(roomID string, noty map[string]interface{}) error {

	chatMessageJSON, err := json.Marshal(noty)
	if err != nil {
		return err
	}
	err = r.redisClient.Publish(
		context.Background(),
		roomID,
		string(chatMessageJSON),
	).Err()
	if err != nil {
		return err
	}

	return err
}
func (r *StreamSummaryRepository) AddUniqueUserInteractionInStreamSummary(Room primitive.ObjectID) (int64, error) {
	key := Room.Hex() + ":uniqueinteractions"

	setSize, err := r.redisClient.SCard(context.Background(), key).Result()

	if err != nil {
		if err != redis.Nil {
			return 0, err
		}
		return 0, nil
	}

	return setSize, nil
}

func sumTopViewers(averageViewersByTime map[string]int) int {
	viewerCounts := []int{}
	for _, viewers := range averageViewersByTime {
		viewerCounts = append(viewerCounts, viewers)
		if len(viewerCounts) == 5 {
			break
		}
	}

	total := 0
	for _, count := range viewerCounts {
		total += count
	}
	return total
}
func (r *StreamSummaryRepository) GetLastSixStreamSummariesBeforeDate(StreamerID primitive.ObjectID, date time.Time) ([]StreamSummarydomain.StreamSummary, error) {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")

	filter := bson.M{
		"StreamerID": StreamerID,
		"StartOfStream": bson.M{
			"$lt": date,
		},
		"Available": true,
	}

	opts := options.Find().SetSort(bson.D{{Key: "StartOfStream", Value: -1}}).SetLimit(6)

	cursor, err := GoMongoDBCollStreamSummary.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var summaries []StreamSummarydomain.StreamSummary
	for cursor.Next(ctx) {
		var summary StreamSummarydomain.StreamSummary
		if err := cursor.Decode(&summary); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}
func (r *StreamSummaryRepository) AWeekOfStreaming(StreamerID primitive.ObjectID, page int) ([]StreamSummarydomain.StreamSummary, error) {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")

	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -((int(now.Weekday()) + 6) % 7))

	startDate := startOfWeek.AddDate(0, 0, (page-1)*7)
	endDate := startDate.AddDate(0, 0, 7)

	filter := bson.M{
		"StreamerID": StreamerID,
		"StartOfStream": bson.M{
			"$gte": startDate,
			"$lt":  endDate,
		},
	}

	pageSize := 12

	opts := options.Find().
		SetSort(bson.D{{Key: "StartOfStream", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := GoMongoDBCollStreamSummary.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var summaries []StreamSummarydomain.StreamSummary
	for cursor.Next(ctx) {
		var summary StreamSummarydomain.StreamSummary
		if err := cursor.Decode(&summary); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}
