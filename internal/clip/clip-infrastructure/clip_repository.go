// clip-infrastructure/clip_repository.go
package clipinfrastructure

import (
	"PINKKER-BACKEND/config"
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClipRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewClipRepository(redisClient *redis.Client, mongoClient *mongo.Client) *ClipRepository {
	return &ClipRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}

func (c *ClipRepository) GetClipsByNameUserIDOrdenación(UserID primitive.ObjectID, filterType string, dateRange string, page int, limit int) ([]clipdomain.GetClip, error) {
	clipCollection := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	// Definimos el pipeline inicial
	pipeline := mongo.Pipeline{
		// Filtrar por UserID
		bson.D{{Key: "$match", Value: bson.D{{Key: "UserID", Value: UserID}}}},
	}

	if dateRange != "" {
		dateFilter := c.getDateFilter(dateRange)
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: dateFilter}})
	}

	sortStage := c.getSortStage(filterType)

	if filterType == "random" {
		pipeline = append(pipeline, bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: limit}}}})
	} else {
		pipeline = append(pipeline, sortStage)

		pipeline = append(pipeline,
			bson.D{{Key: "$skip", Value: (page - 1) * limit}},
			bson.D{{Key: "$limit", Value: limit}},
		)
	}

	// Campos adicionales como el conteo de likes y comentarios
	// pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.D{
	// 	{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
	// 	{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
	// }}})

	// Ejecutar la consulta con el pipeline
	cursor, err := clipCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// Recoger los resultados
	var clips []clipdomain.GetClip
	if err := cursor.All(context.Background(), &clips); err != nil {
		return nil, err
	}

	return clips, nil
}

func (c *ClipRepository) getDateFilter(dateRange string) bson.D {
	currentTime := time.Now()
	switch dateRange {
	case "day":
		oneDayAgo := currentTime.Add(-24 * time.Hour)
		return bson.D{{Key: "timestamps.createdAt", Value: bson.D{{Key: "$gte", Value: oneDayAgo}}}}
	case "week":
		oneWeekAgo := currentTime.AddDate(0, 0, -7)
		return bson.D{{Key: "timestamps.createdAt", Value: bson.D{{Key: "$gte", Value: oneWeekAgo}}}}
	case "month":
		oneMonthAgo := currentTime.AddDate(0, -1, 0)
		return bson.D{{Key: "timestamps.createdAt", Value: bson.D{{Key: "$gte", Value: oneMonthAgo}}}}
	case "year":
		oneYearAgo := currentTime.AddDate(-1, 0, 0)
		return bson.D{{Key: "timestamps.createdAt", Value: bson.D{{Key: "$gte", Value: oneYearAgo}}}}
	default:
		return bson.D{}
	}
}

func (c *ClipRepository) getSortStage(filterType string) bson.D {
	switch filterType {
	case "most_viewed":
		return bson.D{{Key: "$sort", Value: bson.D{{Key: "views", Value: -1}}}}
	case "recent":
		return bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamps.createdAt", Value: -1}}}}
	case "random":
		return bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}} // Muestra aleatoria
	default:
		return bson.D{{Key: "$sort", Value: bson.D{{Key: "timestamps.createdAt", Value: -1}}}}

	}
}
func (c *ClipRepository) DeleteClipByIDAndUserID(clipID, userID primitive.ObjectID) error {
	// Definir los criterios de búsqueda
	ctx := context.Background()
	filter := bson.M{
		"_id":    clipID,
		"UserID": userID,
	}

	Database := c.mongoClient.Database("PINKKER-BACKEND")
	result, err := Database.Collection("Clips").DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no se encontró ningún clip con el ID y UserID especificados")
	}

	return nil
}

// Actualizar el título de un clip por ID y UserID
func (c *ClipRepository) UpdateClipTitle(clipID, userID primitive.ObjectID, newTitle string) error {

	ctx := context.Background()
	filter := bson.M{
		"_id":    clipID,
		"UserID": userID,
	}

	// Definir los datos a actualizar
	update := bson.M{
		"$set": bson.M{
			"ClipTitle": newTitle,
		},
	}

	result, err := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips").UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("no se encontró ningún clip con el ID y UserID especificados")
	}

	return nil
}

func (c *ClipRepository) TimeOutClipCreate(id primitive.ObjectID) error {
	key := "ClipCreate" + id.Hex()
	value := id.Hex()

	set, err := c.redisClient.SetNX(context.TODO(), key, value, 3*time.Minute).Result()
	if err != nil {
		return fmt.Errorf("failed to set key in redis: %w", err)
	}

	if !set {
		return fmt.Errorf("key already exists in redis: %s", key)
	}

	return nil
}

func (c *ClipRepository) GetClipsByTitle(title string, limit int) ([]clipdomain.GetClip, error) {
	ctx := context.Background()
	clipsDB := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	// Crear un índice en el campo ClipTitle
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "ClipTitle", Value: "text"}},
	}
	_, err := clipsDB.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"ClipTitle": primitive.Regex{Pattern: title, Options: "i"},
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
		}}},
		bson.D{{Key: "$limit", Value: limit}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "likeCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "id", Value: "$_id"},
			{Key: "NameUserCreator", Value: 1},
			{Key: "IDCreator", Value: 1},
			{Key: "NameUser", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "Category", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Avatar", Value: 1},
			{Key: "ClipTitle", Value: 1},
			{Key: "url", Value: 1},
			{Key: "duration", Value: 1},
			{Key: "views", Value: 1},
			{Key: "cover", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "timestamps", Value: 1},
		}}},
	}

	cursor, err := clipsDB.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clips []clipdomain.GetClip
	for cursor.Next(ctx) {
		var clip clipdomain.GetClip
		if err := cursor.Decode(&clip); err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return clips, nil
}

// getUser obtiene un usuario de la colección Users
func (c *ClipRepository) getUser(ctx context.Context, idT primitive.ObjectID, UsersDb *mongo.Collection) (*userdomain.GetUser, error) {
	var user *userdomain.GetUser
	err := UsersDb.FindOne(ctx, bson.D{{Key: "_id", Value: idT}}).Decode(&user)
	return user, err
}

// getFollowingIDs obtiene los IDs de los usuarios que sigue el usuario
func (c *ClipRepository) getFollowingIDs(idT primitive.ObjectID, UsersDB *mongo.Collection, ctx context.Context) ([]primitive.ObjectID, error) {
	userPipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.M{"_id": idT}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Following", Value: bson.D{{Key: "$objectToArray", Value: "$Following"}}},
		}}},
		bson.D{{Key: "$unwind", Value: "$Following"}},
		bson.D{{Key: "$limit", Value: 100}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Following.k", Value: 1},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{}},
			{Key: "followingIDs", Value: bson.D{{Key: "$push", Value: "$Following.k"}}},
		}}},
	}

	cursor, err := UsersDB.Aggregate(ctx, userPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var userResult struct {
		FollowingIDs []primitive.ObjectID `bson:"followingIDs"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&userResult); err != nil {
			return nil, err
		}
	} else {
		// No se encontró ningún seguidor o la lista está vacía
		return []primitive.ObjectID{}, nil
	}

	return userResult.FollowingIDs, nil
}

// ClipRepository obtiene los IDs de los clips que deben ser excluidos
func (t *ClipRepository) getExcludedIDs(excludeIDs []primitive.ObjectID) []interface{} {
	excludedIDs := make([]interface{}, len(excludeIDs))
	for i, id := range excludeIDs {
		excludedIDs[i] = id
	}
	return excludedIDs
}

func (c *ClipRepository) getFirstFourCategories(user *userdomain.GetUser) []string {
	var categories []string

	// Obtener las categorías ordenadas
	for category := range user.CategoryPreferences {
		categories = append(categories, category)
		if len(categories) == 4 {
			break
		}
	}

	if len(categories) < 5 {
		for len(categories) < 5 {
			categories = append(categories, "nothing")
		}
	}

	// Elegir una categoría aleatoria si hay al menos una categoría
	var randomCategory string
	if len(user.CategoryPreferences) > 0 {
		rand.Seed(time.Now().UnixNano()) // Inicializar el generador de números aleatorios
		allCategories := make([]string, 0, len(user.CategoryPreferences))
		for category := range user.CategoryPreferences {
			allCategories = append(allCategories, category)
		}
		randomCategory = allCategories[rand.Intn(len(allCategories))]
	} else {
		randomCategory = "nothing"
	}

	// Añadir la categoría aleatoria a la lista
	if len(categories) > 0 {
		categories = append(categories[:len(categories)-1], randomCategory)
	} else {
		categories = append(categories, randomCategory)
	}

	return categories
}

// getRelevantClips obtiene los clips relevantes basados en los seguidores y categorías del usuario
func (c *ClipRepository) getRelevantClips(ctx context.Context, clipsDB *mongo.Collection, followingIDs []primitive.ObjectID, excludeFilter bson.D, categories []string, limit int, idT primitive.ObjectID) ([]clipdomain.GetClip, error) {
	timeLimit := time.Now().Add(-72 * time.Hour)
	pipeline := mongo.Pipeline{
		// Filtrar por categorías y clips creados en las últimas 48 horas
		bson.D{{Key: "$match", Value: bson.M{
			"timestamps.createdAt": bson.M{"$gte": timeLimit},
			"Category":             bson.M{"$in": categories},
		}}},
		bson.D{{Key: "$match", Value: bson.M{"Type": bson.M{"$ne": "Ad"}}}},
		// Aplicar filtro adicional para excluir ciertos clips
		bson.D{{Key: "$match", Value: excludeFilter}},
		// Agregar campos auxiliares
		bson.D{{Key: "isFollowingUser", Value: bson.D{
			{Key: "$in", Value: bson.A{"$UserID", bson.D{
				{Key: "$ifNull", Value: bson.A{followingIDs, bson.A{}}},
			}}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likedByFollowing", Value: bson.D{{Key: "$setIntersection", Value: bson.A{"$Likes", followingIDs}}}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idT, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},
		// Calcular el factor de relevancia
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "relevanceFactor", Value: bson.D{{Key: "$add", Value: bson.A{
				// Ponderar más fuertemente los clips de los usuarios seguidos
				bson.D{{Key: "$multiply", Value: bson.A{bson.D{{Key: "$cond", Value: bson.D{
					{Key: "if", Value: "$isFollowingUser"},
					{Key: "then", Value: 5}, // Mayor ponderación para los clips de usuarios seguidos
					{Key: "else", Value: 0},
				}}}, 3}}},
				// Ponderar más fuertemente los "me gusta" de los usuarios seguidos
				bson.D{{Key: "$multiply", Value: bson.A{bson.D{{Key: "$size", Value: "$likedByFollowing"}}, 15}}},
				// Frescura del clip
				bson.D{{Key: "$subtract", Value: bson.A{1000, bson.D{{Key: "$divide", Value: bson.A{bson.D{{Key: "$subtract", Value: bson.A{time.Now(), "$timestamps.createdAt"}}}, 3600000}}}}}},
			}}}},
		}}},
		// Ordenar por factor de relevancia en orden descendente
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "relevanceFactor", Value: -1},
		}}},
		// Limitar el número de resultados
		bson.D{{Key: "$limit", Value: limit}},
		// Proyección de los campos necesarios
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "likeCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "isLikedByID", Value: 1},

			{Key: "id", Value: "$_id"},
			{Key: "NameUserCreator", Value: 1},
			{Key: "IDCreator", Value: 1},
			{Key: "NameUser", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "Category", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Avatar", Value: 1},
			{Key: "ClipTitle", Value: 1},
			{Key: "url", Value: 1},
			{Key: "Likes", Value: 1},
			{Key: "duration", Value: 1},
			{Key: "views", Value: 1},
			{Key: "cover", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "timestamps", Value: 1},
		}}},
	}

	cursor, err := clipsDB.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clips []clipdomain.GetClip
	for cursor.Next(ctx) {
		var clip clipdomain.GetClip
		if err := cursor.Decode(&clip); err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}

	return clips, nil
}

func (c *ClipRepository) getRandomClips(ctx context.Context, excludeFilter bson.D, limit int, clipsDB *mongo.Collection, idT primitive.ObjectID) ([]clipdomain.GetClip, error) {

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			// {Key: "timestamps.createdAt", Value: bson.D{{Key: "$gte", Value: timeLimit}}},
		}}},
		bson.D{{Key: "$match", Value: excludeFilter}},
		// Agregar campos auxiliares
		bson.D{{Key: "$match", Value: bson.M{"Type": bson.M{"$ne": "Ad"}}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idT, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},
		// Calcular el factor de relevancia
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "relevanceScore", Value: bson.D{{Key: "$add", Value: bson.A{
				// Ponderar más fuertemente los likes
				bson.D{{Key: "$multiply", Value: bson.A{"$likeCount", 5}}},
				// Frescura del clip
				bson.D{{Key: "$subtract", Value: bson.A{1000, bson.D{{Key: "$divide", Value: bson.A{bson.D{{Key: "$subtract", Value: bson.A{time.Now(), "$timestamps.createdAt"}}}, 86400000}}}}}},
			}}}},
		}}},
		// Ordenar por factor de relevancia en orden descendente
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "relevanceScore", Value: -1},
		}}},
		// Limitar el número de resultados
		bson.D{{Key: "$limit", Value: limit}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},

			{Key: "likeCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "isLikedByID", Value: 1},

			{Key: "NameUserCreator", Value: 1},
			{Key: "IDCreator", Value: 1},
			{Key: "NameUser", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "Category", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Avatar", Value: 1},
			{Key: "ClipTitle", Value: 1},
			{Key: "url", Value: 1},
			{Key: "Likes", Value: 1},
			{Key: "duration", Value: 1},
			{Key: "views", Value: 1},
			{Key: "cover", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "timestamps", Value: 1},
		}}},
	}

	cursor, err := clipsDB.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clips []clipdomain.GetClip
	for cursor.Next(ctx) {
		var clip clipdomain.GetClip
		if err := cursor.Decode(&clip); err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}

	return clips, nil
}

func (c *ClipRepository) ClipsRecommended(idT primitive.ObjectID, limit int, excludeIDs []primitive.ObjectID) ([]clipdomain.GetClip, error) {
	ctx := context.Background()
	Database := c.mongoClient.Database("PINKKER-BACKEND")
	UsersDB := Database.Collection("Users")
	user, err := c.getUser(ctx, idT, UsersDB)
	if err != nil {
		return nil, err
	}

	followingIDs, err := c.getFollowingIDs(idT, UsersDB, ctx)
	if err != nil {
		return nil, err
	}
	excludedIDs := c.getExcludedIDs(excludeIDs)
	excludeFilter := bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: excludedIDs}}}}

	categories := c.getFirstFourCategories(user)
	var recommendedClips []clipdomain.GetClip
	clipsDB := Database.Collection("Clips")
	if len(followingIDs) == 0 {
		return c.getRandomClips(ctx, excludeFilter, limit-len(recommendedClips), clipsDB, idT)
	}

	recommendedClips, err = c.getRelevantClips(ctx, clipsDB, followingIDs, excludeFilter, categories, limit, idT)
	if err != nil {
		recommendedClips = []clipdomain.GetClip{}
	}

	if len(recommendedClips) < limit {
		var recommendedClipIDs []interface{}
		for _, clip := range recommendedClips {
			recommendedClipIDs = append(recommendedClipIDs, clip.ID)
		}

		excludeFilter := bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "$nin", Value: append(excludedIDs, recommendedClipIDs...)},
			}},
		}
		randomClips, err := c.getRandomClips(ctx, excludeFilter, limit-len(recommendedClips), clipsDB, idT)
		if err != nil {

			return nil, err
		}
		recommendedClips = append(recommendedClips, randomClips...)
	}

	return recommendedClips, nil
}

func (c *ClipRepository) SaveClip(clip *clipdomain.Clip) (*clipdomain.Clip, error) {
	database := c.mongoClient.Database("PINKKER-BACKEND")
	clipCollection := database.Collection("Clips")
	userCollection := database.Collection("Users")

	result, err := clipCollection.InsertOne(context.Background(), clip)
	if err != nil {
		return nil, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("no se pudo obtener el ID insertado")
	}
	clip.ID = insertedID

	filterUser := bson.M{"_id": clip.UserID}
	update := bson.M{"$push": bson.M{"Clips": insertedID}}

	opts := options.Update().SetUpsert(false)

	resultuserCollection, err := userCollection.UpdateOne(context.Background(), filterUser, update, opts)
	if err != nil {
		return clip, err
	}

	if resultuserCollection.ModifiedCount == 0 {
		return clip, errors.New("no se encontraron documentos para actualizar")
	}

	return clip, err
}
func (c *ClipRepository) UpdateClip(clipID primitive.ObjectID, newURL string) {
	clipCollection := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	filter := bson.M{"_id": clipID}

	update := bson.M{"$set": bson.M{"url": newURL}}

	opts := options.Update().SetUpsert(false)

	clipCollection.UpdateOne(context.Background(), filter, update, opts)

}
func (c *ClipRepository) UpdateClipPreviouImage(clipID primitive.ObjectID, newURL string) {
	clipCollection := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	filter := bson.M{"_id": clipID}

	update := bson.M{"$set": bson.M{"PreviouImage": newURL}}

	opts := options.Update().SetUpsert(false)

	clipCollection.UpdateOne(context.Background(), filter, update, opts)

}
func (c *ClipRepository) FindClipById(IdClip primitive.ObjectID) (clipdomain.GetClip, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	pipeline := mongo.Pipeline{
		// Match the clip by ID
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: IdClip}}}},

		// Add fields to count likes and comments from arrays
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
		}}},
		// Project the required fields
		{{Key: "$project", Value: bson.D{
			{Key: "AdId", Value: 1},
			{Key: "likeCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "id", Value: "$_id"},
			{Key: "NameUserCreator", Value: 1},
			{Key: "IDCreator", Value: 1},
			{Key: "NameUser", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "Category", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Avatar", Value: 1},
			{Key: "ClipTitle", Value: 1},
			{Key: "url", Value: 1},
			{Key: "Likes", Value: 1},
			{Key: "duration", Value: 1},
			{Key: "views", Value: 1},
			{Key: "cover", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "timestamps", Value: 1},
			{Key: "Type", Value: 1},
		}}},
	}

	// Execute the aggregation pipeline
	cursor, err := GoMongoDBColl.Aggregate(context.Background(), pipeline)
	if err != nil {
		return clipdomain.GetClip{}, err
	}
	defer cursor.Close(context.Background())

	var clip clipdomain.GetClip
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&clip); err != nil {
			return clipdomain.GetClip{}, err
		}
	} else {
		return clipdomain.GetClip{}, fmt.Errorf("clip not found")
	}

	// Return the found clip
	return clip, nil
}
func (c *ClipRepository) GetClipIdLogueado(IdClip, idValueObj primitive.ObjectID) (*clipdomain.GetClip, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	pipeline := mongo.Pipeline{
		// Match the clip by ID
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: IdClip}}}},

		// Add fields to count likes and comments from arrays
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idValueObj, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},
		// Project the required fields
		{{Key: "$project", Value: bson.D{
			{Key: "likeCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "isLikedByID", Value: 1},

			{Key: "id", Value: "$_id"},
			{Key: "NameUserCreator", Value: 1},
			{Key: "IDCreator", Value: 1},
			{Key: "NameUser", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "Category", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Avatar", Value: 1},
			{Key: "ClipTitle", Value: 1},
			{Key: "url", Value: 1},
			{Key: "Likes", Value: 1},
			{Key: "duration", Value: 1},
			{Key: "views", Value: 1},
			{Key: "cover", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "timestamps", Value: 1},
		}}},
	}

	// Execute the aggregation pipeline
	cursor, err := GoMongoDBColl.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var clip clipdomain.GetClip
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&clip); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("clip not found")
	}

	// Return the found clip
	return &clip, nil
}

func (c *ClipRepository) FindUser(totalKey string) (userdomain.User, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	FindUserInDb := bson.D{
		{Key: "KeyTransmission", Value: totalKey},
	}

	projection := bson.D{
		{Key: "NameUser", Value: 1},
		{Key: "Avatar", Value: 1},
		{Key: "_id", Value: 1},
	}

	var user userdomain.User
	err := GoMongoDBCollUsers.FindOne(
		context.Background(),
		FindUserInDb,
		options.FindOne().SetProjection(projection),
	).Decode(&user)

	if err != nil {
		return userdomain.User{}, err
	}

	return user, nil
}

func (c *ClipRepository) FindUserId(FindUserId string) (*userdomain.User, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	FindUserInDb := bson.D{
		{Key: "NameUser", Value: FindUserId},
	}
	var findUserInDbExist *userdomain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&findUserInDbExist)
	return findUserInDbExist, errCollUsers
}
func (c *ClipRepository) FindCategorieStream(StreamerID primitive.ObjectID) (*streamdomain.Stream, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	FindInDb := bson.D{
		{Key: "StreamerID", Value: StreamerID},
	}
	var findStream *streamdomain.Stream
	errCollUsers := GoMongoDBColl.FindOne(context.Background(), FindInDb).Decode(&findStream)
	return findStream, errCollUsers
}
func (c *ClipRepository) GetClipsNameUser(page int, GetClipsNameUser string) ([]clipdomain.GetClip, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	pipeline := mongo.Pipeline{
		// Filtrar por el nombre de usuario
		bson.D{{Key: "$match", Value: bson.D{{Key: "NameUser", Value: GetClipsNameUser}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "TimeStamp", Value: -1}}}},
		bson.D{{Key: "$skip", Value: (page - 1) * 10}},
		bson.D{{Key: "$limit", Value: 10}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "likeCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "id", Value: "$_id"},
			{Key: "NameUserCreator", Value: 1},
			{Key: "IDCreator", Value: 1},
			{Key: "NameUser", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "Category", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Avatar", Value: 1},
			{Key: "ClipTitle", Value: 1},
			{Key: "url", Value: 1},
			{Key: "duration", Value: 1},
			{Key: "views", Value: 1},
			{Key: "cover", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "timestamps", Value: 1},
		}}},
	}

	cursor, err := GoMongoDBColl.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var clips []clipdomain.GetClip
	if err := cursor.All(context.Background(), &clips); err != nil {
		return nil, err
	}

	return clips, nil
}

func (c *ClipRepository) GetClipsCategory(page int, Category string, lastClipID primitive.ObjectID) ([]clipdomain.Clip, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	options := options.Find()
	options.SetSort(bson.D{{Key: "_id", Value: -1}})
	options.SetSkip(int64((page - 1) * 10))
	options.SetLimit(10)

	filter := bson.D{}
	if Category != "" {
		filter = bson.D{{Key: "Category", Value: Category}}
	}

	if !lastClipID.IsZero() {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$lt": lastClipID}})
	}

	cursor, err := GoMongoDBColl.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var clips []clipdomain.Clip
	if err := cursor.All(context.Background(), &clips); err != nil {
		return nil, err
	}

	return clips, nil
}
func (c *ClipRepository) GetClipsMostViewed(page int) ([]clipdomain.Clip, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	options := options.Find()
	options.SetSort(bson.D{{Key: "Views", Value: -1}})
	options.SetSkip(int64((page - 1) * 10))
	options.SetLimit(10)

	// Filtro para excluir clips de tipo "Ad"
	filter := bson.D{
		{Key: "Type", Value: bson.M{"$ne": "Ad"}},
	}

	cursor, err := GoMongoDBColl.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var clips []clipdomain.Clip
	if err := cursor.All(context.Background(), &clips); err != nil {
		return nil, err
	}

	return clips, nil
}

func (c *ClipRepository) GetClipsMostViewedLast48Hours(page int) ([]clipdomain.Clip, error) {
	twoDaysAgo := time.Now().Add(-48 * time.Hour)

	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	options := options.Find()
	options.SetSort(bson.D{{Key: "Views", Value: -1}})
	options.SetSkip(int64((page - 1) * 10))
	options.SetLimit(10)

	filter := bson.D{
		{Key: "timestamps.createdAt", Value: bson.D{{Key: "$gte", Value: twoDaysAgo}}},
	}

	cursor, err := GoMongoDBColl.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var clips []clipdomain.Clip
	if err := cursor.All(context.Background(), &clips); err != nil {
		return nil, err
	}
	return clips, nil
}

func (c *ClipRepository) LikeClip(clipID, userID primitive.ObjectID) error {
	ctx := context.Background()
	GoMongoDB := c.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollClips := GoMongoDB.Collection("Clips")

	// Verificar si el clip existe
	err := GoMongoDBCollClips.FindOne(ctx, bson.M{"_id": clipID}).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("clip with ID %s does not exist", clipID.Hex())
		}
		return fmt.Errorf("error fetching clip: %v", err)
	}

	// Incrementar el contador de likes del clip
	filter := bson.M{"_id": clipID}
	update := bson.M{"$addToSet": bson.M{"Likes": userID}} // Evita duplicados
	_, err = GoMongoDBCollClips.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Agregar el clip a la lista de likes del usuario
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")
	userFilter := bson.M{"_id": userID}
	userUpdate := bson.M{
		"$addToSet": bson.M{"ClipsLikes": clipID}, // Evitar duplicados
	}

	_, err = GoMongoDBCollUsers.UpdateOne(ctx, userFilter, userUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClipRepository) UpdateUserCategoryPreference(userID primitive.ObjectID, clipCategory string) error {
	ctx := context.Background()
	GoMongoDB := c.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")

	// Obtener solo las preferencias de categorías del usuario
	var result struct {
		CategoryPreferences map[string]float64 `bson:"categoryPreferences"`
	}
	err := GoMongoDBCollUsers.FindOne(ctx, bson.M{"_id": userID}, options.FindOne().SetProjection(bson.M{"categoryPreferences": 1})).Decode(&result)
	if err != nil {
		return err
	}

	// Obtener el puntaje actual de la categoría del clip
	categoryPreferences := result.CategoryPreferences
	currentScore, exists := categoryPreferences[clipCategory]
	if !exists {
		// Si la categoría no existe en las preferencias, inicializarla con 0
		currentScore = 0
	}

	// Incrementar el puntaje de la categoría
	currentScore += 1.0

	// Reordenar categorías si el puntaje alcanza 3
	if currentScore >= 3.0 {
		// Reiniciar el puntaje de la categoría
		currentScore = 0.0

		// Crear una lista de categorías con puntajes
		type CategoryScore struct {
			Category string
			Score    float64
		}

		categoryScores := []CategoryScore{}
		for cat, score := range categoryPreferences {
			categoryScores = append(categoryScores, CategoryScore{Category: cat, Score: score})
		}

		// Encontrar la posición actual de la categoría que recibió el like
		var currentPos int
		for i, cs := range categoryScores {
			if cs.Category == clipCategory {
				currentPos = i
				break
			}
		}

		// Mover la categoría un puesto hacia arriba si no está ya en el primer puesto
		if currentPos > 0 {
			// Intercambiar la categoría con la categoría anterior
			categoryScores[currentPos], categoryScores[currentPos-1] = categoryScores[currentPos-1], categoryScores[currentPos]
		}

		// Actualizar el mapa de preferencias basado en el nuevo orden
		newPreferences := make(map[string]float64)
		for i, cs := range categoryScores {
			newPreferences[cs.Category] = float64(i) // Asignar posición como el nuevo puntaje
		}

		// Actualizar en la base de datos
		update := bson.M{
			"$set": bson.M{"categoryPreferences": newPreferences},
		}
		_, err = GoMongoDBCollUsers.UpdateOne(ctx, bson.M{"_id": userID}, update)
		if err != nil {
			return err
		}

		return nil
	}

	// Actualizar el valor en CategoryPreferences si no se realiza un reordenamiento
	update := bson.M{
		"$set": bson.M{
			"categoryPreferences." + clipCategory: currentScore,
		},
	}

	_, err = GoMongoDBCollUsers.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClipRepository) GetClipByID(clipID primitive.ObjectID) (*clipdomain.ClipCategoryInfo, error) {
	ctx := context.Background()
	GoMongoDB := c.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollClips := GoMongoDB.Collection("Clips")

	projection := bson.M{
		"Category": 1,
	}

	var clipCategoryInfo clipdomain.ClipCategoryInfo
	err := GoMongoDBCollClips.FindOne(ctx, bson.M{"_id": clipID}, options.FindOne().SetProjection(projection)).Decode(&clipCategoryInfo)
	if err != nil {
		return nil, err
	}

	return &clipCategoryInfo, nil
}

func (c *ClipRepository) LikeAndUpdateCategory(clipID, userID primitive.ObjectID) error {
	err := c.LikeClip(clipID, userID)
	if err != nil {
		return err
	}

	clip, err := c.GetClipByID(clipID)
	if err != nil {
		return err
	}

	err = c.UpdateUserCategoryPreference(userID, clip.Category)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClipRepository) ClipDislike(ClipId, idValueToken primitive.ObjectID) error {
	GoMongoDB := c.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBColl := GoMongoDB.Collection("Clips")

	err := GoMongoDBColl.FindOne(context.Background(), bson.M{"_id": ClipId}).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("clip with ID  does not exist")
		}
		return fmt.Errorf("error fetching clip")
	}

	filter := bson.D{{Key: "_id", Value: ClipId}}
	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "Likes", Value: idValueToken}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	GoMongoDBColl = GoMongoDB.Collection("Users")

	filter = bson.D{{Key: "_id", Value: idValueToken}}
	update = bson.D{{Key: "$pull", Value: bson.D{{Key: "ClipsLikes", Value: ClipId}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil

}
func (c *ClipRepository) MoreViewOfTheClip(ClipId primitive.ObjectID, idt primitive.ObjectID) error {
	GoMongoDB := c.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBColl := GoMongoDB.Collection("Clips")

	// Crear filtro para encontrar el clip y verificar que el usuario no lo haya visto antes
	filter := bson.M{
		"_id": ClipId, // Filtro por ID del clip
		"$or": []bson.M{
			{"IdOfTheUsersWhoViewed": bson.M{"$ne": idt}},       // Si el usuario no está en el array
			{"IdOfTheUsersWhoViewed": bson.M{"$exists": false}}, // Si el campo no existe
			{"IdOfTheUsersWhoViewed": bson.M{"$eq": nil}},       // Si el campo es null
		},
	}

	// Actualización para inicializar el campo IdOfTheUsersWhoViewed si es null o no existe, luego incrementar las vistas
	update := bson.M{
		"$set": bson.M{
			"IdOfTheUsersWhoViewed": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$or": bson.A{bson.M{"$eq": bson.A{"$IdOfTheUsersWhoViewed", nil}}, bson.M{"$not": bson.M{"$exists": "$IdOfTheUsersWhoViewed"}}}},
					"then": bson.A{}, // Si es null o no existe, lo inicializamos como un array vacío
					"else": "$IdOfTheUsersWhoViewed",
				},
			},
		},
		"$push": bson.M{
			"IdOfTheUsersWhoViewed": bson.M{
				"$each":     []primitive.ObjectID{idt}, // Agregar el ID del usuario actual
				"$position": -1,                        // Añadir al final del array
				"$slice":    -20,                       // Mantener solo los últimos 20 IDs
			},
		},
		"$inc": bson.M{
			"views": 1, // Incrementar el contador de vistas
		},
	}

	// Actualizar el documento del clip en la colección
	_, err := GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClipRepository) CommentClip(clipID, userID primitive.ObjectID, username, comment string) (clipdomain.ClipCommentGet, error) {
	ctx := context.Background()
	db := c.mongoClient.Database("PINKKER-BACKEND")

	clipCollection := db.Collection("Clips")
	var clip struct {
		Type string `bson:"Type"`
	}
	err := clipCollection.FindOne(ctx, bson.M{"_id": clipID}, options.FindOne().SetProjection(bson.M{"Type": 1})).Decode(&clip)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return clipdomain.ClipCommentGet{}, errors.New("el clip no existe")
		}
		return clipdomain.ClipCommentGet{}, err
	}

	if clip.Type == "Ad" {
		return clipdomain.ClipCommentGet{}, errors.New("no se puede comentar en clips de tipo Ad")
	}

	commentCollection := db.Collection("CommentsClips")
	commentDoc := clipdomain.ClipComment{
		ClipID:    clipID,
		UserID:    userID,
		NameUser:  username,
		Comment:   comment,
		CreatedAt: time.Now(),
		Likes:     []primitive.ObjectID{},
	}
	insertResult, err := commentCollection.InsertOne(ctx, commentDoc)
	if err != nil {
		return clipdomain.ClipCommentGet{}, err
	}

	// Actualizar la información del usuario con el ID del comentario creado
	userCollection := db.Collection("Users")
	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "ClipsComment", Value: insertResult.InsertedID}}}}
	_, err = userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return clipdomain.ClipCommentGet{}, err
	}

	// Obtener el comentario con los datos del usuario
	cursor, err := commentCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"_id": insertResult.InsertedID}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},

		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{userID, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},
		// Project the required fields
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "likeCount", Value: 1},
			{Key: "isLikedByID", Value: 1},
			{Key: "clipId", Value: 1},
			{Key: "UserID", Value: "$UserID"},
			{Key: "FullName", Value: "$UserInfo.FullName"},
			{Key: "Avatar", Value: "$UserInfo.Avatar"},
			{Key: "comment", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "Likes", Value: 1},
			{Key: "nameUser", Value: 1},
		}}},
	})
	if err != nil {
		return clipdomain.ClipCommentGet{}, err
	}
	defer cursor.Close(ctx)

	var comments []clipdomain.ClipCommentGet
	for cursor.Next(ctx) {
		var comment clipdomain.ClipCommentGet
		if err := cursor.Decode(&comment); err != nil {
			return clipdomain.ClipCommentGet{}, err
		}
		comments = append(comments, comment)
	}
	if err := cursor.Err(); err != nil {
		return clipdomain.ClipCommentGet{}, err
	}

	if len(comments) > 0 {
		return comments[0], nil
	}

	return clipdomain.ClipCommentGet{}, errors.New("no se encontró ningún comentario")
}

func (c *ClipRepository) LikeComment(commentID, userID primitive.ObjectID) error {
	ctx := context.Background()
	db := c.mongoClient.Database("PINKKER-BACKEND")

	commentCollection := db.Collection("CommentsClips")
	filter := bson.M{"_id": commentID}
	update := bson.M{"$addToSet": bson.M{"Likes": userID}}

	_, err := commentCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
func (c *ClipRepository) UnlikeComment(commentID, userID primitive.ObjectID) error {
	ctx := context.Background()
	db := c.mongoClient.Database("PINKKER-BACKEND")

	commentCollection := db.Collection("CommentsClips")
	filter := bson.M{"_id": commentID}
	update := bson.M{"$pull": bson.M{"Likes": userID}}
	_, err := commentCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
func (c *ClipRepository) GetClipComments(clipID primitive.ObjectID, page int) ([]clipdomain.ClipCommentGet, error) {
	ctx := context.Background()
	db := c.mongoClient.Database("PINKKER-BACKEND")

	// Colección de comentarios de clips
	commentCollection := db.Collection("CommentsClips")

	// Filtrar comentarios por ID de clip
	filter := bson.M{"clipId": clipID}

	// Calcular la cantidad de documentos para omitir
	skip := (page - 1) * 15

	// Consultar la base de datos para obtener los comentarios del clip, paginados
	cursor, err := commentCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
		}}},
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "likeCount", Value: 1},
			{Key: "isLikedByID", Value: 1},
			{Key: "clipId", Value: 1},
			{Key: "UserID", Value: "$UserID"},
			{Key: "FullName", Value: "$UserInfo.FullName"},
			{Key: "Avatar", Value: "$UserInfo.Avatar"},
			{Key: "comment", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "Likes", Value: 1},
			{Key: "nameUser", Value: 1},
		}}},
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: 15}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Iterar sobre los documentos devueltos y decodificarlos en la estructura ClipComment
	var comments []clipdomain.ClipCommentGet
	for cursor.Next(ctx) {
		var comment clipdomain.ClipCommentGet
		if err := cursor.Decode(&comment); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (c *ClipRepository) GetClipCommentsLoguedo(clipID primitive.ObjectID, page int, idt primitive.ObjectID) ([]clipdomain.ClipCommentGet, error) {
	ctx := context.Background()
	db := c.mongoClient.Database("PINKKER-BACKEND")

	// Colección de comentarios de clips
	commentCollection := db.Collection("CommentsClips")

	// Filtrar comentarios por ID de clip
	filter := bson.M{"clipId": clipID}

	// Calcular la cantidad de documentos para omitir
	skip := (page - 1) * 15

	// Consultar la base de datos para obtener los comentarios del clip, paginados
	cursor, err := commentCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idt, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "likeCount", Value: 1},
			{Key: "isLikedByID", Value: 1},
			{Key: "clipId", Value: 1},
			{Key: "UserID", Value: "$UserID"},
			{Key: "FullName", Value: "$UserInfo.FullName"},
			{Key: "Avatar", Value: "$UserInfo.Avatar"},
			{Key: "comment", Value: 1},
			{Key: "createdAt", Value: 1},
			{Key: "Likes", Value: 1},
			{Key: "nameUser", Value: 1},
		}}},
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: 15}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Iterar sobre los documentos devueltos y decodificarlos en la estructura ClipComment
	var comments []clipdomain.ClipCommentGet
	for cursor.Next(ctx) {
		var comment clipdomain.ClipCommentGet
		if err := cursor.Decode(&comment); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (t *ClipRepository) GetAdClips() (primitive.ObjectID, error) {
	db := t.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollAdvertisements := db.Collection("Advertisements")
	ctx := context.TODO()

	pipelineRandom := bson.A{
		bson.M{"$match": bson.M{
			"Destination": "ClipAds",
			"State":       "accepted",
			"$expr": bson.M{
				"$lte": bson.A{"$Impressions", "$ImpressionsMax"},
			},
		}},
		bson.M{"$sample": bson.M{"size": 1}},
		bson.M{"$project": bson.M{
			"ClipId": 1,
			"_id":    1,
		}},
	}

	var advertisement struct {
		ClipId primitive.ObjectID `bson:"ClipId"`
		ID     primitive.ObjectID `bson:"_id"`
	}

	cursor, err := GoMongoDBCollAdvertisements.Aggregate(ctx, pipelineRandom)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		if err := cursor.Decode(&advertisement); err != nil {
			return primitive.ObjectID{}, err
		}

		// Incrementar impresiones y actualizar registros de impresiones diarias
		currentDate := time.Now().Format("2006-01-02")
		advertisementFilter := bson.M{"_id": advertisement.ID}

		// Actualización para incrementar impresiones
		updateImpressions := bson.M{
			"$inc": bson.M{
				"Impressions": 1,
			},
		}

		// Ejecutar la actualización de impresiones
		_, err := GoMongoDBCollAdvertisements.UpdateOne(ctx, advertisementFilter, updateImpressions)
		if err != nil {
			return primitive.ObjectID{}, err
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
			return primitive.ObjectID{}, err
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

			_, err = GoMongoDBCollAdvertisements.UpdateOne(ctx, advertisementFilter, newDateUpdate)
			if err != nil {
				return primitive.ObjectID{}, err
			}
		}
		err = t.updatePinkkerProfitPerMonth(ctx)
		// Retornar ClipId después de manejar las impresiones
		return advertisement.ClipId, err
	}

	return primitive.ObjectID{}, errors.New("no advertisements found")
}

func (r *ClipRepository) AddAds(idValueObj primitive.ObjectID, ids []clipdomain.GetClip) error {
	ctx := context.Background()

	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")

	// Verificar si hay 10 clips
	if len(ids) == 10 {
		// Crear un mapa para contar cuántas veces aparece cada IDCreator
		pixelIncrementMap := make(map[primitive.ObjectID]int)

		// Contar las apariciones de cada IDCreator
		for _, clip := range ids {
			pixelIncrementMap[clip.IDCreator]++
		}

		// Iniciar una sesión para realizar la operación en una transacción
		session, err := GoMongoDB.Client().StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		// Usar la sesión para ejecutar la transacción
		_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			// Iterar sobre el mapa de incrementos y ejecutar las actualizaciones
			for creatorID, pixelCount := range pixelIncrementMap {
				userFilter := bson.M{"_id": creatorID}
				updatePixel := bson.M{
					"$inc": bson.M{
						"Pixeles": pixelCount, // Incrementar según el número de ocurrencias
					},
				}

				// Actualizar el Pixeles para cada creador dentro de la transacción
				_, err := GoMongoDBCollUsers.UpdateOne(sessCtx, userFilter, updatePixel)
				if err != nil {
					return nil, err
				}
			}

			// Si todo salió bien, devolver un valor de éxito
			return nil, nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ClipRepository) updatePinkkerProfitPerMonth(ctx context.Context) error {
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

	// Paso 2: Incrementa los valores en 'weeks.week_x'
	monthlyUpdate := bson.M{
		"$inc": bson.M{
			"total":                                 AdvertisementsPayPerPrintFloat,
			"weeks." + currentWeek + ".impressions": impressions,
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
