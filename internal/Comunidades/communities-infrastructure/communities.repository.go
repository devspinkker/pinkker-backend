package communitiestinfrastructure

import (
	"PINKKER-BACKEND/config"
	communitiesdomain "PINKKER-BACKEND/internal/Comunidades/communities"
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
func (repo *CommunitiesRepository) CreateCommunity(ctx context.Context, req communitiesdomain.CreateCommunity, Banner string, creatorID primitive.ObjectID) (*communitiesdomain.Community, error) {
	var user struct {
		PinkkerPrime *userdomain.PinkkerPrime `bson:"PinkkerPrime"`
		Pixeles      float64                  `bson:"Pixeles"`
	}

	// CostToCreateCommunityStr := config.CostToCreateCommunity()
	// CostToCreateCommunity, err := strconv.ParseFloat(CostToCreateCommunityStr, 64)
	// if err != nil {
	// 	return nil, fmt.Errorf("error converting PinkkerPrime cost to integer: %v", err)
	// }

	usersCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	err := usersCollection.FindOne(ctx, primitive.M{"_id": creatorID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	// Verificar si el usuario tiene una suscripción activa de PinkkerPrime
	if user.PinkkerPrime == nil || user.PinkkerPrime.SubscriptionEnd.Before(time.Now()) {
		return nil, fiber.NewError(fiber.StatusForbidden, "El usuario no tiene una suscripción activa de PinkkerPrime")
	}

	// Verificar si el usuario tiene al menos 5000 pixeles
	// if user.Pixeles < CostToCreateCommunity {
	// 	return nil, fiber.NewError(fiber.StatusForbidden, "El usuario no tiene suficientes pixeles")
	// }

	// Verificar si ya existe una comunidad con el mismo nombre
	communitiesCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")
	var existingCommunity struct {
		CommunityName string `bson:"CommunityName"`
	}
	err = communitiesCollection.FindOne(ctx, primitive.M{"CommunityName": req.CommunityName}).Decode(&existingCommunity)

	if err == nil {
		return nil, fiber.NewError(fiber.StatusConflict, "Una comunidad con este nombre ya existe")
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}
	if req.IsPaid {
		req.IsPrivate = true
	}
	// Crear la nueva comunidad

	community := &communitiesdomain.Community{
		CommunityName:      req.CommunityName,
		Description:        req.Description,
		CreatorID:          creatorID,
		Members:            []primitive.ObjectID{creatorID},
		Mods:               []primitive.ObjectID{},
		BannedUsers:        []primitive.ObjectID{},
		Rules:              "Por defecto",
		IsPrivate:          req.IsPrivate,
		Categories:         req.Categories,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		IsPaid:             req.IsPaid,
		SubscriptionAmount: req.SubscriptionAmount,
		Banner:             Banner,
		AdPricePerDay:      req.AdPricePerDay,
	}
	if Banner == "" {
		community.Banner = config.BANNER()
	}
	result, err := communitiesCollection.InsertOne(ctx, community)
	if err != nil {
		return nil, err
	}

	communityID := result.InsertedID.(primitive.ObjectID)

	_, err = usersCollection.UpdateOne(
		ctx,
		primitive.M{"_id": creatorID},
		primitive.M{
			// "$inc": primitive.M{"Pixeles": -CostToCreateCommunity},
			"$addToSet": primitive.M{
				"InCommunities":    communityID,
				"OwnerCommunities": communityID,
			},
		},
	)
	// repo.updatePinkkerProfitPerMonth(ctx, CostToCreateCommunity)
	if err != nil {
		return nil, err
	}

	return community, nil
}

func (r *CommunitiesRepository) updatePinkkerProfitPerMonth(ctx context.Context, CostToCreateCommunity float64) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollMonthly := GoMongoDB.Collection("PinkkerProfitPerMonth")

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
	_, err := GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, bson.M{
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
			"total":                                CostToCreateCommunity,
			"days." + currentDay + ".communityBuy": CostToCreateCommunity,
		},
	}

	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdate)
	if err != nil {
		return err
	}

	return nil
}

func (repo *CommunitiesRepository) EditCommunity(ctx context.Context, req communitiesdomain.EditCommunity, Banner string, creatorID primitive.ObjectID) (*communitiesdomain.Community, error) {

	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	var community struct {
		CreatorID primitive.ObjectID `bson:"CreatorID"`
	}

	err := collection.FindOne(ctx, bson.M{"_id": req.CommunityID}, options.FindOne().SetProjection(bson.M{
		"CreatorID": 1,
	})).Decode(&community)
	if err != nil {
		return nil, err
	}

	if community.CreatorID != creatorID {
		return nil, errors.New("no autorizado")
	}

	update := bson.M{
		"$set": bson.M{
			"AdPricePerDay":      req.AdPricePerDay,
			"Description":        req.Description,
			"SubscriptionAmount": req.SubscriptionAmount,
			"Categories":         req.Categories,
			"IsPaid":             req.IsPaid,
			"Banner":             Banner,
			"UpdatedAt":          time.Now(),
		},
	}

	// Ejecutamos el update
	_, err = collection.UpdateOne(ctx, bson.M{"_id": req.CommunityID}, update)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar la comunidad: %v", err)
	}

	var updatedCommunity communitiesdomain.Community
	err = collection.FindOne(ctx, bson.M{"_id": req.CommunityID}).Decode(&updatedCommunity)
	if err != nil {
		return nil, fmt.Errorf("error al recuperar la comunidad actualizada: %v", err)
	}

	return &updatedCommunity, nil
}

func (repo *CommunitiesRepository) FindUserCommunities(ctx context.Context, userID primitive.ObjectID) ([]communitiesdomain.CommunityDetails, error) {
	usersCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	var user struct {
		InCommunities []primitive.ObjectID `bson:"InCommunities"`
	}

	err := usersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	if len(user.InCommunities) == 0 {
		return []communitiesdomain.CommunityDetails{}, nil
	}

	communitiesCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": user.InCommunities}}}},
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
			{Key: "AdPricePerDay", Value: 1},
			{Key: "Banner", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$Members"}}},
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.banner", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
			}},
		}}},
	}

	// Ejecutar el pipeline
	cursor, err := communitiesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Almacenar los resultados
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
func (repo *CommunitiesRepository) CommunityOwnerUser(ctx context.Context, userID primitive.ObjectID) ([]communitiesdomain.CommunityDetails, error) {
	usersCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	var user struct {
		OwnerCommunities []primitive.ObjectID `bson:"OwnerCommunities"`
	}

	err := usersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	if len(user.OwnerCommunities) == 0 {
		return []communitiesdomain.CommunityDetails{}, nil
	}

	communitiesCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": user.OwnerCommunities}}}},
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
			{Key: "SubscriptionAmount", Value: 1},
			{Key: "AdPricePerDay", Value: 1},
			{Key: "Banner", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$Members"}}},
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.banner", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
			}},
		}}},
	}

	// Ejecutar el pipeline
	cursor, err := communitiesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Almacenar los resultados
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
func (repo *CommunitiesRepository) AddModerator(ctx context.Context, communityID, newModID, modID primitive.ObjectID) error {
	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	var community struct {
		CreatorID  primitive.ObjectID   `bson:"CreatorID"`
		Moderators []primitive.ObjectID `bson:"Mods"`
	}

	err := collection.FindOne(ctx, primitive.M{"_id": communityID}, options.FindOne().SetProjection(primitive.M{
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
			"$addToSet": primitive.M{"Mods": newModID},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// handleCommunityPayment gestiona el pago para las comunidades pagas, descontando el monto del usuario y
// transfiriendo el saldo restante al creador de la comunidad, menos una comisión del 5%.
func (repo *CommunitiesRepository) handleCommunityPayment(sc mongo.SessionContext, community communitiesdomain.Community, userID primitive.ObjectID) error {
	usersCollection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	// Obtener el saldo actual del usuario
	var user struct {
		Pixeles float64 `json:"Pixeles" bson:"Pixeles"`
	} // Asumiendo que tienes una estructura User para representar al usuario
	err := usersCollection.FindOne(sc, primitive.M{"_id": userID}).Decode(&user)
	if err != nil {
		return err
	}
	if int(user.Pixeles) < community.SubscriptionAmount {
		return fiber.NewError(fiber.StatusPaymentRequired, "Insufficient Pixeles")
	}

	// Calcular el monto a pagar al creador de la comunidad (95% del valor de la suscripción)
	paymentToCreator := float64(community.SubscriptionAmount) * 0.95

	// Descontar el valor de la suscripción de los Pixeles del usuario
	_, err = usersCollection.UpdateOne(
		sc,
		primitive.M{"_id": userID},
		primitive.M{"$inc": primitive.M{"Pixeles": -community.SubscriptionAmount}},
	)
	if err != nil {
		return err
	}

	// Añadir el monto correspondiente en Pixeles al saldo del creador de la comunidad
	_, err = usersCollection.UpdateOne(
		sc,
		primitive.M{"_id": community.CreatorID},
		primitive.M{"$inc": primitive.M{"Pixeles": paymentToCreator}},
	)
	if err != nil {
		return err
	}

	return nil
}

// AddMember añade un miembro a la comunidad y gestiona el pago si la comunidad es paga.
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

		// Obtener los detalles de la comunidad
		var community communitiesdomain.Community
		err = communitiesCollection.FindOne(sc, primitive.M{"_id": communityID}).Decode(&community)
		if err != nil {
			session.AbortTransaction(sc)
			return err
		}

		// Si la comunidad es paga, gestionar el pago
		if community.IsPaid {
			err = repo.handleCommunityPayment(sc, community, userID)
			if err != nil {
				session.AbortTransaction(sc)
				return err
			}

			PaidCommunities := float64(community.SubscriptionAmount) * 0.05
			err := repo.updatePinkkerProfitPerMonthPaidCommunities(context.TODO(), PaidCommunities)
			if err != nil {
				session.AbortTransaction(sc)
				return err
			}
		} else {
			_, err = repo.GetSubscriptionByUserIDs(userID, community)
			if err != nil {
				return err
			}
		}

		// Añadir el usuario a la comunidad
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
func (r *CommunitiesRepository) updatePinkkerProfitPerMonthPaidCommunities(ctx context.Context, costo float64) error {
	GoMongoDB := r.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollMonthly := GoMongoDB.Collection("PinkkerProfitPerMonth")

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
	_, err := GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, bson.M{
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
			"total": costo,
			"days." + currentDay + ".PaidCommunities": costo,
		},
	}

	_, err = GoMongoDBCollMonthly.UpdateOne(ctx, monthlyFilter, monthlyUpdate)
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
		bson.D{{Key: "$match", Value: bson.M{"CommunityName": bson.M{"$regex": communityName, "$options": "i"}}}}, // Búsqueda insensible a mayúsculas
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
			{Key: "AdPricePerDay", Value: 1},
			{Key: "Categories", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$Members"}}},
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Banner", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"membersCount": -1}}},
		bson.D{{Key: "$limit", Value: 10}},
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
			{Key: "Categories", Value: 1},
			{Key: "AdPricePerDay", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$Members"}}},
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Banner", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"membersCount": -1}}},
		bson.D{{Key: "$limit", Value: 10}},
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
func (repo *CommunitiesRepository) GetTop10CommunitiesByMembersNoMember(ctx context.Context, userID primitive.ObjectID) ([]communitiesdomain.CommunityDetails, error) {
	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	pipeline := bson.A{
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "CreatorID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "creatorDetails"},
		}}},
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "Members", Value: bson.D{{Key: "$ne", Value: userID}}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "CommunityName", Value: 1},
			{Key: "Description", Value: 1},
			{Key: "Banner", Value: 1},
			{Key: "IsPrivate", Value: 1},
			{Key: "CreatedAt", Value: 1},
			{Key: "UpdatedAt", Value: 1},
			{Key: "Categories", Value: 1},
			{Key: "AdPricePerDay", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$Members"}}},
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Banner", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
			}},
		}}},
		// Ordenamos por el tamaño de los miembros en orden descendente
		bson.D{{Key: "$sort", Value: bson.M{"membersCount": -1}}},
		// Limitamos a las 10 comunidades más grandes
		bson.D{{Key: "$limit", Value: 10}},
	}

	// Ejecutamos el pipeline
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decodificamos los resultados
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

func (repo *CommunitiesRepository) GetCommunityWithUserMembership(ctx context.Context, communityID, userID primitive.ObjectID) (*communitiesdomain.CommunityDetails, error) {
	collection := repo.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	pipeline := bson.A{
		// Match the community by its ID
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: communityID}}}},

		// Lookup creator details from the "Users" collection
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "CreatorID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "creatorDetails"},
		}}},

		// Add a projection to select specific fields
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "CommunityName", Value: 1},
			{Key: "Description", Value: 1},
			{Key: "IsPrivate", Value: 1},
			{Key: "Banner", Value: 1},
			{Key: "CreatedAt", Value: 1},
			{Key: "UpdatedAt", Value: 1},
			{Key: "SubscriptionAmount", Value: 1},
			{Key: "Categories", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$Members"}}},
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Banner", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
			}},
			{Key: "isUserMember", Value: bson.D{
				{Key: "$in", Value: bson.A{userID, "$Members"}},
			}},
			{Key: "isUserModerator", Value: bson.D{
				{Key: "$in", Value: bson.A{userID, "$Mods"}},
			}},
		}}},

		// Sort by the number of members
		bson.D{{Key: "$sort", Value: bson.M{"membersCount": -1}}},

		// Limit to 1 result
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

func (t *CommunitiesRepository) addOriginalPostDataCommunity(ctx context.Context, collTweets *mongo.Collection, Posts []communitiesdomain.PostGetCommunityReq) error {
	var originalPostIDs []primitive.ObjectID
	for _, tweet := range Posts {
		if tweet.OriginalPost != primitive.NilObjectID {
			originalPostIDs = append(originalPostIDs, tweet.OriginalPost)
		}
	}
	if len(originalPostIDs) > 0 {
		originalPostPipeline := bson.A{
			bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: originalPostIDs}}}}}},
			bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "Users"},
				{Key: "localField", Value: "UserID"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "UserInfo"},
			}}},
			bson.D{{Key: "$unwind", Value: "$UserInfo"}},
			bson.D{{Key: "$project", Value: bson.D{
				{Key: "id", Value: "$_id"},
				{Key: "Type", Value: "$Type"},
				{Key: "Status", Value: "$Status"},
				{Key: "PostImage", Value: "$PostImage"},
				{Key: "TimeStamp", Value: "$TimeStamp"},
				{Key: "UserID", Value: "$UserID"},
				{Key: "Likes", Value: "$Likes"},
				{Key: "Comments", Value: "$Comments"},
				{Key: "RePosts", Value: "$RePosts"},
				{Key: "Views", Value: "$Views"},
				{Key: "OriginalPost", Value: "$OriginalPost"},
				{Key: "UserInfo.FullName", Value: 1},
				{Key: "UserInfo.Avatar", Value: 1},
				{Key: "UserInfo.NameUser", Value: 1},
			}}},
		}

		cursorOriginalPosts, err := collTweets.Aggregate(ctx, originalPostPipeline)
		if err != nil {
			return err
		}
		defer cursorOriginalPosts.Close(ctx)

		var originalPostMap = make(map[primitive.ObjectID]communitiesdomain.PostGetCommunityReq)
		for cursorOriginalPosts.Next(ctx) {
			var originalPost communitiesdomain.PostGetCommunityReq
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return err
			}
			originalPostMap[originalPost.ID] = originalPost
		}

		for i, tweet := range Posts {
			if tweet.OriginalPost != primitive.NilObjectID {
				originalPost, found := originalPostMap[tweet.OriginalPost]
				if found {
					Posts[i].OriginalPostData = &originalPost
				}
			}
		}
	}
	return nil
}

// solo lo pase abajo
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
			{Key: "AdPricePerDay", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$Members"}}},
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Banner", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"membersCount": -1}}},
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

func (r *CommunitiesRepository) GetSubscriptionByUserIDs(sourceUserID primitive.ObjectID, community communitiesdomain.Community) (*subscriptiondomain.Subscription, error) {

	if community.CreatorID == sourceUserID {
		return nil, nil
	}

	if !community.IsPrivate {
		return nil, nil
	}

	subscriptionsCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscriptions")

	filter := bson.M{
		"sourceUserID":      sourceUserID,
		"destinationUserID": community.CreatorID,
	}

	var subscription subscriptiondomain.Subscription
	err := subscriptionsCollection.FindOne(context.Background(), filter).Decode(&subscription)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no se encontró ninguna suscripción entre el usuario y el creador")
		}
		return nil, fmt.Errorf("error al buscar la suscripción: %v", err)
	}

	if subscription.SubscriptionEnd.Before(time.Now()) {
		return nil, fmt.Errorf("la suscripción ha expirado")
	}

	return &subscription, nil
}
func (r *CommunitiesRepository) GetSubscriptionByUserIDsAndcommunityID(sourceUserID, communityID primitive.ObjectID) (*subscriptiondomain.Subscription, error) {
	communitiesCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	redisKey := fmt.Sprintf("community:%s", communityID.Hex())
	redisData, err := r.redisClient.Get(context.Background(), redisKey).Result()

	var community struct {
		CreatorID primitive.ObjectID `bson:"CreatorID"`
		IsPrivate bool               `bson:"IsPrivate"`
	}

	if err == redis.Nil {
		err := communitiesCollection.FindOne(context.Background(), bson.M{"_id": communityID}).Decode(&community)
		if err != nil {
			return nil, fmt.Errorf("error al obtener el creador de la comunidad: %v", err)
		}

		communityJSON, err := json.Marshal(community)
		if err == nil {
			r.redisClient.Set(context.Background(), redisKey, communityJSON, time.Minute*1)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error al obtener datos de Redis: %v", err)
	} else {
		// fmt.Println("Encontrado en Redis")
		err = json.Unmarshal([]byte(redisData), &community)
		if err != nil {
			return nil, fmt.Errorf("error al deserializar datos de Redis: %v", err)
		}
	}

	if community.CreatorID == sourceUserID {
		return nil, nil
	}

	if !community.IsPrivate {
		return nil, nil
	}

	subscriptionsCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Subscriptions")

	filter := bson.M{
		"sourceUserID":      sourceUserID,
		"destinationUserID": community.CreatorID,
	}

	var subscription subscriptiondomain.Subscription
	err = subscriptionsCollection.FindOne(context.Background(), filter).Decode(&subscription)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no se encontró ninguna suscripción entre el usuario y el creador")
		}
		return nil, fmt.Errorf("error al buscar la suscripción: %v", err)
	}

	if subscription.SubscriptionEnd.Before(time.Now()) {
		return nil, fmt.Errorf("la suscripción ha expirado")
	}

	return &subscription, nil
}
func (repo *CommunitiesRepository) GetCommunityRecommended(ctx context.Context, userID primitive.ObjectID, page int, limit int) ([]communitiesdomain.CommunityDetails, error) {
	Database := repo.mongoClient.Database("PINKKER-BACKEND")
	collection := Database.Collection("communities")
	userCollection := Database.Collection("Users")

	// Obtener los IDs de los usuarios que sigue el usuario actual
	userPipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.M{"_id": userID}}},
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

	cursor, err := userCollection.Aggregate(ctx, userPipeline)
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
	}

	followingIDs := userResult.FollowingIDs
	if followingIDs == nil {
		followingIDs = []primitive.ObjectID{}
	}

	skip := (page - 1) * limit

	pipeline := bson.A{
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "CreatorID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "creatorDetails"},
		}}},
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "Members", Value: bson.D{{Key: "$ne", Value: userID}}},
			{Key: "BannedUsers", Value: bson.D{{Key: "$ne", Value: userID}}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "relevanceScore", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{{Key: "$in", Value: bson.A{"$CreatorID", followingIDs}}}},
					{Key: "then", Value: 10},
					{Key: "else", Value: 0},
				}},
			}},
		}}},

		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "CommunityName", Value: 1},
			{Key: "Description", Value: 1},
			{Key: "Banner", Value: 1},
			{Key: "IsPrivate", Value: 1},
			{Key: "CreatedAt", Value: 1},
			{Key: "UpdatedAt", Value: 1},
			{Key: "Categories", Value: 1},
			{Key: "AdPricePerDay", Value: 1},
			{Key: "membersCount", Value: bson.D{{Key: "$size", Value: "$Members"}}},
			{Key: "relevanceScore", Value: 1},
			{Key: "creator", Value: bson.D{
				{Key: "userID", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails._id", 0}}}},
				{Key: "avatar", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Avatar", 0}}}},
				{Key: "banner", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.Banner", 0}}}},
				{Key: "nameUser", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$creatorDetails.NameUser", 0}}}},
			}},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"relevanceScore": -1, "membersCount": -1}}},
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: limit}},
	}

	// Ejecutamos el pipeline
	cursor, err = collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decodificamos los resultados
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
