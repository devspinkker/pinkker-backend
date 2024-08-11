package StreamSummaryinfrastructure

import (
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-domain"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	"context"
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
func (r *StreamSummaryRepository) GetTopVodsLast48Hours() ([]StreamSummarydomain.StreamSummaryGet, error) {
	ctx := context.Background()

	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("StreamSummary")

	fortyEightHoursAgo := time.Now().Add(-48 * time.Hour)

	filter := bson.M{
		"StartOfStream": bson.M{
			"$gte": fortyEightHoursAgo,
		},
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

	filter := bson.M{"Title": primitive.Regex{Pattern: title, Options: "i"}}

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

	filter := bson.M{"_id": id}

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
func (r *StreamSummaryRepository) AddAds(idValueObj primitive.ObjectID, AddAds StreamSummarydomain.AddAds) error {
	ctx := context.Background()

	// Inicializar colecciones
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")
	GoMongoDBCollAdvertisements := GoMongoDB.Collection("Advertisements")

	// Verificar si el usuario existe
	filter := bson.M{"_id": idValueObj}
	result := GoMongoDBCollUsers.FindOne(ctx, filter)
	if result.Err() != nil {
		return result.Err()
	}

	// Verificar si la entrada ya existe en Redis
	key := "ADS_" + idValueObj.Hex()
	exists, err := r.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		return nil
	}

	// Obtener el resumen del stream más reciente
	filterSummary := bson.M{"StreamerID": AddAds.StreamerID}
	opts := options.FindOne().SetSort(bson.D{{Key: "StartOfStream", Value: -1}})
	var streamSummary StreamSummarydomain.StreamSummary
	err = GoMongoDBCollStreamSummary.FindOne(ctx, filterSummary, opts).Decode(&streamSummary)
	if err != nil {
		return err
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

	// Obtener la fecha actual
	currentDate := time.Now().Format("2006-01-02")

	// Actualización principal para impresiones
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
		"$setOnInsert": bson.M{
			"ImpressionsPerDay": bson.M{
				"$each": []bson.M{
					{
						"Date":        currentDate,
						"Impressions": 1,
					},
				},
				"$position": -1,
			},
		},
	}

	// Filtros de array para encontrar la fecha actual
	arrayFilter := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.Date": currentDate},
		},
	}

	_, err = GoMongoDBCollAdvertisements.UpdateOne(ctx, advertisementFilter, updateImpressionsPerDay, options.Update().SetArrayFilters(arrayFilter))
	if err != nil {
		return err
	}

	err = r.redisClient.Set(ctx, key, "true", time.Minute*5).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *StreamSummaryRepository) AverageViewers(StreamerID primitive.ObjectID) error {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollStreamSummary := GoMongoDB.Collection("StreamSummary")
	GoMongoDBColl := GoMongoDB.Collection("Streams")

	filter := bson.M{"StreamerID": StreamerID}
	Stream := streamdomain.Stream{}
	result := GoMongoDBColl.FindOne(ctx, filter).Decode(&Stream)
	if result != nil {
		return result
	}

	filter = bson.M{"StreamerID": StreamerID}

	opts := options.FindOne().SetSort(bson.D{{Key: "StartOfStream", Value: -1}})

	var streamSummary StreamSummarydomain.StreamSummary
	err := GoMongoDBCollStreamSummary.FindOne(ctx, filter, opts).Decode(&streamSummary)
	if err != nil {
		return err
	}
	currentDateTime := time.Now().Format("2006-01-02 15:04:05")

	update := bson.M{
		"$set": bson.M{
			"AverageViewersByTime." + currentDateTime: Stream.ViewerCount,
		},
	}
	filterUpdate := bson.M{"_id": streamSummary.ID}
	_, err = GoMongoDBCollStreamSummary.UpdateOne(ctx, filterUpdate, update)
	if err != nil {
		return err
	}

	return nil
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
