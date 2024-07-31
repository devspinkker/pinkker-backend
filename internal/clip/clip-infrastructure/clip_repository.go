// clip-infrastructure/clip_repository.go
package clipinfrastructure

import (
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"errors"
	"fmt"
	"sort"
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
func (c *ClipRepository) GetClipsByTitle(title string, limit int) ([]clipdomain.Clip, error) {
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

	findOptions := options.Find().SetLimit(int64(limit))

	cursor, err := clipsDB.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clips []clipdomain.Clip
	for cursor.Next(ctx) {
		var clip clipdomain.Clip
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
func (c *ClipRepository) getUser(ctx context.Context, idT primitive.ObjectID, UsersDb *mongo.Collection) (*userdomain.User, error) {
	var user *userdomain.User
	err := UsersDb.FindOne(ctx, bson.D{{Key: "_id", Value: idT}}).Decode(&user)
	return user, err
}

// getFollowingIDs obtiene los IDs de los usuarios que sigue el usuario
func (c *ClipRepository) getFollowingIDs(user *userdomain.User) []primitive.ObjectID {
	var followingIDs []primitive.ObjectID
	for id := range user.Following {
		followingIDs = append(followingIDs, id)
	}
	return followingIDs
}

// ClipRepository obtiene los IDs de los clips que deben ser excluidos
func (t *ClipRepository) getExcludedIDs(excludeIDs []primitive.ObjectID) []interface{} {
	excludedIDs := make([]interface{}, len(excludeIDs))
	for i, id := range excludeIDs {
		excludedIDs[i] = id
	}
	return excludedIDs
}

// getFirstFourCategories obtiene las primeras cuatro categorías de preferencias del usuario
func (c *ClipRepository) getFirstFourCategories(user *userdomain.User) []string {
	var categories []string
	for category := range user.CategoryPreferences {
		categories = append(categories, category)
		if len(categories) == 4 {
			break
		}
	}
	if len(categories) == 0 {
		categories = append(categories, "nothing")
	}
	return categories
}

// getRelevantClips obtiene los clips relevantes basados en los seguidores y categorías del usuario
func (c *ClipRepository) getRelevantClips(ctx context.Context, clipsDB *mongo.Collection, followingIDs []primitive.ObjectID, excludeFilter bson.D, categories []string, limit int) ([]clipdomain.Clip, error) {
	timeLimit := time.Now().Add(-72 * time.Hour)
	pipeline := mongo.Pipeline{
		// Filtrar por categorías y clips creados en las últimas 48 horas
		bson.D{{Key: "$match", Value: bson.M{
			"Category":             bson.M{"$in": categories},
			"timestamps.createdAt": bson.M{"$gte": timeLimit},
		}}},
		// Aplicar filtro adicional para excluir ciertos clips
		bson.D{{Key: "$match", Value: excludeFilter}},
		// Agregar campos auxiliares
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "isFollowingUser", Value: bson.D{{Key: "$in", Value: bson.A{"$UserID", followingIDs}}}},
			{Key: "likedByFollowing", Value: bson.D{{Key: "$setIntersection", Value: bson.A{"$Likes", followingIDs}}}},
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

	var clips []clipdomain.Clip
	for cursor.Next(ctx) {
		var clip clipdomain.Clip
		if err := cursor.Decode(&clip); err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}

	return clips, nil
}

func (c *ClipRepository) getRandomClips(ctx context.Context, excludeFilter bson.D, limit int, clipsDB *mongo.Collection) ([]clipdomain.Clip, error) {

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			// {Key: "timestamps.createdAt", Value: bson.D{{Key: "$gte", Value: timeLimit}}},
		}}},
		bson.D{{Key: "$match", Value: excludeFilter}},
		// Agregar campos auxiliares
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
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
	}

	cursor, err := clipsDB.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clips []clipdomain.Clip
	for cursor.Next(ctx) {
		var clip clipdomain.Clip
		if err := cursor.Decode(&clip); err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}

	return clips, nil
}

func (c *ClipRepository) ClipsRecommended(idT primitive.ObjectID, limit int, excludeIDs []primitive.ObjectID) ([]clipdomain.Clip, error) {
	ctx := context.Background()
	Database := c.mongoClient.Database("PINKKER-BACKEND")
	UsersDB := Database.Collection("Users")
	user, err := c.getUser(ctx, idT, UsersDB)
	if err != nil {
		return nil, err
	}

	followingIDs := c.getFollowingIDs(user)
	excludedIDs := c.getExcludedIDs(excludeIDs)
	excludeFilter := bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: excludedIDs}}}}

	categories := c.getFirstFourCategories(user)
	var recommendedClips []clipdomain.Clip
	clipsDB := Database.Collection("Clips")
	if len(followingIDs) == 0 {
		return c.getRandomClips(ctx, excludeFilter, limit-len(recommendedClips), clipsDB)
	}

	recommendedClips, err = c.getRelevantClips(ctx, clipsDB, followingIDs, excludeFilter, categories, limit)
	if err != nil {
		recommendedClips = []clipdomain.Clip{}
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
		randomClips, err := c.getRandomClips(ctx, excludeFilter, limit-len(recommendedClips), clipsDB)
		if err != nil {
			fmt.Println(err)

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
func (c *ClipRepository) FindClipById(IdClip primitive.ObjectID) (*clipdomain.GetClip, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	pipeline := mongo.Pipeline{
		// Match the clip by ID
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: IdClip}}}},

		// Add fields to count likes and comments from arrays
		{{Key: "$addFields", Value: bson.D{
			{Key: "LikeCount", Value: bson.D{{Key: "$size", Value: "$Likes"}}},
			{Key: "CommentCount", Value: bson.D{{Key: "$size", Value: "$Comments"}}},
		}}},

		// Project the required fields
		{{Key: "$project", Value: bson.D{
			{Key: "LikeCount", Value: 1},
			{Key: "CommentCount", Value: 1},
			{Key: "ID", Value: "$_id"},
			{Key: "NameUserCreator", Value: 1},
			{Key: "IDCreator", Value: 1},
			{Key: "NameUser", Value: 1},
			{Key: "StreamThumbnail", Value: 1},
			{Key: "Category", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Avatar", Value: 1},
			{Key: "ClipTitle", Value: 1},
			{Key: "URL", Value: 1},
			{Key: "Duration", Value: 1},
			{Key: "Views", Value: 1},
			{Key: "Cover", Value: 1},
			{Key: "Timestamps", Value: 1},
			{Key: "IsLikedByID", Value: 1},
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

func (c *ClipRepository) FindUser(totalKey string) (*userdomain.User, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	FindUserInDb := bson.D{
		{Key: "KeyTransmission", Value: totalKey},
	}
	var findUserInDbExist *userdomain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&findUserInDbExist)
	return findUserInDbExist, errCollUsers
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
func (c *ClipRepository) GetClipsNameUser(page int, GetClipsNameUser string) ([]clipdomain.Clip, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	options := options.Find()
	options.SetSort(bson.D{{Key: "TimeStamp", Value: -1}})
	options.SetSkip(int64((page - 1) * 10))
	options.SetLimit(10)
	filter := bson.D{{Key: "NameUser", Value: GetClipsNameUser}}

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

	filter := bson.D{}

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

	var clip clipdomain.Clip
	err := GoMongoDBCollClips.FindOne(ctx, bson.M{"_id": clipID}).Decode(&clip)
	if err != nil {
		return err
	}

	// Incrementar el contador de likes del clip
	filter := bson.M{"_id": clipID}
	update := bson.M{"$addToSet": bson.M{"Likes": userID}}
	_, err = GoMongoDBCollClips.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Incrementar el puntaje de la categoría del usuario y sumarle el like dado
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")
	userFilter := bson.M{"_id": userID}
	userUpdate := bson.M{
		"$inc":      bson.M{"categoryPreferences." + clip.Category: 0.01},
		"$addToSet": bson.M{"ClipsLikes": clipID},
	}

	_, err = GoMongoDBCollUsers.UpdateOne(ctx, userFilter, userUpdate)
	if err != nil {
		return err
	}

	// Actualizar las categoryPreferences del usuario para que estén ordenadas según los puntajes
	err = updateCategoryPreferences(ctx, GoMongoDBCollUsers, userID)
	if err != nil {
		return err
	}

	return nil
}

func updateCategoryPreferences(ctx context.Context, collection *mongo.Collection, userID primitive.ObjectID) error {
	// Obtener el usuario
	var user userdomain.User
	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}

	// Ordenar las categorías en la propiedad categoryPreferences
	sortedCategories := make([]string, 0, len(user.CategoryPreferences))
	for category := range user.CategoryPreferences {
		sortedCategories = append(sortedCategories, category)
	}
	sort.Slice(sortedCategories, func(i, j int) bool {
		return user.CategoryPreferences[sortedCategories[i]] > user.CategoryPreferences[sortedCategories[j]]
	})

	// Actualizar la propiedad categoryPreferences con las categorías ordenadas
	update := bson.M{"$set": bson.M{"categoryPreferences": user.CategoryPreferences}}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClipRepository) ClipDislike(ClipId, idValueToken primitive.ObjectID) error {
	GoMongoDB := c.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBColl := GoMongoDB.Collection("Clips")

	count, err := GoMongoDBColl.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: ClipId}})
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("el ClipId no existe")
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
func (c *ClipRepository) MoreViewOfTheClip(ClipId primitive.ObjectID) error {

	GoMongoDB := c.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBColl := GoMongoDB.Collection("Clips")

	filter := bson.D{{Key: "_id", Value: ClipId}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "views", Value: 1}}}}

	_, err := GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	return err
}
func (c *ClipRepository) CommentClip(clipID, userID primitive.ObjectID, username, comment string) (clipdomain.ClipCommentGet, error) {
	ctx := context.Background()
	db := c.mongoClient.Database("PINKKER-BACKEND")

	clipCollection := db.Collection("Clips")
	count, err := clipCollection.CountDocuments(ctx, bson.M{"_id": clipID})
	if err != nil {
		return clipdomain.ClipCommentGet{}, err
	}
	if count == 0 {
		return clipdomain.ClipCommentGet{}, errors.New("el clip no existe")
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
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$project", Value: bson.D{
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
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$project", Value: bson.D{
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
