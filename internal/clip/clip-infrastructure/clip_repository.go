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
func (c *ClipRepository) ClipsRecommended(idT primitive.ObjectID, limit int, excludeIDs []primitive.ObjectID) ([]clipdomain.Clip, error) {
	ctx := context.Background()
	GoMongoDB := c.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollClip := GoMongoDB.Collection("Clips")
	GoMongoDBCollUser := GoMongoDB.Collection("Users")

	// Obtener el usuario
	var findUser *userdomain.User
	err := GoMongoDBCollUser.FindOne(ctx, bson.D{{Key: "_id", Value: idT}}).Decode(&findUser)
	if err != nil {
		return nil, err
	}

	if findUser.Following == nil || len(findUser.Following) == 0 {
		findUser.Following = map[primitive.ObjectID]userdomain.FollowInfo{}
	}

	// Obtener las primeras 4 categorías del usuario

	firstFourCategories := make([]string, 0, 4)
	found := false

	for category := range findUser.CategoryPreferences {
		firstFourCategories = append(firstFourCategories, category)
		if len(firstFourCategories) == 4 {
			found = true
			break
		}
	}

	if !found {
		firstFourCategories = append(firstFourCategories, "nothing")
	}
	var followingIDs []primitive.ObjectID
	if len(findUser.Following) == 0 {
		followingIDs = make([]primitive.ObjectID, 0)
	} else {
		for userID := range findUser.Following {
			followingIDs = append(followingIDs, userID)
		}
	}
	excludedIDs := make([]interface{}, len(excludeIDs))
	for i, id := range excludeIDs {
		excludedIDs[i] = id
	}
	excludeFilter := bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: excludedIDs}}}}

	// Agregar el filtro de exclusión al pipeline
	pipelineFirstFour := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "likes", Value: bson.D{{Key: "$in", Value: followingIDs}}},
			{Key: "timestamps.createdAt", Value: bson.D{{Key: "$gte", Value: time.Now().Add(-23 * time.Hour)}}},
			{Key: "category", Value: bson.D{{Key: "$in", Value: firstFourCategories}}}, // Filtrar por las primeras 4 categorías
		}}},
		bson.D{{Key: "$match", Value: excludeFilter}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "relevanceFactor", Value: bson.D{
				{Key: "$multiply", Value: []interface{}{
					bson.D{{Key: "$log10", Value: "$views"}}, // Aplicar logaritmo a las vistas para suavizar el efecto
					bson.D{{Key: "$size", Value: bson.D{{Key: "$setIntersection", Value: []interface{}{"$likes", followingIDs}}}}}, // Obtener la cantidad de likes compartidos
				}},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "relevanceFactor", Value: -1}}}},
		bson.D{{Key: "$limit", Value: limit}},
	}

	cursorFirstFour, err := GoMongoDBCollClip.Aggregate(ctx, pipelineFirstFour)
	if err != nil {
		return nil, err
	}
	defer cursorFirstFour.Close(ctx)

	// Recopilar los clips de las primeras 4 categorías
	var recommendedClips []clipdomain.Clip
	for cursorFirstFour.Next(ctx) {
		var clip clipdomain.Clip
		err := cursorFirstFour.Decode(&clip)
		if err != nil {
			return nil, err
		}
		recommendedClips = append(recommendedClips, clip)
	}
	// Crear el pipeline de agregación para obtener clips de categorías distintas a las primeras 4 categorías
	pipelineRandom := mongo.Pipeline{
		// bson.D{{Key: "$match", Value: bson.D{
		//	{Key: "likes", Value: bson.D{{Key: "$in", Value: followingIDs}}},
		//	{Key: "timestamps.createdAt", Value: bson.D{{Key: "$gte", Value: time.Now().Add(-23 * time.Hour)}}},
		//	{Key: "category", Value: bson.D{{Key: "$nin", Value: firstFourCategories}}}, // Filtrar por categorías distintas a las primeras 4
		// }}},
		bson.D{{Key: "$match", Value: excludeFilter}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "relevanceFactor", Value: bson.D{
				{Key: "$multiply", Value: []interface{}{
					bson.D{{Key: "$log10", Value: "$views"}},
					bson.D{{Key: "$size", Value: bson.D{{Key: "$setIntersection", Value: []interface{}{"$likes", followingIDs}}}}},
				}},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "relevanceFactor", Value: -1}}}},
		bson.D{{Key: "$limit", Value: limit - len(recommendedClips)}}, // Limitar la cantidad de clips devueltos por categorías distintas
	}
	// Ejecutar el pipeline de agregación para obtener clips de categorías distintas
	cursorRandom, err := GoMongoDBCollClip.Aggregate(ctx, pipelineRandom)
	if err != nil {
		return nil, err
	}

	defer cursorRandom.Close(ctx)

	// Recopilar los clips de categorías distintas
	for cursorRandom.Next(ctx) {
		var clip clipdomain.Clip
		err := cursorRandom.Decode(&clip)
		if err != nil {
			return nil, err
		}
		recommendedClips = append(recommendedClips, clip)
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
		return clip, errors.New("No se encontraron documentos para actualizar.")
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
func (c *ClipRepository) FindrClipId(IdClip primitive.ObjectID) (*clipdomain.Clip, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")
	FindClipInDb := bson.D{
		{Key: "_id", Value: IdClip},
	}

	var findClipInDbExist *clipdomain.Clip
	err := GoMongoDBColl.FindOne(context.Background(), FindClipInDb).Decode(&findClipInDbExist)
	if err != nil {
		return nil, err
	}

	update := bson.D{
		{Key: "$inc", Value: bson.D{{Key: "views", Value: 1}}},
	}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), FindClipInDb, update)
	if err != nil {
		return nil, err
	}

	err = GoMongoDBColl.FindOne(context.Background(), FindClipInDb).Decode(&findClipInDbExist)
	if err != nil {
		return nil, err
	}

	return findClipInDbExist, nil
}

func (c *ClipRepository) FindUser(totalKey string) (*userdomain.User, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var FindUserInDb primitive.D
	FindUserInDb = bson.D{
		{Key: "KeyTransmission", Value: totalKey},
	}
	var findUserInDbExist *userdomain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&findUserInDbExist)
	return findUserInDbExist, errCollUsers
}
func (c *ClipRepository) FindUserId(FindUserId string) (*userdomain.User, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var FindUserInDb primitive.D
	FindUserInDb = bson.D{
		{Key: "NameUser", Value: FindUserId},
	}
	var findUserInDbExist *userdomain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&findUserInDbExist)
	return findUserInDbExist, errCollUsers
}
func (c *ClipRepository) FindCategorieStream(StreamerID primitive.ObjectID) (*streamdomain.Stream, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Streams")
	var FindInDb primitive.D
	FindInDb = bson.D{
		{Key: "StreamerID", Value: StreamerID},
	}
	var findStream *streamdomain.Stream
	errCollUsers := GoMongoDBColl.FindOne(context.Background(), FindInDb).Decode(&findStream)
	return findStream, errCollUsers
}
func (c *ClipRepository) GetClipsNameUser(page int, GetClipsNameUser string) ([]clipdomain.Clip, error) {
	GoMongoDBColl := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	options := options.Find()
	options.SetSort(bson.D{{"TimeStamp", -1}})
	options.SetSkip(int64((page - 1) * 10))
	options.SetLimit(10)
	filter := bson.D{{"NameUserCreator", GetClipsNameUser}}

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
	options.SetSort(bson.D{{"_id", -1}})
	options.SetSkip(int64((page - 1) * 10))
	options.SetLimit(10)

	filter := bson.D{}
	if Category != "" {
		filter = bson.D{{"Category", Category}}
	}

	if !lastClipID.IsZero() {
		filter = append(filter, bson.E{"_id", bson.M{"$lt": lastClipID}})
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
	options.SetSort(bson.D{{"Views", -1}})
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

func (c *ClipRepository) CommentClip(clipID, userID primitive.ObjectID, username, comment string) error {
	ctx := context.Background()
	db := c.mongoClient.Database("PINKKER-BACKEND")

	clipCollection := db.Collection("Clips")
	count, err := clipCollection.CountDocuments(ctx, bson.M{"_id": clipID})
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("el clip no existe")
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
		return err
	}

	// Actualizar la información del usuario con el ID del comentario creado
	userCollection := db.Collection("Users")
	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "ClipsComment", Value: insertResult.InsertedID}}}}
	_, err = userCollection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClipRepository) LikeComment(commentID, userID primitive.ObjectID) error {
	ctx := context.Background()
	db := c.mongoClient.Database("PINKKER-BACKEND")

	commentCollection := db.Collection("CommentsClips")
	filter := bson.M{"_id": commentID}
	update := bson.M{"$addToSet": bson.M{"likes": userID}}

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
	update := bson.M{"$pull": bson.M{"likes": userID}}
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
			{Key: "likes", Value: 1},
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
