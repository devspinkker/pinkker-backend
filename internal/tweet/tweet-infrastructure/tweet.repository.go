package tweetinfrastructure

import (
	communitiesdomain "PINKKER-BACKEND/internal/Comunidades/communities"
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	notificationsdomain "PINKKER-BACKEND/internal/notifications/notifications"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	tweetdomain "PINKKER-BACKEND/internal/tweet/tweet-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TweetRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewTweetRepository(redisClient *redis.Client, mongoClient *mongo.Client) *TweetRepository {
	return &TweetRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}

func (t *TweetRepository) GetTweetsRecommended(idT primitive.ObjectID, excludeIDs []primitive.ObjectID, limit int) ([]tweetdomain.TweetGetFollowReq, error) {
	ctx := context.Background()
	db := t.mongoClient.Database("PINKKER-BACKEND")
	collTweets := db.Collection("Post")

	// Extraer los ObjectID de los usuarios que sigue el usuario
	last24Hours := time.Now().Add(-24 * time.Hour)
	excludedIDs := t.getExcludedIDs(excludeIDs)
	excludeFilter := bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: excludedIDs}}}}

	// Ejecutar pipeline principal para obtener tweets relevantes
	tweetsWithUserInfo, err := t.getRelevantTweets(ctx, idT, collTweets, excludeFilter, last24Hours, limit)
	if err != nil {
		return nil, err
	}
	// Calcular el nuevo límite para el pipeline secundario
	newLimit := limit - len(tweetsWithUserInfo)
	if newLimit > 0 {
		var recommendedPostsIDs []interface{}
		for _, clip := range tweetsWithUserInfo {
			recommendedPostsIDs = append(recommendedPostsIDs, clip.ID)
		}

		excludeFilter := bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "$nin", Value: append(excludedIDs, recommendedPostsIDs...)},
			}},
		}
		randomTweets, err := t.getRandomTweets(ctx, idT, collTweets, excludeFilter, newLimit)
		if err != nil {
			return nil, err
		}
		tweetsWithUserInfo = append(tweetsWithUserInfo, randomTweets...)
	}

	// Actualizar el campo Views sumando 1 para los posts obtenidos en ambos pipelines
	if err := t.updateTweetViews(ctx, collTweets, tweetsWithUserInfo, idT); err != nil {
		return nil, err
	}

	// Obtener datos de los posts originales
	if err := t.addOriginalPostData(ctx, collTweets, tweetsWithUserInfo); err != nil {
		return nil, err
	}

	return tweetsWithUserInfo, nil
}

func (t *TweetRepository) getUser(ctx context.Context, collUsers *mongo.Collection, idT primitive.ObjectID) (userdomain.GetUser, error) {
	var user userdomain.GetUser
	err := collUsers.FindOne(ctx, bson.D{{Key: "_id", Value: idT}}).Decode(&user)
	return user, err
}

func (t *TweetRepository) getExcludedIDs(excludeIDs []primitive.ObjectID) []interface{} {
	excludedIDs := make([]interface{}, len(excludeIDs))
	for i, id := range excludeIDs {
		excludedIDs[i] = id
	}
	return excludedIDs
}
func (t *TweetRepository) getRandomTweets(ctx context.Context, idT primitive.ObjectID, collTweets *mongo.Collection, excludeFilter bson.D, limit int) ([]tweetdomain.TweetGetFollowReq, error) {
	pipelineRandom := bson.A{

		bson.D{{Key: "$match", Value: bson.D{

			{Key: "$or", Value: bson.A{
				bson.D{{Key: "communityID", Value: primitive.NilObjectID}},
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "IsPrivate", Value: false}},
				bson.D{{Key: "IsPrivate", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
			{Key: "Type", Value: bson.M{"$in": []string{"Post", "RePost", "CitaPost"}}}},
		}},
		bson.D{{Key: "$match", Value: excludeFilter}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idT, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "relevanceScore", Value: bson.D{{Key: "$add", Value: bson.A{
				// Ponderar más fuertemente los likes
				bson.D{{Key: "$multiply", Value: bson.A{"$likeCount", 5}}},
				// Frescura del post
				bson.D{{Key: "$subtract", Value: bson.A{1000, bson.D{{Key: "$divide", Value: bson.A{bson.D{{Key: "$subtract", Value: bson.A{time.Now(), "$TimeStamp"}}}, 3600000}}}}}},
			}}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "relevanceScore", Value: -1},
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

			{Key: "OriginalPost", Value: 1},
			{Key: "Views", Value: 1},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},

			{Key: "CommentsCount", Value: 1},
			{Key: "isLikedByID", Value: 1},
		}}},
	}

	cursor, err := collTweets.Aggregate(ctx, pipelineRandom)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tweetsWithUserInfo []tweetdomain.TweetGetFollowReq
	for cursor.Next(ctx) {
		var tweetWithUserInfo tweetdomain.TweetGetFollowReq
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}
		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	return tweetsWithUserInfo, nil
}
func (t *TweetRepository) getRelevantTweets(ctx context.Context, idT primitive.ObjectID, collTweets *mongo.Collection, excludeFilter bson.D, last24Hours time.Time, limit int) ([]tweetdomain.TweetGetFollowReq, error) {
	userCollection := collTweets.Database().Collection("Users")
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

	tweetPipeline := bson.A{
		bson.D{{Key: "$match", Value: excludeFilter}},

		bson.D{{Key: "$match", Value: bson.D{

			{Key: "$or", Value: bson.A{
				bson.D{{Key: "communityID", Value: primitive.NilObjectID}},
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "IsPrivate", Value: false}},
				bson.D{{Key: "IsPrivate", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
			{Key: "Type", Value: bson.M{"$in": []string{"Post", "RePost", "CitaPost"}}},
			{Key: "TimeStamp", Value: bson.M{"$gte": last24Hours}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "isFollowingUser", Value: bson.D{
				{Key: "$in", Value: bson.A{"$UserID", bson.D{
					{Key: "$ifNull", Value: bson.A{followingIDs, bson.A{}}},
				}}},
			}},
			{Key: "likedByFollowing", Value: bson.D{{Key: "$setIntersection", Value: bson.A{"$Likes", followingIDs}}}},
			{Key: "repostedByFollowing", Value: bson.D{{Key: "$setIntersection", Value: bson.A{"$RePosts", followingIDs}}}},
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idT, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "relevanceScore", Value: bson.D{{Key: "$add", Value: bson.A{
				// Ponderar más fuertemente los posts de los usuarios seguidos
				bson.D{{Key: "$multiply", Value: bson.A{bson.D{{Key: "$cond", Value: bson.D{
					{Key: "if", Value: "$isFollowingUser"},
					{Key: "then", Value: 5}, // Mayor ponderación para los posts de usuarios seguidos
					{Key: "else", Value: 0},
				}}}, 3}}},
				// Ponderar más fuertemente los "me gusta" de los usuarios seguidos
				bson.D{{Key: "$multiply", Value: bson.A{
					bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$likedByFollowing", bson.A{}}}}}},
					5,
				}}},
				// Ponderar los reposts de los usuarios seguidos
				bson.D{{Key: "$multiply", Value: bson.A{
					bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$repostedByFollowing", bson.A{}}}}}},
					2,
				}}},
				// Frescura del post
				bson.D{{Key: "$subtract", Value: bson.A{1000, bson.D{{Key: "$divide", Value: bson.A{bson.D{{Key: "$subtract", Value: bson.A{time.Now(), "$TimeStamp"}}}, 3600000}}}}}},
			}}}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "relevanceScore", Value: -1},
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

			{Key: "OriginalPost", Value: 1},
			{Key: "Views", Value: 1},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
			{Key: "likeCount", Value: 1},

			{Key: "RePostsCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "isLikedByID", Value: 1},
		}}},
	}

	cursor, err = collTweets.Aggregate(ctx, tweetPipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tweetsWithUserInfo []tweetdomain.TweetGetFollowReq
	for cursor.Next(ctx) {
		var tweetWithUserInfo tweetdomain.TweetGetFollowReq
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}
		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	return tweetsWithUserInfo, nil
}

func (t *TweetRepository) GetRandomPostcommunities(idT primitive.ObjectID, excludeIDs []primitive.ObjectID, limit int) ([]tweetdomain.GetPostcommunitiesRandom, error) {

	ctx := context.Background()
	db := t.mongoClient.Database("PINKKER-BACKEND")
	collTweets := db.Collection("Post")

	// last24Hours := time.Now().Add(-24 * time.Hour)

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$nin", Value: excludeIDs}}},
			// {Key: "TimeStamp", Value: bson.D{{Key: "$gte", Value: last24Hours}}},
			{Key: "Type", Value: bson.M{"$in": []string{"Post", "RePost", "CitaPost"}}},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "IsPrivate", Value: false}},
				bson.D{{Key: "IsPrivate", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
			{Key: "$and", Value: bson.A{
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$ne", Value: []primitive.ObjectID{idT}}}}},
			}},
		}}},

		// Unimos la información de la comunidad
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "communities"},
			{Key: "localField", Value: "communityID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "CommunityInfo"},
		}}},

		// Unimos la información del usuario
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},

		// Descomponemos la información del usuario y de la comunidad
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$unwind", Value: "$CommunityInfo"}},

		// Añadimos campos adicionales (likes, reposts, comentarios, etc.)
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idT, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},

		// Ordenamos por relevancia (basado en el conteo de likes y la antigüedad del tweet)
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "relevanceScore", Value: bson.D{{Key: "$add", Value: bson.A{
				bson.D{{Key: "$multiply", Value: bson.A{"$likeCount", 5}}},
				bson.D{{Key: "$subtract", Value: bson.A{1000, bson.D{{Key: "$divide", Value: bson.A{bson.D{{Key: "$subtract", Value: bson.A{time.Now(), "$TimeStamp"}}}, 3600000}}}}}}}}}}},
		}},

		// Ordenamos por relevancia
		bson.D{{Key: "$sort", Value: bson.D{{Key: "relevanceScore", Value: -1}}}},

		// Limitamos los resultados
		bson.D{{Key: "$limit", Value: limit}},

		// Proyectamos los campos necesarios
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: 1},
			{Key: "PostImage", Value: 1},
			{Key: "Type", Value: 1},
			{Key: "TimeStamp", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "OriginalPost", Value: 1},
			{Key: "Views", Value: 1},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
			{Key: "CommunityInfo.CommunityName", Value: 1},
			{Key: "CommunityInfo._id", Value: 1},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "isLikedByID", Value: 1},
		}}},
	}

	cursor, err := collTweets.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tweetsWithUserInfo []tweetdomain.GetPostcommunitiesRandom
	for cursor.Next(ctx) {
		var tweetWithUserInfo tweetdomain.GetPostcommunitiesRandom
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}
		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}
	if err := t.addOriginalPostDataCommunity(ctx, collTweets, tweetsWithUserInfo); err != nil {
		return nil, err
	}
	return tweetsWithUserInfo, nil
}

func (t *TweetRepository) GetPostsFromUserCommunities(idT primitive.ObjectID, excludeIDs []primitive.ObjectID, limit int) ([]tweetdomain.GetPostcommunitiesRandom, error) {
	ctx := context.Background()
	db := t.mongoClient.Database("PINKKER-BACKEND")
	collTweets := db.Collection("Post")
	collUsers := db.Collection("Users")

	var user struct {
		InCommunities []primitive.ObjectID `bson:"InCommunities"`
	}

	err := collUsers.FindOne(ctx, bson.M{"_id": idT}).Decode(&user)
	if err != nil {
		return nil, err
	}

	if user.InCommunities == nil || len(user.InCommunities) == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "The user is not part of any communities")
	}

	pipeline := bson.A{
		// Filtramos por posts de las comunidades en las que el usuario está
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$nin", Value: excludeIDs}}},
			// {Key: "TimeStamp", Value: bson.D{{Key: "$gte", Value: last24Hours}}},
			{Key: "Type", Value: bson.M{"$in": []string{"Post", "RePost", "CitaPost"}}},
			// {Key: "$or", Value: bson.A{
			// 	bson.D{{Key: "IsPrivate", Value: false}},
			// 	bson.D{{Key: "IsPrivate", Value: bson.D{{Key: "$exists", Value: false}}}},
			// }},
			{Key: "$and", Value: bson.A{
				// Filtramos los posts de las comunidades en las que el usuario está
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$in", Value: user.InCommunities}}}},
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$ne", Value: primitive.NilObjectID}}}}}}}},
		},

		// Unimos la información de la comunidad
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "communities"},
			{Key: "localField", Value: "communityID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "CommunityInfo"},
		}}},

		// Unimos la información del usuario
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},

		// Descomponemos la información del usuario y de la comunidad
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		bson.D{{Key: "$unwind", Value: "$CommunityInfo"}},

		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idT, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},

		// Ordenamos por relevancia (basado en el conteo de likes y la antigüedad del tweet)
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "relevanceScore", Value: bson.D{{Key: "$add", Value: bson.A{
				bson.D{{Key: "$multiply", Value: bson.A{"$likeCount", 5}}},
				bson.D{{Key: "$subtract", Value: bson.A{1000, bson.D{{Key: "$divide", Value: bson.A{bson.D{{Key: "$subtract", Value: bson.A{time.Now(), "$TimeStamp"}}}, 3600000}}}}}}}}}}}},
		},

		// Ordenamos por relevancia
		bson.D{{Key: "$sort", Value: bson.D{{Key: "relevanceScore", Value: -1}}}},

		// Limitamos los resultados
		bson.D{{Key: "$limit", Value: limit}},

		// Proyectamos los campos necesarios
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: 1},
			{Key: "PostImage", Value: 1},
			{Key: "Type", Value: 1},
			{Key: "TimeStamp", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "OriginalPost", Value: 1},
			{Key: "Views", Value: 1},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
			{Key: "CommunityInfo.CommunityName", Value: 1},
			{Key: "CommunityInfo._id", Value: 1},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "isLikedByID", Value: 1},
		}}},
	}

	cursor, err := collTweets.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tweetsWithUserInfo []tweetdomain.GetPostcommunitiesRandom
	for cursor.Next(ctx) {
		var tweetWithUserInfo tweetdomain.GetPostcommunitiesRandom
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}
		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	if err := t.addOriginalPostDataCommunity(ctx, collTweets, tweetsWithUserInfo); err != nil {
		return nil, err
	}
	if err := t.PostcommunitiesRandomAddView(context.TODO(), collTweets, tweetsWithUserInfo, idT); err != nil {
		return nil, err
	}
	return tweetsWithUserInfo, nil
}
func (t *TweetRepository) addOriginalPostDataCommunity(ctx context.Context, collTweets *mongo.Collection, tweets []tweetdomain.GetPostcommunitiesRandom) error {
	var originalPostIDs []primitive.ObjectID
	for _, tweet := range tweets {
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

		var originalPostMap = make(map[primitive.ObjectID]tweetdomain.TweetGetFollowReq)
		for cursorOriginalPosts.Next(ctx) {
			var originalPost tweetdomain.TweetGetFollowReq
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return err
			}
			originalPostMap[originalPost.ID] = originalPost
		}

		for i, tweet := range tweets {
			if tweet.OriginalPost != primitive.NilObjectID {
				originalPost, found := originalPostMap[tweet.OriginalPost]
				if found {
					tweets[i].OriginalPostData = &originalPost
				}
			}
		}
	}
	return nil
}

func (t *TweetRepository) addOriginalPostData(ctx context.Context, collTweets *mongo.Collection, tweets []tweetdomain.TweetGetFollowReq) error {
	var originalPostIDs []primitive.ObjectID
	for _, tweet := range tweets {
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

		var originalPostMap = make(map[primitive.ObjectID]tweetdomain.TweetGetFollowReq)
		for cursorOriginalPosts.Next(ctx) {
			var originalPost tweetdomain.TweetGetFollowReq
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return err
			}
			originalPostMap[originalPost.ID] = originalPost
		}

		for i, tweet := range tweets {
			if tweet.OriginalPost != primitive.NilObjectID {
				originalPost, found := originalPostMap[tweet.OriginalPost]
				if found {
					tweets[i].OriginalPostData = &originalPost
				}
			}
		}
	}
	return nil
}
func (t *TweetRepository) addOriginalPostDataGetCommunityPosts(ctx context.Context, collTweets *mongo.Collection, tweets []communitiesdomain.PostGetCommunityReq) error {
	var originalPostIDs []primitive.ObjectID
	for _, tweet := range tweets {
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

		for i, tweet := range tweets {
			if tweet.OriginalPost != primitive.NilObjectID {
				originalPost, found := originalPostMap[tweet.OriginalPost]
				if found {
					tweets[i].OriginalPostData = &originalPost
				}
			}
		}
	}
	return nil
}
func (t *TweetRepository) GetAdsMuroAndPost() (tweetdomain.PostAds, error) {
	Advertisements, err := t.GetAdsMuro()
	if err != nil {
		return tweetdomain.PostAds{}, err
	}
	post, err := t.GetPostId(Advertisements.DocumentToBeAnnounced)
	if err != nil {
		return tweetdomain.PostAds{}, err
	}
	var PostAds tweetdomain.PostAds

	PostAds.AdvertisementsId = Advertisements.ID
	PostAds.ReferenceLink = Advertisements.ReferenceLink

	PostAds.ID = post.ID
	PostAds.Status = post.Status
	PostAds.TimeStamp = post.TimeStamp
	PostAds.Type = post.Type
	PostAds.Hashtags = post.Hashtags
	PostAds.UserInfo = post.UserInfo
	PostAds.OriginalPostData = post.OriginalPostData
	PostAds.Views = post.Views
	PostAds.IsLikedByID = post.IsLikedByID
	PostAds.LikeCount = post.LikeCount
	PostAds.RePostsCount = post.RePostsCount
	PostAds.CommentsCount = post.CommentsCount
	PostAds.PostImage = post.PostImage

	return PostAds, err

}
func (t *TweetRepository) GetAdsMuro() (advertisements.Advertisements, error) {
	db := t.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollAdvertisements := db.Collection("Advertisements")
	ctx := context.TODO()

	pipelineRandom := bson.A{
		bson.M{"$match": bson.M{
			"Destination": "Muro",
			"State":       "accepted",
			"$expr": bson.M{
				"$lte": bson.A{"$Clicks", "$ClicksMax"},
			},
		}},
		bson.M{"$sample": bson.M{"size": 1}},
		bson.M{"$project": bson.M{
			"IdOfTheUsersWhoClicked": 0,
			"ClicksPerDay":           0,
		}},
	}

	var advertisement advertisements.Advertisements

	cursor, err := GoMongoDBCollAdvertisements.Aggregate(ctx, pipelineRandom)
	if err != nil {
		return advertisements.Advertisements{}, err
	}
	defer cursor.Close(ctx)

	// Decodificar el resultado si se encuentra
	if cursor.Next(ctx) {
		if err := cursor.Decode(&advertisement); err != nil {
			return advertisements.Advertisements{}, err
		}
		return advertisement, nil
	}

	return advertisements.Advertisements{}, errors.New("no advertisements found")
}

func (t *TweetRepository) GetAdsMuroByCommunityId(communityId primitive.ObjectID) (tweetdomain.PostAds, error) {
	Advertisements, err := t.GetAdsByCommunityId(communityId)
	if err != nil {
		fmt.Println(err)
		return tweetdomain.PostAds{}, err
	}
	post, err := t.GetPostForAd(Advertisements.DocumentToBeAnnounced)
	if err != nil {
		return tweetdomain.PostAds{}, err
	}
	var PostAds tweetdomain.PostAds
	PostAds.AdvertisementsId = Advertisements.ID
	PostAds.ReferenceLink = Advertisements.ReferenceLink

	PostAds.ID = post.ID
	PostAds.Status = post.Status
	PostAds.TimeStamp = post.TimeStamp
	PostAds.Type = post.Type
	PostAds.Hashtags = post.Hashtags
	PostAds.UserInfo = post.UserInfo
	PostAds.OriginalPostData = post.OriginalPostData
	PostAds.Views = post.Views
	PostAds.IsLikedByID = post.IsLikedByID
	PostAds.LikeCount = post.LikeCount
	PostAds.RePostsCount = post.RePostsCount
	PostAds.CommentsCount = post.CommentsCount
	PostAds.PostImage = post.PostImage

	return PostAds, err

}
func (t *TweetRepository) GetAdsByCommunityId(communityId primitive.ObjectID) (advertisements.Advertisements, error) {
	db := t.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollAdvertisements := db.Collection("Advertisements")
	ctx := context.TODO()

	// Obtiene la fecha actual en formato UTC
	currentTime := time.Now().UTC()

	pipeline := bson.A{
		bson.M{"$match": bson.M{
			"CommunityId": communityId,
			"Destination": "Comunidades",
			"State":       "accepted",
			"$expr": bson.M{
				"$gt": bson.A{"$EndAdCommunity", currentTime},
			},
		}},
		bson.M{"$sample": bson.M{"size": 1}}, // Para tomar un anuncio aleatorio
		bson.M{"$project": bson.M{
			"IdOfTheUsersWhoClicked": 0,
			"ClicksPerDay":           0,
		}},
	}

	var advertisement advertisements.Advertisements

	cursor, err := GoMongoDBCollAdvertisements.Aggregate(ctx, pipeline)
	if err != nil {
		return advertisements.Advertisements{}, err
	}
	defer cursor.Close(ctx)

	// Decodificar el resultado si se encuentra
	if cursor.Next(ctx) {
		if err := cursor.Decode(&advertisement); err != nil {
			return advertisements.Advertisements{}, err
		}
		return advertisement, nil
	}

	return advertisements.Advertisements{}, errors.New("no advertisements found")
}

func (t *TweetRepository) IsUserBanned(userId primitive.ObjectID) (bool, error) {
	// Conexión a la colección de Usuarios
	GoMongoDBCollUsers := t.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	// Buscar al usuario por su ID
	var user struct {
		Banned bool `bson:"Banned"`
	}
	err := GoMongoDBCollUsers.FindOne(context.Background(), bson.D{{Key: "_id", Value: userId}}).Decode(&user)
	if err != nil {
		return false, fmt.Errorf("no se pudo encontrar al usuario: %v", err)
	}

	// Devolver el estado de baneado
	return user.Banned, nil
}

// Save
func (t *TweetRepository) TweetSave(Tweet tweetdomain.Post) (primitive.ObjectID, error) {
	banned, err := t.IsUserBanned(Tweet.UserID)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	if banned {
		return primitive.ObjectID{}, errors.New("without permission")

	}
	GoMongoDBCollUsers := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	result, errInsertOne := GoMongoDBCollUsers.InsertOne(context.Background(), Tweet)
	if errInsertOne != nil {
		return primitive.ObjectID{}, errInsertOne
	}
	insertedID := result.InsertedID.(primitive.ObjectID)
	return insertedID, nil
}
func (t *TweetRepository) UpdateTrends(hashtags []string) error {
	GoMongoDB := t.mongoClient.Database("PINKKER-BACKEND").Collection("Trends")
	for _, hashtag := range hashtags {
		filter := bson.M{"Hashtags": hashtag}
		// Comprueba si el hashtag ya existe en la base de datos
		var existingTrend tweetdomain.Trend
		err := GoMongoDB.FindOne(context.Background(), filter).Decode(&existingTrend)
		if err == mongo.ErrNoDocuments {
			// Si el hashtag no existe, lo inserta en la base de datos
			trend := tweetdomain.Trend{
				Hashtag:  hashtag,
				Count:    1,
				LastSeen: time.Now(),
			}
			_, err := GoMongoDB.InsertOne(context.Background(), trend)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			update := bson.M{"$inc": bson.M{"count": 1}, "$set": bson.M{"last_seen": time.Now()}}
			opts := options.Update().SetUpsert(true)
			_, err := GoMongoDB.UpdateOne(context.Background(), filter, update, opts)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (t *TweetRepository) SaveComment(tweetComment *tweetdomain.PostComment) (primitive.ObjectID, error) {
	banned, err := t.IsUserBanned(tweetComment.UserID)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	if banned {
		return primitive.ObjectID{}, errors.New("without permission")

	}
	GoMongoDBCollComments := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	res, errInsertOne := GoMongoDBCollComments.InsertOne(context.Background(), tweetComment)
	if errInsertOne != nil {
		return primitive.ObjectID{}, errInsertOne
	}

	insertedID := res.InsertedID.(primitive.ObjectID)
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	filter := bson.M{"_id": tweetComment.OriginalPost}
	update := bson.M{"$push": bson.M{"Comments": insertedID}}

	_, err = GoMongoDBCollTweets.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return insertedID, err
	}

	return insertedID, nil
}
func (t *TweetRepository) FindTweetbyId(idTweet, idT primitive.ObjectID) (tweetdomain.TweetGetFollowReq, error) {
	GoMongoDBCollUsers := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.M{"_id": idTweet}}},

		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Likes"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "PostID"},
			{Key: "as", Value: "likesInfo"},
		}}},
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: "$likesInfo"}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idT, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "Status", Value: 1},
			{Key: "PostImage", Value: 1},
			{Key: "TimeStamp", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Comments", Value: 1},

			{Key: "OriginalPost", Value: 1},
			{Key: "Type", Value: 1},
			{Key: "Hashtags", Value: 1},
			{Key: "UserInfo", Value: bson.D{
				{Key: "FullName", Value: 1},
				{Key: "Avatar", Value: 1},
				{Key: "NameUser", Value: 1},
				{Key: "Online", Value: 1},
			}},
			{Key: "OriginalPostData", Value: 1},
			{Key: "Views", Value: 1},
			{Key: "isLikedByID", Value: 1},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
		}}},
	}

	cursor, err := GoMongoDBCollUsers.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return tweetdomain.TweetGetFollowReq{}, err
	}
	defer cursor.Close(context.TODO())

	// Obtener el resultado
	var PostDocument tweetdomain.TweetGetFollowReq
	if cursor.Next(context.TODO()) {
		if err := cursor.Decode(&PostDocument); err != nil {
			return tweetdomain.TweetGetFollowReq{}, err
		}
	}
	var posts []tweetdomain.TweetGetFollowReq
	posts = append(posts, PostDocument)

	if err := t.updateTweetViews(context.TODO(), GoMongoDBCollUsers, posts, idT); err != nil {
		return tweetdomain.TweetGetFollowReq{}, err
	}

	return PostDocument, nil
}

func (t *TweetRepository) UpdateTweetbyId(tweet tweetdomain.Post) error {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	filter := bson.D{{Key: "_id", Value: tweet.ID}}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Likes", Value: tweet.Likes},
		}},
	}

	_, err := GoMongoDBCollTweets.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// Like
func (t *TweetRepository) LikeTweet(TweetId, idValueToken primitive.ObjectID) (primitive.ObjectID, string, error) {
	GoMongoDBColl := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	// Obtener el UserID (propietario del Tweet)
	var tweet struct {
		UserID primitive.ObjectID `bson:"UserID"`
	}

	err := GoMongoDBColl.FindOne(context.Background(), bson.D{{Key: "_id", Value: TweetId}}).Decode(&tweet)
	if err != nil {
		return primitive.NilObjectID, "", fmt.Errorf("error obteniendo el UserID: %v", err)
	}

	// Obtener el avatar del usuario que da el "like"
	userLike, err := t.GetUserByID(idValueToken)
	if err != nil {
		return primitive.NilObjectID, "", err
	}

	// Agregar el idValueToken a los Likes del Tweet
	filter := bson.D{{Key: "_id", Value: TweetId}}
	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "Likes", Value: idValueToken}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return primitive.NilObjectID, "", err
	}

	// Agregar el TweetId a los Likes del usuario
	GoMongoDBColl = t.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	filter = bson.D{{Key: "_id", Value: idValueToken}}
	update = bson.D{{Key: "$addToSet", Value: bson.D{{Key: "Likes", Value: TweetId}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return primitive.NilObjectID, "", err
	}

	// Retornar el UserID y los Avatares de ambos usuarios
	return tweet.UserID, userLike.Avatar, nil
}
func (t *TweetRepository) GetUserByID(userID primitive.ObjectID) (struct {
	UserID primitive.ObjectID `bson:"_id"`
	Avatar string             `bson:"Avatar"`
}, error) {
	GoMongoDBColl := t.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	var user struct {
		UserID primitive.ObjectID `bson:"_id"`
		Avatar string             `bson:"Avatar"`
	}

	err := GoMongoDBColl.FindOne(context.Background(), bson.D{{Key: "_id", Value: userID}}).Decode(&user)
	if err != nil {
		return user, fmt.Errorf("error obteniendo el usuario: %v", err)
	}

	return user, nil
}

func (t *TweetRepository) TweetDislike(TweetId, idValueToken primitive.ObjectID) error {
	GoMongoDBColl := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	count, err := GoMongoDBColl.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: TweetId}})
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("el TweetId no existe")
	}
	filter := bson.D{{Key: "_id", Value: TweetId}}
	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "Likes", Value: idValueToken}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	GoMongoDBColl = t.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	filter = bson.D{{Key: "_id", Value: idValueToken}}
	update = bson.D{{Key: "$pull", Value: bson.D{{Key: "Likes", Value: TweetId}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil

}

// find
func (t *TweetRepository) GetFollowedUsers(idValueObj primitive.ObjectID) (userdomain.User, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	filter := bson.D{{Key: "_id", Value: idValueObj}}
	var user userdomain.User
	err := GoMongoDBCollTweets.FindOne(context.Background(), filter).Decode(&user)
	return user, err
}
func (t *TweetRepository) GetPost(page int) ([]tweetdomain.GetPostcommunitiesRandom, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	skip := (page - 1) * 10
	pipeline := []bson.D{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$sort", Value: bson.D{
			{Key: "TimeStamp", Value: -1},
		}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: 10}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "IsLikedByID", Value: false},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Likes"},
			{Key: "let", Value: bson.D{{Key: "tweetID", Value: "$_id"}}},
			{Key: "pipeline", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{{Key: "$eq", Value: bson.A{"$TweetID", "$$tweetID"}}}},
				}}},
				bson.D{{Key: "$count", Value: "LikesCount"}},
			}},
			{Key: "as", Value: "LikesInfo"},
		}}},
		{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$LikesInfo"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "LikesCount", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$LikesInfo.LikesCount", 0}}}},
		}}},
		// Unimos la información de la comunidad
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "communities"},
			{Key: "localField", Value: "communityID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "CommunityInfo"},
		}}},
		{{Key: "$unwind", Value: "$CommunityInfo"}},

		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "Type", Value: "$Type"},
			{Key: "TimeStamp", Value: "$TimeStamp"},
			{Key: "UserID", Value: "$UserID"},
			{Key: "Likes", Value: "$LikesCount"},
			{Key: "Comments", Value: "$Comments"},
			{Key: "RePosts", Value: "$RePosts"},
			{Key: "Views", Value: "$Views"},
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
			{Key: "CommunityInfo.CommunityName", Value: 1},
			{Key: "CommunityInfo._id", Value: 1},
			{Key: "IsLikedByID", Value: "$IsLikedByID"},
		}}},
	}

	// Ejecutamos el pipeline y recolectamos los datos
	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var tweetsWithUserInfo []tweetdomain.GetPostcommunitiesRandom
	for cursor.Next(context.Background()) {
		var tweetWithUserInfo tweetdomain.GetPostcommunitiesRandom
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}

		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	// Recolectamos los IDs de los OriginalPosts
	var originalPostIDs []primitive.ObjectID
	for _, tweet := range tweetsWithUserInfo {
		if tweet.OriginalPost != primitive.NilObjectID {
			originalPostIDs = append(originalPostIDs, tweet.OriginalPost)
		}
	}

	// Ejecutamos el segundo pipeline si hay originalPostIDs
	if len(originalPostIDs) > 0 {
		originalPostPipeline := []bson.D{
			{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: originalPostIDs}}}}}},
			{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "Users"},
				{Key: "localField", Value: "UserID"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "UserInfo"},
			}}},
			{{Key: "$unwind", Value: "$UserInfo"}},
			{{Key: "$project", Value: bson.D{
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
				{Key: "UserInfo.Online", Value: 1},
			}}},
		}

		cursorOriginalPosts, err := GoMongoDBCollTweets.Aggregate(context.Background(), originalPostPipeline)
		if err != nil {
			return nil, err
		}
		defer cursorOriginalPosts.Close(context.Background())

		var originalPostMap = make(map[primitive.ObjectID]tweetdomain.TweetGetFollowReq)
		for cursorOriginalPosts.Next(context.Background()) {
			var originalPost tweetdomain.TweetGetFollowReq
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return nil, err
			}
			originalPostMap[originalPost.ID] = originalPost
		}

		// Vinculamos los datos del originalPost con los tweets
		for i, tweet := range tweetsWithUserInfo {
			if tweet.OriginalPost != primitive.NilObjectID {
				originalPost, found := originalPostMap[tweet.OriginalPost]
				if found {
					tweetsWithUserInfo[i].OriginalPostData = &originalPost
				}
			}
		}
	}

	return tweetsWithUserInfo, nil
}

func (t *TweetRepository) GetPostId(id primitive.ObjectID) (tweetdomain.TweetGetFollowReq, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	// Ejecutamos el pipeline de agregación para obtener los datos
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: id},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "IsPrivate", Value: false}},
				bson.D{{Key: "IsPrivate", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "Type", Value: "$Type"},
			{Key: "TimeStamp", Value: "$TimeStamp"},
			{Key: "UserID", Value: "$UserID"},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},
			{Key: "CommentsCount", Value: 1},

			{Key: "isLikedByUser", Value: 1},
			{Key: "Comments", Value: "$Comments"},
			{Key: "RePosts", Value: "$RePosts"},
			{Key: "Views", Value: "$Views"},
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
		}}},
	}

	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return tweetdomain.TweetGetFollowReq{}, err
	}
	defer cursor.Close(context.Background())

	var tweetWithUserInfo tweetdomain.TweetGetFollowReq
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return tweetdomain.TweetGetFollowReq{}, err
		}
	}

	// Obtener los datos del OriginalPost si existe
	if tweetWithUserInfo.OriginalPost != primitive.NilObjectID {
		originalPostPipeline := bson.A{
			bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: tweetWithUserInfo.OriginalPost}}}},
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
				{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			}}},
			bson.D{{Key: "$project", Value: bson.D{
				{Key: "id", Value: "$_id"},
				{Key: "Type", Value: "$Type"},
				{Key: "Status", Value: "$Status"},
				{Key: "PostImage", Value: "$PostImage"},
				{Key: "TimeStamp", Value: "$TimeStamp"},
				{Key: "UserID", Value: "$UserID"},
				{Key: "likeCount", Value: 1},
				{Key: "RePostsCount", Value: 1},
				{Key: "CommentsCount", Value: 1},

				{Key: "isLikedByUser", Value: 1},
				{Key: "Comments", Value: "$Comments"},
				{Key: "RePosts", Value: "$RePosts"},
				{Key: "Views", Value: "$Views"},
				{Key: "OriginalPost", Value: "$OriginalPost"},
				{Key: "UserInfo.FullName", Value: 1},
				{Key: "UserInfo.Avatar", Value: 1},
				{Key: "UserInfo.Online", Value: 1},
				{Key: "UserInfo.NameUser", Value: 1},
			}}},
		}

		cursorOriginalPosts, err := GoMongoDBCollTweets.Aggregate(context.Background(), originalPostPipeline)
		if err != nil {
			return tweetdomain.TweetGetFollowReq{}, err
		}
		defer cursorOriginalPosts.Close(context.Background())

		var originalPost tweetdomain.TweetGetFollowReq
		if cursorOriginalPosts.Next(context.Background()) {
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return tweetdomain.TweetGetFollowReq{}, err
			}
			tweetWithUserInfo.OriginalPostData = &originalPost
		}
	}

	return tweetWithUserInfo, nil
}
func (t *TweetRepository) GetPostForAd(id primitive.ObjectID) (tweetdomain.TweetGetFollowReq, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	// Ejecutamos el pipeline de agregación para obtener los datos
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: id},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "Type", Value: "$Type"},
			{Key: "TimeStamp", Value: "$TimeStamp"},
			{Key: "UserID", Value: "$UserID"},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},
			{Key: "CommentsCount", Value: 1},

			{Key: "isLikedByUser", Value: 1},
			{Key: "Comments", Value: "$Comments"},
			{Key: "RePosts", Value: "$RePosts"},
			{Key: "Views", Value: "$Views"},
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
		}}},
	}

	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return tweetdomain.TweetGetFollowReq{}, err
	}
	defer cursor.Close(context.Background())

	var tweetWithUserInfo tweetdomain.TweetGetFollowReq
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return tweetdomain.TweetGetFollowReq{}, err
		}
	}

	// Obtener los datos del OriginalPost si existe
	if tweetWithUserInfo.OriginalPost != primitive.NilObjectID {
		originalPostPipeline := bson.A{
			bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: tweetWithUserInfo.OriginalPost}}}},
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
				{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			}}},
			bson.D{{Key: "$project", Value: bson.D{
				{Key: "id", Value: "$_id"},
				{Key: "Type", Value: "$Type"},
				{Key: "Status", Value: "$Status"},
				{Key: "PostImage", Value: "$PostImage"},
				{Key: "TimeStamp", Value: "$TimeStamp"},
				{Key: "UserID", Value: "$UserID"},
				{Key: "likeCount", Value: 1},
				{Key: "RePostsCount", Value: 1},
				{Key: "CommentsCount", Value: 1},

				{Key: "isLikedByUser", Value: 1},
				{Key: "Comments", Value: "$Comments"},
				{Key: "RePosts", Value: "$RePosts"},
				{Key: "Views", Value: "$Views"},
				{Key: "OriginalPost", Value: "$OriginalPost"},
				{Key: "UserInfo.FullName", Value: 1},
				{Key: "UserInfo.Avatar", Value: 1},
				{Key: "UserInfo.Online", Value: 1},
				{Key: "UserInfo.NameUser", Value: 1},
			}}},
		}

		cursorOriginalPosts, err := GoMongoDBCollTweets.Aggregate(context.Background(), originalPostPipeline)
		if err != nil {
			return tweetdomain.TweetGetFollowReq{}, err
		}
		defer cursorOriginalPosts.Close(context.Background())

		var originalPost tweetdomain.TweetGetFollowReq
		if cursorOriginalPosts.Next(context.Background()) {
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return tweetdomain.TweetGetFollowReq{}, err
			}
			tweetWithUserInfo.OriginalPostData = &originalPost
		}
	}

	return tweetWithUserInfo, nil
}
func (t *TweetRepository) GetPostIdLogueado(id primitive.ObjectID, userID primitive.ObjectID) (tweetdomain.TweetGetFollowReq, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: id}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "isLikedByUser", Value: bson.D{{Key: "$in", Value: bson.A{userID, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
		}}},

		// usuario token
		{{Key: "$addFields", Value: bson.D{
			{Key: "userIdParam", Value: userID},
		}}},

		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "userIdParam"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfoParam"},
		}}},
		{{Key: "$unwind", Value: "$UserInfoParam"}},
		{{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{
				{Key: "$or", Value: bson.A{
					// Si es público o no tiene la propiedad IsPrivate
					bson.D{{Key: "$eq", Value: bson.A{"$IsPrivate", false}}},
					// Si es privado, verificar que sea miembro de la comunidad
					bson.D{{Key: "$and", Value: bson.A{
						bson.D{{Key: "$eq", Value: bson.A{"$IsPrivate", true}}},
						bson.D{{Key: "$in", Value: bson.A{"$communityID", "$UserInfoParam.InCommunities"}}},
					}}},
				}},
			}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "Type", Value: "$Type"},
			{Key: "TimeStamp", Value: "$TimeStamp"},
			{Key: "UserID", Value: "$UserID"},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "communityID", Value: 1},
			{Key: "isLikedByUser", Value: 1},
			{Key: "Comments", Value: "$Comments"},
			{Key: "RePosts", Value: "$RePosts"},
			{Key: "Views", Value: "$Views"},
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
			{Key: "UserInfo.InCommunities", Value: 1},
		}}},
	}

	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return tweetdomain.TweetGetFollowReq{}, err
	}
	defer cursor.Close(context.Background())

	var tweetWithUserInfo tweetdomain.TweetGetFollowReq
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return tweetdomain.TweetGetFollowReq{}, err
		}
	}

	// Obtener los datos del OriginalPost si existe
	if tweetWithUserInfo.OriginalPost != primitive.NilObjectID {
		originalPostPipeline := bson.A{
			bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: tweetWithUserInfo.OriginalPost}}}},
			bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "Users"},
				{Key: "localField", Value: "UserID"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "UserInfo"},
			}}},
			bson.D{{Key: "$unwind", Value: "$UserInfo"}},
			bson.D{{Key: "$addFields", Value: bson.D{
				{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
				{Key: "isLikedByUser", Value: bson.D{{Key: "$in", Value: bson.A{userID, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
				{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
				{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			}}},
			bson.D{{Key: "$project", Value: bson.D{
				{Key: "id", Value: "$_id"},
				{Key: "Type", Value: "$Type"},
				{Key: "Status", Value: "$Status"},
				{Key: "PostImage", Value: "$PostImage"},
				{Key: "TimeStamp", Value: "$TimeStamp"},
				{Key: "UserID", Value: "$UserID"},
				{Key: "likeCount", Value: 1},
				{Key: "RePostsCount", Value: 1},
				{Key: "CommentsCount", Value: 1},
				{Key: "isLikedByUser", Value: 1},
				{Key: "Comments", Value: "$Comments"},
				{Key: "RePosts", Value: "$RePosts"},
				{Key: "Views", Value: "$Views"},
				{Key: "OriginalPost", Value: "$OriginalPost"},
				{Key: "UserInfo.FullName", Value: 1},
				{Key: "UserInfo.Avatar", Value: 1},
				{Key: "UserInfo.NameUser", Value: 1},
				{Key: "UserInfo.Online", Value: 1},
			}}},
		}

		cursorOriginalPosts, err := GoMongoDBCollTweets.Aggregate(context.Background(), originalPostPipeline)
		if err != nil {
			return tweetdomain.TweetGetFollowReq{}, err
		}
		defer cursorOriginalPosts.Close(context.Background())

		var originalPost tweetdomain.TweetGetFollowReq
		if cursorOriginalPosts.Next(context.Background()) {
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return tweetdomain.TweetGetFollowReq{}, err
			}
			tweetWithUserInfo.OriginalPostData = &originalPost
		}
	}
	var posts []tweetdomain.TweetGetFollowReq
	posts = append(posts, tweetWithUserInfo)

	if err := t.updateTweetViews(context.TODO(), GoMongoDBCollTweets, posts, id); err != nil {
		return tweetdomain.TweetGetFollowReq{}, err
	}

	return tweetWithUserInfo, nil
}

func (t *TweetRepository) GetPostuser(page int, id primitive.ObjectID, limit int) ([]tweetdomain.GetPostcommunitiesRandom, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	skip := (page - 1) * 10
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{
			{Key: "UserID", Value: id},

			{Key: "$or", Value: bson.A{
				bson.D{{Key: "communityID", Value: primitive.NilObjectID}},
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "IsPrivate", Value: false}},
				bson.D{{Key: "IsPrivate", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$sort", Value: bson.D{
			{Key: "TimeStamp", Value: -1},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
		}}},

		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "Type", Value: "$Type"},
			{Key: "TimeStamp", Value: "$TimeStamp"},
			{Key: "UserID", Value: "$UserID"},
			{Key: "Likes", Value: "$Likes"},
			{Key: "Comments", Value: "$Comments"},
			{Key: "RePosts", Value: "$RePosts"},
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "Views", Value: "$Views"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
		}}},
	}

	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var tweetsWithUserInfo []tweetdomain.GetPostcommunitiesRandom
	var originalPostIDs []primitive.ObjectID
	for cursor.Next(context.Background()) {
		var tweetWithUserInfo tweetdomain.GetPostcommunitiesRandom
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}

		if tweetWithUserInfo.OriginalPost != primitive.NilObjectID {
			originalPostIDs = append(originalPostIDs, tweetWithUserInfo.OriginalPost)
		}

		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	// Obtener los datos del OriginalPost si existen
	if len(originalPostIDs) > 0 {
		originalPostPipeline := bson.A{
			bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: originalPostIDs}}}}}},
			bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "Users"},
				{Key: "localField", Value: "UserID"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "UserInfo"},
			}}},
			bson.D{{Key: "$addFields", Value: bson.D{
				{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
				{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
				{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			}}},
			bson.D{{Key: "$unwind", Value: "$UserInfo"}},
			bson.D{{Key: "$project", Value: bson.D{
				{Key: "likeCount", Value: 1},
				{Key: "RePostsCount", Value: 1},
				{Key: "CommentsCount", Value: 1},
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
				{Key: "UserInfo.Online", Value: 1},
				{Key: "CommunityInfo.CommunityName", Value: 1},
				{Key: "CommunityInfo._id", Value: 1},
			}}},
		}

		cursorOriginalPosts, err := GoMongoDBCollTweets.Aggregate(context.Background(), originalPostPipeline)
		if err != nil {
			return nil, err
		}
		defer cursorOriginalPosts.Close(context.Background())

		var originalPostMap = make(map[primitive.ObjectID]tweetdomain.TweetGetFollowReq)
		for cursorOriginalPosts.Next(context.Background()) {
			var originalPost tweetdomain.TweetGetFollowReq
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return nil, err
			}
			originalPostMap[originalPost.ID] = originalPost
		}
		for i, tweet := range tweetsWithUserInfo {
			if tweet.OriginalPost != primitive.NilObjectID {
				originalPost, found := originalPostMap[tweet.OriginalPost]
				if found {
					tweetsWithUserInfo[i].OriginalPostData = &originalPost
				}
			}
		}
	}

	return tweetsWithUserInfo, nil
}

func (t *TweetRepository) GetPostuserLogueado(page int, id, idt primitive.ObjectID, limit int) ([]tweetdomain.GetPostcommunitiesRandom, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	skip := (page - 1) * 10
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{
			{Key: "UserID", Value: id},

			{Key: "$and", Value: bson.A{
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$exists", Value: true}}}},
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$ne", Value: primitive.NilObjectID}}}}}},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "IsPrivate", Value: false}},
				bson.D{{Key: "IsPrivate", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idt, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "TimeStamp", Value: -1},
		}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "Type", Value: "$Type"},
			{Key: "TimeStamp", Value: "$TimeStamp"},
			{Key: "UserID", Value: "$UserID"},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},
			{Key: "isLikedByID", Value: 1},
			{Key: "CommentsCount", Value: 1},
			{Key: "CommunityInfo.CommunityName", Value: 1},
			{Key: "CommunityInfo._id", Value: 1},
			{Key: "Comments", Value: "$Comments"},
			{Key: "RePosts", Value: "$RePosts"},
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "Views", Value: "$Views"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
		}}},
	}

	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var tweetsWithUserInfo []tweetdomain.GetPostcommunitiesRandom
	var originalPostIDs []primitive.ObjectID
	for cursor.Next(context.Background()) {
		var tweetWithUserInfo tweetdomain.GetPostcommunitiesRandom
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}

		if tweetWithUserInfo.OriginalPost != primitive.NilObjectID {
			originalPostIDs = append(originalPostIDs, tweetWithUserInfo.OriginalPost)
		}

		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	// Incrementar Views en 1 solo para los documentos obtenidos
	if err := t.PostcommunitiesRandomAddView(context.TODO(), GoMongoDBCollTweets, tweetsWithUserInfo, idt); err != nil {
		return tweetsWithUserInfo, err
	}

	// Obtener los datos del OriginalPost si existen
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
			bson.D{{Key: "$addFields", Value: bson.D{
				{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
				{Key: "isLikedByID", Value: bson.D{{Key: "$in", Value: bson.A{idt, bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}}},
				{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
				{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			}}},
			bson.D{{Key: "$project", Value: bson.D{
				{Key: "id", Value: "$_id"},
				{Key: "Type", Value: "$Type"},
				{Key: "Status", Value: "$Status"},
				{Key: "PostImage", Value: "$PostImage"},
				{Key: "TimeStamp", Value: "$TimeStamp"},
				{Key: "UserID", Value: "$UserID"},
				{Key: "Likes", Value: "$Likes"},
				{Key: "CommentsCount", Value: 1},
				{Key: "likeCount", Value: 1},
				{Key: "RePostsCount", Value: 1},
				{Key: "isLikedByID", Value: 1},
				{Key: "Comments", Value: "$Comments"},
				{Key: "CommunityInfo.CommunityName", Value: 1},
				{Key: "CommunityInfo._id", Value: 1},
				{Key: "Views", Value: "$Views"},
				{Key: "OriginalPost", Value: "$OriginalPost"},
				{Key: "UserInfo.FullName", Value: 1},
				{Key: "UserInfo.Avatar", Value: 1},
				{Key: "UserInfo.NameUser", Value: 1},
				{Key: "UserInfo.Online", Value: 1},
			}}},
		}

		cursorOriginalPosts, err := GoMongoDBCollTweets.Aggregate(context.Background(), originalPostPipeline)
		if err != nil {
			return nil, err
		}
		defer cursorOriginalPosts.Close(context.Background())

		var originalPostMap = make(map[primitive.ObjectID]tweetdomain.TweetGetFollowReq)
		for cursorOriginalPosts.Next(context.Background()) {
			var originalPost tweetdomain.TweetGetFollowReq
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return nil, err
			}
			originalPostMap[originalPost.ID] = originalPost
		}

		for i, tweet := range tweetsWithUserInfo {
			if tweet.OriginalPost != primitive.NilObjectID {
				originalPost, found := originalPostMap[tweet.OriginalPost]
				if found {
					tweetsWithUserInfo[i].OriginalPostData = &originalPost
				}
			}
		}
	}

	return tweetsWithUserInfo, nil
}
func (t *TweetRepository) GetPostsWithImages(page int, id primitive.ObjectID, limit int) ([]tweetdomain.GetPostcommunitiesRandom, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	skip := (page - 1) * 10
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{
			{Key: "UserID", Value: id},
			{Key: "Type", Value: bson.M{"$in": []string{"Post", "CitaPost"}}},
			{Key: "PostImage", Value: bson.D{{Key: "$ne", Value: ""}}},

			{Key: "$and", Value: bson.A{
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$exists", Value: true}}}},
				bson.D{{Key: "communityID", Value: bson.D{{Key: "$ne", Value: primitive.NilObjectID}}}}}},
			{Key: "$or", Value: bson.A{
				bson.D{{Key: "IsPrivate", Value: false}},
				bson.D{{Key: "IsPrivate", Value: bson.D{{Key: "$exists", Value: false}}}},
			}},
		}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$sort", Value: bson.D{
			{Key: "TimeStamp", Value: -1},
		}}},
		// Agregar campos adicionales como contador de likes, comentarios, reposts
		{{Key: "$addFields", Value: bson.D{
			{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
			{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
			{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
		}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: limit}},
		// Seleccionar solo los campos necesarios
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "Type", Value: "$Type"},
			{Key: "TimeStamp", Value: "$TimeStamp"},
			{Key: "UserID", Value: "$UserID"},
			{Key: "Likes", Value: "$Likes"},
			{Key: "Comments", Value: "$Comments"},
			{Key: "RePosts", Value: "$RePosts"},
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "Views", Value: "$Views"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
			{Key: "likeCount", Value: 1},
			{Key: "RePostsCount", Value: 1},
			{Key: "CommentsCount", Value: 1},
		}}},
	}

	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var tweetsWithUserInfo []tweetdomain.GetPostcommunitiesRandom
	var originalPostIDs []primitive.ObjectID
	for cursor.Next(context.Background()) {
		var tweetWithUserInfo tweetdomain.GetPostcommunitiesRandom
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}

		if tweetWithUserInfo.OriginalPost != primitive.NilObjectID {
			originalPostIDs = append(originalPostIDs, tweetWithUserInfo.OriginalPost)
		}

		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	// Incrementar las vistas en 1
	// if err := t.PostcommunitiesRandomAddView(context.TODO(), GoMongoDBCollTweets, tweetsWithUserInfo, idt); err != nil {
	// 	return tweetsWithUserInfo, err
	// }

	// Obtener los datos del OriginalPost si existen
	if len(originalPostIDs) > 0 {
		originalPostPipeline := bson.A{
			bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: originalPostIDs}}}}}},
			bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "Users"},
				{Key: "localField", Value: "UserID"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "UserInfo"},
			}}},
			bson.D{{Key: "$addFields", Value: bson.D{
				{Key: "likeCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Likes", bson.A{}}}}}}},
				{Key: "CommentsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$Comments", bson.A{}}}}}}},
				{Key: "RePostsCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$RePosts", bson.A{}}}}}}},
			}}},
			bson.D{{Key: "$unwind", Value: "$UserInfo"}},
			bson.D{{Key: "$project", Value: bson.D{
				{Key: "likeCount", Value: 1},
				{Key: "RePostsCount", Value: 1},
				{Key: "CommentsCount", Value: 1},
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
				{Key: "UserInfo.Online", Value: 1},
			}}},
		}

		cursorOriginalPosts, err := GoMongoDBCollTweets.Aggregate(context.Background(), originalPostPipeline)
		if err != nil {
			return nil, err
		}
		defer cursorOriginalPosts.Close(context.Background())

		var originalPostMap = make(map[primitive.ObjectID]tweetdomain.TweetGetFollowReq)
		for cursorOriginalPosts.Next(context.Background()) {
			var originalPost tweetdomain.TweetGetFollowReq
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return nil, err
			}
			originalPostMap[originalPost.ID] = originalPost
		}

		for i, tweet := range tweetsWithUserInfo {
			if tweet.OriginalPost != primitive.NilObjectID {
				originalPost, found := originalPostMap[tweet.OriginalPost]
				if found {
					tweetsWithUserInfo[i].OriginalPostData = &originalPost
				}
			}
		}
	}

	return tweetsWithUserInfo, nil
}

// Función de utilidad para extraer los IDs de los tweets

func (t *TweetRepository) GetTweetsLast24HoursFollow(userIDs []primitive.ObjectID, page int) ([]tweetdomain.TweetGetFollowReq, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	currentTime := time.Now()
	last24Hours := currentTime.Add(-24 * time.Hour)
	skip := (page - 1) * 13
	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "UserID", Value: bson.D{{Key: "$in", Value: userIDs}}},
		{Key: "TimeStamp", Value: bson.D{{Key: "$gte", Value: last24Hours}}},

		{Key: "$or", Value: bson.A{
			bson.D{{Key: "communityID", Value: primitive.NilObjectID}},
			bson.D{{Key: "communityID", Value: bson.D{{Key: "$exists", Value: false}}}},
		}},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "IsPrivate", Value: false}},
			bson.D{{Key: "IsPrivate", Value: bson.D{{Key: "$exists", Value: false}}}},
		}},
	}}}

	// Actualizamos todos los documentos coincidentes incrementando el campo Views en 1
	_, err := GoMongoDBCollTweets.UpdateMany(context.Background(), matchStage, bson.D{
		{Key: "$inc", Value: bson.D{{Key: "Views", Value: 1}}},
	})
	if err != nil {
		return nil, err
	}

	pipeline := []bson.D{
		matchStage,
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$sort", Value: bson.D{
			{Key: "TimeStamp", Value: -1},
		}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: 13}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "Type", Value: "$Type"},
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
			{Key: "UserInfo.Online", Value: 1},
		}}},
	}

	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var tweetsWithUserInfo []tweetdomain.TweetGetFollowReq
	for cursor.Next(context.Background()) {
		var tweetWithUserInfo tweetdomain.TweetGetFollowReq
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}

		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	var originalPostIDs []primitive.ObjectID
	for _, tweet := range tweetsWithUserInfo {
		if tweet.OriginalPost != primitive.NilObjectID {
			originalPostIDs = append(originalPostIDs, tweet.OriginalPost)
		}
	}
	if len(originalPostIDs) > 0 {
		originalPostPipeline := []bson.D{
			{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: originalPostIDs}}}}}},
			{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "Users"},
				{Key: "localField", Value: "UserID"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "UserInfo"},
			}}},
			{{Key: "$unwind", Value: "$UserInfo"}},
			{{Key: "$project", Value: bson.D{
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
				{Key: "UserInfo.Online", Value: 1},
				{Key: "UserInfo.NameUser", Value: 1},
			}}},
		}

		cursorOriginalPosts, err := GoMongoDBCollTweets.Aggregate(context.Background(), originalPostPipeline)
		if err != nil {
			return nil, err
		}
		defer cursorOriginalPosts.Close(context.Background())

		var originalPostMap = make(map[primitive.ObjectID]tweetdomain.TweetGetFollowReq)
		for cursorOriginalPosts.Next(context.Background()) {
			var originalPost tweetdomain.TweetGetFollowReq
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return nil, err
			}
			originalPostMap[originalPost.ID] = originalPost
		}

		for i, tweet := range tweetsWithUserInfo {
			if tweet.OriginalPost != primitive.NilObjectID {
				originalPost, found := originalPostMap[tweet.OriginalPost]
				if found {
					tweetsWithUserInfo[i].OriginalPostData = &originalPost
				}
			}
		}
	}

	return tweetsWithUserInfo, nil
}

func (t *TweetRepository) GetCommentPosts(tweetID primitive.ObjectID, page int, idt primitive.ObjectID) ([]tweetdomain.TweetCommentsGetReq, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	var tweet tweetdomain.Post
	if err := GoMongoDBCollTweets.FindOne(context.Background(), bson.M{"_id": tweetID}).Decode(&tweet); err != nil {
		return nil, err
	}

	skip := (page - 1) * 10

	commentIDs := tweet.Comments

	// Pipeline para obtener los comentarios
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: commentIDs}}}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$sort", Value: bson.D{
			{Key: "TimeStamp", Value: -1},
		}}},
		{{Key: "$skip", Value: skip}},
		{{Key: "$limit", Value: 10}},
		// Proyecta los campos necesarios
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "CommentBy", Value: "$CommentBy"},
			{Key: "Comments", Value: "$Comments"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "TimeStamp", Value: "$TimeStamp"},
			{Key: "UserID", Value: "$UserID"},
			{Key: "Views", Value: "$Views"},
			{Key: "Likes", Value: "$Likes"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
		}}},
	}

	var comments []tweetdomain.TweetCommentsGetReq

	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var comment tweetdomain.TweetCommentsGetReq
		if err := cursor.Decode(&comment); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := t.updatePostCommentViews(context.TODO(), GoMongoDBCollTweets, comments, idt); err != nil {
		return comments, err
	}

	return comments, nil
}
func (t *TweetRepository) isOriginalPostPrivate(originalPostID primitive.ObjectID) (tweetdomain.PostDataOp, error) {
	GoMongoDBColl := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	dataPost := tweetdomain.PostDataOp{}
	err := GoMongoDBColl.FindOne(context.Background(), bson.D{{Key: "_id", Value: originalPostID}}).Decode(&dataPost)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return dataPost, errors.New("el post original no existe")
		}
		return dataPost, err
	}

	return dataPost, nil
}

func (t *TweetRepository) RePost(rePost *tweetdomain.RePost) error {
	GoMongoDBColl := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	Post, err := t.isOriginalPostPrivate(rePost.OriginalPost)
	if err != nil {
		return err
	}
	rePost.IsPrivate = false

	if Post.IsPrivate {
		rePost.IsPrivate = true
		member, _, err := t.IsUserMemberOfCommunity(Post.CommunityID, rePost.UserID)
		if err != nil {
			return errors.New("error buscando al member")
		}
		if !member {
			return errors.New("no se puede repostear un post privado")
		}
	}
	rePost.CommunityID = Post.CommunityID

	filterRePost := bson.D{{Key: "UserID", Value: rePost.UserID}, {Key: "OriginalPost", Value: rePost.OriginalPost}}
	existingRePost := GoMongoDBColl.FindOne(context.Background(), filterRePost)

	if existingRePost.Err() != nil {
		if existingRePost.Err() == mongo.ErrNoDocuments {
			updateResult, errAdd := GoMongoDBColl.UpdateOne(
				context.Background(),
				bson.D{{Key: "_id", Value: rePost.OriginalPost}},
				bson.D{{Key: "$addToSet", Value: bson.D{{Key: "RePosts", Value: rePost.UserID}}}},
			)
			if errAdd != nil {
				return errAdd
			}
			if updateResult.ModifiedCount == 0 {
				return errors.New("NoDocuments")
			}
			_, errInsertOne := GoMongoDBColl.InsertOne(context.Background(), rePost)
			return errInsertOne
		}
		return existingRePost.Err()

	}

	_, errDelete := GoMongoDBColl.DeleteOne(context.Background(), filterRePost)
	if errDelete != nil {
		return errDelete
	}
	_, errPull := GoMongoDBColl.UpdateOne(
		context.Background(),
		bson.D{{Key: "_id", Value: rePost.OriginalPost}},
		bson.D{{Key: "$pull", Value: bson.D{{Key: "RePosts", Value: rePost.UserID}}}},
	)
	return errPull

}
func (t *TweetRepository) CitaPost(rePost *tweetdomain.CitaPost) error {
	GoMongoDBColl := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	Post, err := t.isOriginalPostPrivate(rePost.OriginalPost)
	if err != nil {
		return err
	}
	rePost.IsPrivate = false

	if Post.IsPrivate {
		rePost.IsPrivate = true
		member, _, err := t.IsUserMemberOfCommunity(Post.CommunityID, rePost.UserID)
		if err != nil {
			return errors.New("error buscando al member")
		}
		if !member {
			return errors.New("no se puede repostear un post privado")
		}
	}
	rePost.CommunityID = Post.CommunityID

	_, err = GoMongoDBColl.InsertOne(context.Background(), rePost)
	return err
}

func (t *TweetRepository) GetTrends(page int, limit int) ([]tweetdomain.Trend, error) {
	GoMongoDB := t.mongoClient.Database("PINKKER-BACKEND").Collection("Trends")

	options := options.Find().SetSort(bson.D{{Key: "count", Value: -1}}).SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	cursor, err := GoMongoDB.Find(context.Background(), bson.M{}, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var trends []tweetdomain.Trend
	if err := cursor.All(context.Background(), &trends); err != nil {
		return nil, err
	}

	return trends, nil
}

func (t *TweetRepository) GetTweetsByHashtag(hashtag string, page int, limit int) ([]tweetdomain.TweetGetFollowReq, error) {
	ctx := context.Background()
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	// Actualizar el campo Views sumando 1
	updateViews := bson.D{
		{Key: "$inc", Value: bson.D{{Key: "Views", Value: 1}}},
	}

	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "Hashtags", Value: hashtag}}}},
		{{Key: "$sort", Value: bson.D{{Key: "Likes", Value: -1}}}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		{{Key: "$unwind", Value: "$UserInfo"}},
		{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: "$Status"},
			{Key: "PostImage", Value: "$PostImage"},
			{Key: "Type", Value: "$Type"},
			{Key: "TimeStamp", Value: "$TimeStamp"},
			{Key: "UserID", Value: "$UserID"},
			{Key: "Likes", Value: "$Likes"},
			{Key: "Comments", Value: "$Comments"},
			{Key: "RePosts", Value: "$RePosts"},
			{Key: "Hashtags", Value: "$Hashtags"},
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "Views", Value: "$Views"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
			{Key: "UserInfo.Online", Value: 1},
		}}},
		{{Key: "$skip", Value: int64((page - 1) * limit)}}, // Saltar resultados según la paginación
		{{Key: "$limit", Value: int64(limit)}},             // Limitar resultados según la paginación
	}

	// Actualizar el campo Views sumando 1 antes de ejecutar el pipeline de agregación
	if _, err := GoMongoDBCollTweets.UpdateMany(ctx, bson.M{"Hashtags": hashtag}, updateViews); err != nil {
		return nil, err
	}

	cursor, err := GoMongoDBCollTweets.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tweets []tweetdomain.TweetGetFollowReq
	if err := cursor.All(ctx, &tweets); err != nil {
		return nil, err
	}

	return tweets, nil
}

func (t *TweetRepository) GetTrendsByPrefix(prefix string, limit int) ([]tweetdomain.Trend, error) {
	GoMongoDB := t.mongoClient.Database("PINKKER-BACKEND").Collection("Trends")

	regex := primitive.Regex{Pattern: "^" + prefix, Options: "i"}

	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "Hashtags", Value: regex}}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := GoMongoDB.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var trends []tweetdomain.Trend
	if err := cursor.All(context.Background(), &trends); err != nil {
		return nil, err
	}

	return trends, nil
}
func (t *TweetRepository) IsUserMemberOfCommunity(communityID, userID primitive.ObjectID) (bool, bool, error) {
	collCommunities := t.mongoClient.Database("PINKKER-BACKEND").Collection("communities")
	filter := bson.D{
		{Key: "_id", Value: communityID},
		{Key: "Members", Value: bson.D{{Key: "$in", Value: bson.A{userID}}}},
	}

	var community struct {
		ID        primitive.ObjectID `bson:"_id"`
		IsPrivate bool               `bson:"IsPrivate"`
	}

	err := collCommunities.FindOne(context.Background(), filter).Decode(&community)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, false, nil
		}
		return false, false, err
	}

	return true, community.IsPrivate, nil
}

func (repo *TweetRepository) GetCommunityPosts(ctx context.Context, communityID primitive.ObjectID, ExcludeFilterlID []primitive.ObjectID, idT primitive.ObjectID) ([]tweetdomain.GetPostcommunitiesRandom, error) {
	db := repo.mongoClient.Database("PINKKER-BACKEND")
	collPosts := db.Collection("Post")
	_, err := repo.GetSubscriptionByUserIDsAndcommunityID(idT, communityID, idT)
	if err != nil {
		return nil, err
	}
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
		bson.D{{Key: "$limit", Value: 15}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: 1},
			{Key: "PostImage", Value: 1},
			{Key: "Type", Value: 1},
			{Key: "Views", Value: 1},
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
			{Key: "OriginalPost", Value: 1},
		}}},
	}

	cursor, err := collPosts.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var postsWithUserInfo []tweetdomain.GetPostcommunitiesRandom
	for cursor.Next(ctx) {
		var postWithUserInfo tweetdomain.GetPostcommunitiesRandom
		if err := cursor.Decode(&postWithUserInfo); err != nil {
			return nil, err
		}
		postsWithUserInfo = append(postsWithUserInfo, postWithUserInfo)
	}
	if err := repo.addOriginalPostDataCommunity(ctx, collPosts, postsWithUserInfo); err != nil {
		return nil, err
	}
	if err := repo.PostcommunitiesRandomAddView(context.TODO(), collPosts, postsWithUserInfo, idT); err != nil {
		return nil, err
	}
	return postsWithUserInfo, nil
}
func (repo *TweetRepository) GetCommunityPostsGallery(ctx context.Context, communityID primitive.ObjectID, ExcludeFilterlID []primitive.ObjectID, idT primitive.ObjectID) ([]tweetdomain.GetPostcommunitiesRandom, error) {
	db := repo.mongoClient.Database("PINKKER-BACKEND")
	collPosts := db.Collection("Post")
	_, err := repo.GetSubscriptionByUserIDsAndcommunityID(idT, communityID, idT)
	if err != nil {
		return nil, err
	}
	excludeFilter := bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: ExcludeFilterlID}}}}

	includeFilter := bson.D{
		{Key: "communityID", Value: communityID},
		{Key: "Type", Value: bson.M{"$in": []string{"Post", "RePost", "CitaPost"}}},
		{Key: "PostImage", Value: bson.D{{Key: "$ne", Value: ""}}},
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
		bson.D{{Key: "$limit", Value: 15}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: 1},
			{Key: "PostImage", Value: 1},
			{Key: "Type", Value: 1},
			{Key: "Views", Value: 1},
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
			{Key: "OriginalPost", Value: 1},
		}}},
	}

	cursor, err := collPosts.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var postsWithUserInfo []tweetdomain.GetPostcommunitiesRandom
	for cursor.Next(ctx) {
		var postWithUserInfo tweetdomain.GetPostcommunitiesRandom
		if err := cursor.Decode(&postWithUserInfo); err != nil {
			return nil, err
		}
		postsWithUserInfo = append(postsWithUserInfo, postWithUserInfo)
	}
	if err := repo.addOriginalPostDataCommunity(ctx, collPosts, postsWithUserInfo); err != nil {
		return nil, err
	}
	if err := repo.PostcommunitiesRandomAddView(context.TODO(), collPosts, postsWithUserInfo, idT); err != nil {
		return nil, err
	}
	return postsWithUserInfo, nil
}
func (r *TweetRepository) GetSubscriptionByUserIDsAndcommunityID(sourceUserID, communityID, idT primitive.ObjectID) (*subscriptiondomain.Subscription, error) {
	communitiesCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("communities")

	redisKey := fmt.Sprintf("community:%s", communityID.Hex())
	redisData, err := r.redisClient.Get(context.Background(), redisKey).Result()

	var community struct {
		CreatorID primitive.ObjectID `bson:"CreatorID"`
		IsPrivate bool               `bson:"IsPrivate"`
		IsPaid    bool               `bson:"IsPaid"`
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
	if community.IsPaid {
		filter := bson.M{
			"_id":     communityID,
			"Members": bson.M{"$in": []primitive.ObjectID{idT}}, // Verificar si idT está en Members
		}

		// Ejecutar la búsqueda
		err = communitiesCollection.FindOne(context.Background(), filter).Decode(&community)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, fmt.Errorf("el usuario no es miembro de la comunidad")
			}
			return nil, fmt.Errorf("error al buscar membresía en la comunidad: %v", err)
		}
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

// views
func (t *TweetRepository) PostcommunitiesRandomAddView(ctx context.Context, collTweets *mongo.Collection, tweets []tweetdomain.GetPostcommunitiesRandom, idt primitive.ObjectID) error {
	// Obtener los IDs de los tweets
	var tweetIDs []primitive.ObjectID
	for _, tweet := range tweets {
		tweetIDs = append(tweetIDs, tweet.ID)
	}

	// Asegurarse de que haya tweets para actualizar
	if len(tweetIDs) > 0 {
		// Crear un único filtro para todos los tweets
		filter := bson.M{
			"_id":                   bson.M{"$in": tweetIDs}, // Filtro para los tweets seleccionados
			"IdOfTheUsersWhoViewed": bson.M{"$ne": idt},      // Solo actualizamos si el usuario no ha visto ya el tweet
		}

		// Actualización que agrega el ID del usuario y mantiene el límite de 50 IDs
		update := bson.M{
			"$push": bson.M{
				"IdOfTheUsersWhoViewed": bson.M{
					"$each":     []primitive.ObjectID{idt}, // Agregar ID del usuario actual
					"$position": -1,                        // Agregar al final del array
					"$slice":    -50,                       // Mantener solo los últimos 50 usuarios
				},
			},
			"$inc": bson.M{
				"Views": 1, // Incrementar las vistas
			},
		}

		// Usamos UpdateMany para actualizar todos los tweets de una sola vez
		_, err := collTweets.UpdateMany(ctx, filter, update)
		if err != nil {
			return err
		}
	}

	return nil
}
func (t *TweetRepository) updateTweetViews(ctx context.Context, collTweets *mongo.Collection, tweets []tweetdomain.TweetGetFollowReq, idt primitive.ObjectID) error {
	// Obtener los IDs de los tweets
	var tweetIDs []primitive.ObjectID
	for _, tweet := range tweets {
		tweetIDs = append(tweetIDs, tweet.ID)
	}

	// Asegurarse de que haya tweets para actualizar
	if len(tweetIDs) > 0 {
		// Crear un único filtro para todos los tweets
		filter := bson.M{
			"_id":                   bson.M{"$in": tweetIDs}, // Filtro para los tweets seleccionados
			"IdOfTheUsersWhoViewed": bson.M{"$ne": idt},      // Solo actualizamos si el usuario no ha visto ya el tweet
		}

		// Actualización que agrega el ID del usuario y mantiene el límite de 50 IDs
		update := bson.M{
			"$push": bson.M{
				"IdOfTheUsersWhoViewed": bson.M{
					"$each":     []primitive.ObjectID{idt}, // Agregar ID del usuario actual
					"$position": -1,                        // Agregar al final del array
					"$slice":    -50,                       // Mantener solo los últimos 50 usuarios
				},
			},
			"$inc": bson.M{
				"Views": 1, // Incrementar las vistas
			},
		}

		// Usamos UpdateMany para actualizar todos los tweets de una sola vez
		_, err := collTweets.UpdateMany(ctx, filter, update)
		if err != nil {
			return err
		}
	}

	return nil
}
func (t *TweetRepository) updatePostCommentViews(ctx context.Context, collTweets *mongo.Collection, tweets []tweetdomain.TweetCommentsGetReq, idt primitive.ObjectID) error {
	// Obtener los IDs de los tweets
	var tweetIDs []primitive.ObjectID
	for _, tweet := range tweets {
		tweetIDs = append(tweetIDs, tweet.ID)
	}

	// Asegurarse de que haya tweets para actualizar
	if len(tweetIDs) > 0 {
		// Crear un único filtro para todos los tweets
		filter := bson.M{
			"_id":                   bson.M{"$in": tweetIDs}, // Filtro para los tweets seleccionados
			"IdOfTheUsersWhoViewed": bson.M{"$ne": idt},      // Solo actualizamos si el usuario no ha visto ya el tweet
		}

		// Actualización que agrega el ID del usuario y mantiene el límite de 50 IDs
		update := bson.M{
			"$push": bson.M{
				"IdOfTheUsersWhoViewed": bson.M{
					"$each":     []primitive.ObjectID{idt}, // Agregar ID del usuario actual
					"$position": -1,                        // Agregar al final del array
					"$slice":    -50,                       // Mantener solo los últimos 50 usuarios
				},
			},
			"$inc": bson.M{
				"Views": 1, // Incrementar las vistas
			},
		}

		// Usamos UpdateMany para actualizar todos los tweets de una sola vez
		_, err := collTweets.UpdateMany(ctx, filter, update)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TweetRepository) SaveNotification(userID primitive.ObjectID, notification notificationsdomain.Notification) error {
	// Colección de notificaciones
	ctx := context.Background()
	db := r.mongoClient.Database("PINKKER-BACKEND")

	notificationsCollection := db.Collection("Notifications")

	// Asegurar que la notificación tenga una marca de tiempo
	if notification.Timestamp.IsZero() {
		notification.Timestamp = time.Now()
	}

	// Filtro para buscar el documento del usuario
	filter := bson.M{"userId": userID}

	// Actualización para agregar la notificación y crear el documento si no existe
	update := bson.M{
		"$push":        bson.M{"notifications": notification}, // Agrega la notificación al array
		"$setOnInsert": bson.M{"userId": userID},              // Crea el documento si no existe
	}

	// Realizar la operación con upsert
	_, err := notificationsCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("error al guardar la notificación: %v", err)
	}

	return nil
}

// get notificaciones
func (u *TweetRepository) GetWebSocketClientsInRoom(roomID string) ([]*websocket.Conn, error) {
	clients, err := utils.NewChatService().GetWebSocketClientsInRoom(roomID)

	return clients, err
}
