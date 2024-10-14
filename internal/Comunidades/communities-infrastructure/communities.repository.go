package communitiestinfrastructure

import (
	communitiesdomain "PINKKER-BACKEND/internal/Comunidades/communities"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CommunitiesRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewcommunitiesRepository(redisClient *redis.Client, mongoClient *mongo.Client) *CommunitiesRepository {
	return &CommunitiesRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}

// CreateCommunity crea una nueva comunidad y la guarda en MongoDB
func (repo *CommunitiesRepository) CreateCommunity(ctx context.Context, communityName string, creatorID primitive.ObjectID, description string, isPrivate bool, categories []string) (*communitiesdomain.Community, error) {
	// Crear una nueva comunidad
	var user struct {
		PinkkerPrime *userdomain.PinkkerPrime `bson:"PinkkerPrime"`
	}

	usersCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	err := usersCollection.FindOne(ctx, primitive.M{"_id": creatorID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	if user.PinkkerPrime == nil || user.PinkkerPrime.SubscriptionEnd.Before(time.Now()) {
		return nil, fiber.NewError(fiber.StatusForbidden, "The user does not have an active PinkkerPrime subscription")
	}

	community := &communitiesdomain.Community{
		CommunityName: communityName,
		Description:   description,
		CreatorID:     creatorID,
		Members:       []primitive.ObjectID{creatorID},
		Mods:          []primitive.ObjectID{},
		BannedUsers:   []primitive.ObjectID{},
		Rules:         "Por defecto",
		IsPrivate:     isPrivate,
		Categories:    categories,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	_, err = collection.InsertOne(ctx, community)
	if err != nil {
		return nil, err
	}

	return community, nil
}
func (repo *CommunitiesRepository) AddModerator(ctx context.Context, communityID, newModID, modID primitive.ObjectID) error {
	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	var community struct {
		CreatorID  primitive.ObjectID   `bson:"CreatorID"`
		Moderators []primitive.ObjectID `bson:"Mods"`
	}

	err := collection.FindOne(ctx, primitive.M{"_id": communityID}, options.FindOne().SetProjection(primitive.M{
		"creatorID": 1,
		"CreatorID": 1,
	})).Decode(&community)

	if err != nil {
		return err
	}

	isCreator := community.CreatorID == modID
	isModerator := false
	for _, moderatorID := range community.Moderators {
		if moderatorID == modID {
			isModerator = true
			break
		}
	}

	if !isCreator && !isModerator {
		return fiber.NewError(fiber.StatusForbidden, "You are not authorized to add moderators")
	}

	for _, moderatorID := range community.Moderators {
		if moderatorID == newModID {
			return fiber.NewError(fiber.StatusBadRequest, "User is already a moderator")
		}
	}

	if len(community.Moderators) >= 5 {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot add more than 5 moderators")
	}

	_, err = collection.UpdateOne(
		ctx,
		primitive.M{"_id": communityID},
		primitive.M{
			"$addToSet": primitive.M{"Members": newModID}, // Añadir el nuevo moderador
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (repo *CommunitiesRepository) AddMember(ctx context.Context, communityID primitive.ObjectID, userID primitive.ObjectID) error {
	session, err := repo.mongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		communitiesCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")
		usersCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

		// Verificar si el usuario está en la lista de baneados
		err := communitiesCollection.FindOne(sc, primitive.M{
			"_id":         communityID,
			"BannedUsers": primitive.M{"$in": []primitive.ObjectID{userID}},
		}).Err()

		if err == nil {
			session.AbortTransaction(sc)
			return fiber.NewError(fiber.StatusForbidden, "This user is banned from the community")
		}

		// Si el usuario no está baneado, añadirlo a la comunidad
		_, err = communitiesCollection.UpdateOne(
			sc,
			primitive.M{"_id": communityID},
			primitive.M{"$addToSet": primitive.M{"Members": userID}},
		)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// Añadir la comunidad al usuario
		_, err = usersCollection.UpdateOne(
			sc,
			primitive.M{"_id": userID},
			primitive.M{"$addToSet": primitive.M{"InCommunities": communityID}},
		)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// Commit de la transacción
		if err := session.CommitTransaction(sc); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (repo *CommunitiesRepository) RemoveMember(ctx context.Context, communityID primitive.ObjectID, userID primitive.ObjectID) error {
	session, err := repo.mongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		communitiesCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")
		usersCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

		// Remover el usuario de la comunidad (del campo "Members")
		_, err := communitiesCollection.UpdateOne(
			sc,
			primitive.M{"_id": communityID},
			primitive.M{"$pull": primitive.M{"Members": userID}},
		)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// Remover la comunidad del usuario (del campo "InCommunities")
		_, err = usersCollection.UpdateOne(
			sc,
			primitive.M{"_id": userID},
			primitive.M{"$pull": primitive.M{"InCommunities": communityID}},
		)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// Commit de la transacción
		if err := session.CommitTransaction(sc); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (repo *CommunitiesRepository) BanMember(ctx context.Context, communityID, userID, mod primitive.ObjectID) error {
	session, err := repo.mongoClient.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		communitiesCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")
		usersCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

		var community struct {
			CreatorID  primitive.ObjectID   `bson:"CreatorID"`
			Moderators []primitive.ObjectID `bson:"Moderators"`
		}

		err := communitiesCollection.FindOne(sc, primitive.M{"_id": communityID}, options.FindOne().SetProjection(primitive.M{
			"CreatorID":  1,
			"Moderators": 1,
		})).Decode(&community)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// Verificar si el moderador es autorizado para banear
		isCreator := community.CreatorID == mod
		isModerator := false
		for _, moderatorID := range community.Moderators {
			if moderatorID == mod {
				isModerator = true
				break
			}
		}

		// Verificar si el usuario a banear es un moderador
		isUserModerator := false
		for _, moderatorID := range community.Moderators {
			if moderatorID == userID {
				isUserModerator = true
				break
			}
		}

		// Si el moderador no es ni creador ni moderador, no puede banear
		if !isCreator && !isModerator {
			session.AbortTransaction(sc)
			return fiber.NewError(fiber.StatusForbidden, "You are not authorized to ban members")
		}

		// Si el moderador es un moderador y el usuario a banear es también moderador, denegar la operación
		if isModerator && isUserModerator {
			session.AbortTransaction(sc)
			return fiber.NewError(fiber.StatusForbidden, "Moderators cannot ban other moderators")
		}

		// Solo el creador puede banear a un moderador
		if isUserModerator && !isCreator {
			session.AbortTransaction(sc)
			return fiber.NewError(fiber.StatusForbidden, "Only the creator can ban moderators")
		}

		// Proceder con el baneo si es autorizado
		_, err = communitiesCollection.UpdateOne(
			sc,
			primitive.M{"_id": communityID},
			primitive.M{
				"$pull":     primitive.M{"Members": userID},     // Eliminar de la lista de miembros
				"$addToSet": primitive.M{"BannedUsers": userID}, // Añadir a la lista de baneados
			},
		)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// Remover la comunidad del array InCommunities del usuario
		_, err = usersCollection.UpdateOne(
			sc,
			primitive.M{"_id": userID},
			primitive.M{"$pull": primitive.M{"InCommunities": communityID}}, // Eliminar la comunidad de InCommunities
		)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// Commit de la transacción
		if err := session.CommitTransaction(sc); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (repo *CommunitiesRepository) GetCommunityPosts(ctx context.Context, communityID primitive.ObjectID, ExcludeFilterlID []primitive.ObjectID, idT primitive.ObjectID, limit int) ([]communitiesdomain.PostGetCommunityReq, error) {
	db := repo.mongoClient.Database("PINKKER-BACKEND")
	collPosts := db.Collection("Post")

	excludeFilter := bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: ExcludeFilterlID}}}}

	includeFilter := bson.D{
		{Key: "communityID", Value: communityID},
		{Key: "Type", Value: bson.M{"$in": []string{"Post", "RePost", "CitaPost"}}},
	}

	matchFilter := bson.D{
		{Key: "$and", Value: bson.A{
			excludeFilter,
			includeFilter,
		}},
	}

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: matchFilter}},

		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idT, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "TimeStamp", Value: -1}, // Ordenar por fecha de creación
		}}},
		bson.D{{Key: "$limit", Value: limit}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: 1},
			{Key: "PostImage", Value: 1},
			{Key: "Type", Value: 1},
			{Key: "TimeStamp", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
			{Key: "likeCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "isLikedByID", Value: 1},
		}}},
	}

	cursor, err := collPosts.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var postsWithUserInfo []communitiesdomain.PostGetCommunityReq
	for cursor.Next(ctx) {
		var postWithUserInfo communitiesdomain.PostGetCommunityReq
		if err := cursor.Decode(&postWithUserInfo); err != nil {
			return nil, err
		}
		postsWithUserInfo = append(postsWithUserInfo, postWithUserInfo)
	}

	return postsWithUserInfo, nil
}
func (repo *CommunitiesRepository) DeletePost(ctx context.Context, communityID, postID, userID primitive.ObjectID) error {
	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	// Obtener los detalles de la comunidad
	var community struct {
		CreatorID primitive.ObjectID   `bson:"CreatorID"`
		Mods      []primitive.ObjectID `bson:"Mods"` // Lista de moderadores
	}

	// Obtener los datos de la comunidad para verificar si el usuario es el creador o un moderador
	err := collection.FindOne(ctx, primitive.M{"_id": communityID}, options.FindOne().SetProjection(primitive.M{
		"CreatorID": 1,
		"Mods":      1,
	})).Decode(&community)

	if err != nil {
		return err
	}

	// Verificar si el usuario que intenta eliminar el post es el creador o un moderador
	if community.CreatorID != userID {
		isMod := false
		for _, modID := range community.Mods {
			if modID == userID {
				isMod = true
				break
			}
		}
		if !isMod {
			return fiber.NewError(fiber.StatusForbidden, "Only the creator or a mod can delete posts")
		}
	}

	// Eliminar el post de la colección de posts
	// _, err = collection.UpdateOne(
	// 	ctx,
	// 	primitive.M{"_id": communityID},
	// 	primitive.M{"$pull": primitive.M{"Posts": postID}}, // Eliminar el post de la lista de posts
	// )
	// if err != nil {
	// 	return err
	// }

	postCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	_, err = postCollection.DeleteOne(ctx, primitive.M{"_id": postID})
	if err != nil {
		return err
	}

	return nil
}

func (repo *CommunitiesRepository) DeleteCommunity(ctx context.Context, communityID, creatorID primitive.ObjectID) error {
	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	// Obtener los detalles de la comunidad
	var community struct {
		CreatorID primitive.ObjectID `bson:"CreatorID"`
	}

	err := collection.FindOne(ctx, primitive.M{"_id": communityID}, options.FindOne().SetProjection(primitive.M{
		"CreatorID": 1,
	})).Decode(&community)

	if err != nil {
		return err
	}

	// Verificar si el creador está intentando borrar la comunidad
	if community.CreatorID != creatorID {
		return fiber.NewError(fiber.StatusForbidden, "Only the creator can delete the community")
	}

	// Borrar la comunidad
	_, err = collection.DeleteOne(ctx, primitive.M{"_id": communityID})
	if err != nil {
		return err
	}

	return nil
}
func (repo *CommunitiesRepository) FindCommunityByName(ctx context.Context, communityName string) ([]communitiesdomain.CommunityDetails, error) {
	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.M{"communityName": bson.M{"$regex": communityName, "$options": "i"}}}}, // Búsqueda insensible a mayúsculas
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "CreatorID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "creatorDetails"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$creatorDetails"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "CommunityName", Value: 1},
			{Key: "Description", Value: 1},
			{Key: "IsPrivate", Value: 1},
			{Key: "CreatedAt", Value: 1},
			{Key: "UpdatedAt", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$members"}}}, // Contar miembros
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Banner", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
				{Key: "fullName", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.FullName", 0}}}},
				{Key: "email", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Email", 0}}}},
			}},
		}}},
		bson.D{{Key: "$limit", Value: 5}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var communities []communitiesdomain.CommunityDetails
	for cursor.Next(ctx) {
		var community communitiesdomain.CommunityDetails
		if err := cursor.Decode(&community); err != nil {
			return nil, err
		}
		communities = append(communities, community)
	}

	return communities, nil
}

func (repo *CommunitiesRepository) GetTop10CommunitiesByMembers(ctx context.Context) ([]communitiesdomain.CommunityDetails, error) {
	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	pipeline := bson.A{
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "CreatorID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "creatorDetails"},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "CommunityName", Value: 1},
			{Key: "Description", Value: 1},
			{Key: "IsPrivate", Value: 1},
			{Key: "CreatedAt", Value: 1},
			{Key: "UpdatedAt", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$members"}}}, // Contar miembros
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Banner", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
				{Key: "fullName", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.FullName", 0}}}},
				{Key: "email", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Email", 0}}}},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"membersCount": -1}}}, // Ordenar por membersCount descendente
		bson.D{{Key: "$limit", Value: 10}},                        // Limitar a 10
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var topCommunities []communitiesdomain.CommunityDetails
	for cursor.Next(ctx) {
		var community communitiesdomain.CommunityDetails
		if err := cursor.Decode(&community); err != nil {
			return nil, err
		}
		topCommunities = append(topCommunities, community)
	}

	return topCommunities, nil
}
func (repo *CommunitiesRepository) GetCommunity(ctx context.Context, communityID primitive.ObjectID) (*communitiesdomain.CommunityDetails, error) {
	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: communityID}}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "CreatorID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "creatorDetails"},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "CommunityName", Value: 1},
			{Key: "Description", Value: 1},
			{Key: "IsPrivate", Value: 1},
			{Key: "CreatedAt", Value: 1},
			{Key: "UpdatedAt", Value: 1},
			{Key: "Categories", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$members"}}}, // Contar miembros
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Banner", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
				{Key: "fullName", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.FullName", 0}}}},
				{Key: "email", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Email", 0}}}},
			}},
		}}},
		bson.D{{Key: "$limit", Value: 1}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var community communitiesdomain.CommunityDetails
		if err := cursor.Decode(&community); err != nil {
			return nil, err
		}
		return &community, nil
	}

	return nil, mongo.ErrNoDocuments
}

func (r *CommunitiesRepository) GetTOTPSecret(ctx context.Context, userID primitive.ObjectID) (string, error) {
	usersCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var result struct {
		TOTPSecret string `bson:"TOTPSecret"`
	}

	err := usersCollection.FindOne(
		ctx,
		bson.M{"_id": userID},
		options.FindOne().SetProjection(bson.M{"TOTPSecret": 1}),
	).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.TOTPSecret, nil

}
