package tweetinfrastructure

import (
	tweetdomain "PINKKER-BACKEND/internal/tweet/tweet-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"errors"
	"fmt"
	"time"

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

// Save
func (t *TweetRepository) TweetSave(Tweet tweetdomain.Post) error {
	GoMongoDBCollUsers := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	_, errInsertOne := GoMongoDBCollUsers.InsertOne(context.Background(), Tweet)
	if errInsertOne != nil {
		return errInsertOne
	}
	return nil
}
func (t *TweetRepository) UpdateTrends(hashtags []string) error {
	GoMongoDB := t.mongoClient.Database("PINKKER-BACKEND").Collection("Trends")
	for _, hashtag := range hashtags {
		filter := bson.M{"hashtag": hashtag}
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
func (t *TweetRepository) SaveComment(tweetComment *tweetdomain.PostComment) error {
	GoMongoDBCollComments := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	res, errInsertOne := GoMongoDBCollComments.InsertOne(context.Background(), tweetComment)
	if errInsertOne != nil {
		return errInsertOne
	}

	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	filter := bson.M{"_id": tweetComment.OriginalPost}
	update := bson.M{"$push": bson.M{"Comments": res.InsertedID}}

	_, err := GoMongoDBCollTweets.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (t *TweetRepository) FindTweetbyId(idTweet primitive.ObjectID) (tweetdomain.Post, error) {

	GoMongoDBCollUsers := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	findTweet := bson.D{
		{Key: "_id", Value: idTweet},
	}
	var PostDocument tweetdomain.Post
	PostCollectionErr := GoMongoDBCollUsers.FindOne(context.TODO(), findTweet).Decode(&PostDocument)

	if PostCollectionErr != nil {
		return tweetdomain.Post{}, PostCollectionErr
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
func (t *TweetRepository) LikeTweet(TweetId, idValueToken primitive.ObjectID) error {
	GoMongoDBColl := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	count, err := GoMongoDBColl.CountDocuments(context.Background(), bson.D{{Key: "_id", Value: TweetId}})
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("el TweetId no existe")
	}
	filter := bson.D{{Key: "_id", Value: TweetId}}
	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "Likes", Value: idValueToken}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	GoMongoDBColl = t.mongoClient.Database("PINKKER-BACKEND").Collection("Users")

	filter = bson.D{{Key: "_id", Value: idValueToken}}
	update = bson.D{{Key: "$addToSet", Value: bson.D{{Key: "Likes", Value: TweetId}}}}

	_, err = GoMongoDBColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil

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
func (t *TweetRepository) GetPost(page int) ([]tweetdomain.TweetGetFollowReq, error) {

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
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
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
				{Key: "OriginalPost", Value: "$OriginalPost"},
				{Key: "UserInfo.FullName", Value: 1},
				{Key: "UserInfo.Avatar", Value: 1},
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
func (t *TweetRepository) GetPostId(id primitive.ObjectID) (tweetdomain.TweetGetFollowReq, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: id}}}}, // Filtrar por el _id proporcionado
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
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
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

	return tweetWithUserInfo, nil
}

func (t *TweetRepository) GetPostuser(page int, id primitive.ObjectID, limit int) ([]tweetdomain.TweetGetFollowReq, error) {

	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	skip := (page - 1) * 10
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "UserID", Value: id}}}},
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
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
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
				{Key: "OriginalPost", Value: "$OriginalPost"},
				{Key: "UserInfo.FullName", Value: 1},
				{Key: "UserInfo.Avatar", Value: 1},
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
func (t *TweetRepository) GetTweetsLast24HoursFollow(userIDs []primitive.ObjectID, page int) ([]tweetdomain.TweetGetFollowReq, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")
	currentTime := time.Now()
	last24Hours := currentTime.Add(-24 * time.Hour)
	skip := (page - 1) * 13
	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{
			{Key: "UserID", Value: bson.D{{Key: "$in", Value: userIDs}}},
			{Key: "TimeStamp", Value: bson.D{{Key: "$gte", Value: last24Hours}}},
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
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
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
				{Key: "OriginalPost", Value: "$OriginalPost"},
				{Key: "UserInfo.FullName", Value: 1},
				{Key: "UserInfo.Avatar", Value: 1},
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

func (t *TweetRepository) GetCommentPosts(tweetID primitive.ObjectID, page int) ([]tweetdomain.TweetCommentsGetReq, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	var tweet tweetdomain.Post
	if err := GoMongoDBCollTweets.FindOne(context.Background(), bson.M{"_id": tweetID}).Decode(&tweet); err != nil {
		return nil, err
	}

	skip := (page - 1) * 10

	commentIDs := tweet.Comments

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
			{Key: "Likes", Value: "$Likes"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
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

	return comments, nil
}
func (t *TweetRepository) RePost(rePost *tweetdomain.RePost) error {
	GoMongoDBColl := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

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

	_, err := GoMongoDBColl.InsertOne(context.Background(), rePost)
	return err
}
func (t *TweetRepository) GetTweetsRecommended(idT primitive.ObjectID, excludeIDs []primitive.ObjectID, limit int) ([]tweetdomain.TweetGetFollowReq, error) {
	ctx := context.Background()
	GoMongoDB := t.mongoClient.Database("PINKKER-BACKEND")
	GoMongoDBCollTweets := GoMongoDB.Collection("Post")
	GoMongoDBCollUsers := GoMongoDB.Collection("Users")

	// Obtener los IDs de los usuarios que sigue el usuario actual
	var followingUser userdomain.User
	err := GoMongoDBCollUsers.FindOne(ctx, bson.D{{Key: "_id", Value: idT}}).Decode(&followingUser)
	if err != nil {
		return nil, err
	}

	var followingIDs []primitive.ObjectID
	if len(followingUser.Following) == 0 {
		followingIDs = make([]primitive.ObjectID, 0)
	} else {
		for userID := range followingUser.Following {
			followingIDs = append(followingIDs, userID)
		}
	}

	last24Hours := time.Now().Add(-24 * time.Hour)

	pipeline := bson.A{
		// Etapa de coincidencia para encontrar los tweets que han sido dados like por los usuarios seguidos en las últimas 24 horas
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "Likes", Value: bson.D{{Key: "$in", Value: followingIDs}}},
			{Key: "TimeStamp", Value: bson.D{{Key: "$gte", Value: last24Hours}}},
		}}},
		// Etapa para agregar información del usuario
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		// Desenrollar la información del usuario
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		// Ordenar los tweets por timestamp
		bson.D{{Key: "$sort", Value: bson.D{{Key: "TimeStamp", Value: -1}}}},
		// Limitar la cantidad de tweets devueltos
		// Proyectar los campos deseados
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: 1},
			{Key: "PostImage", Value: 1},
			{Key: "Type", Value: 1},
			{Key: "TimeStamp", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Likes", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "RePosts", Value: 1},
			{Key: "OriginalPost", Value: 1},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
		}}},
		bson.D{{Key: "$limit", Value: limit}},
	}
	excludedIDs := make([]interface{}, len(excludeIDs))
	for i, id := range excludeIDs {
		excludedIDs[i] = id
	}
	excludeFilter := bson.D{{Key: "_id", Value: bson.D{{Key: "$nin", Value: excludedIDs}}}}
	pipeline = append(pipeline, bson.D{{Key: "$match", Value: excludeFilter}})

	cursor, err := GoMongoDBCollTweets.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Recopilar los tweets con información del usuario
	var tweetsWithUserInfo []tweetdomain.TweetGetFollowReq
	for cursor.Next(ctx) {
		var tweetWithUserInfo tweetdomain.TweetGetFollowReq
		if err := cursor.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}

		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	pipelineRandom := bson.A{
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "localField", Value: "UserID"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "UserInfo"},
		}}},
		// Desenrollar la información del usuario
		bson.D{{Key: "$unwind", Value: "$UserInfo"}},
		// Ordenar los tweets por timestamp
		bson.D{{Key: "$sort", Value: bson.D{{Key: "TimeStamp", Value: -1}}}},
		// Limitar la cantidad de tweets devueltos
		bson.D{{Key: "$limit", Value: 13}},
		// Proyectar los campos deseados
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "id", Value: "$_id"},
			{Key: "Status", Value: 1},
			{Key: "PostImage", Value: 1},
			{Key: "Type", Value: 1},
			{Key: "TimeStamp", Value: 1},
			{Key: "UserID", Value: 1},
			{Key: "Likes", Value: 1},
			{Key: "Comments", Value: 1},
			{Key: "RePosts", Value: 1},
			{Key: "OriginalPost", Value: 1},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "relevanceFactor", Value: -1}}}},
		bson.D{{Key: "$limit", Value: limit - len(tweetsWithUserInfo)}}, // Limitar la cantidad de clips devueltos por categorías distintas
	}
	pipelineRandom = append(pipelineRandom, bson.D{{Key: "$match", Value: excludeFilter}})

	cursorRandom, err := GoMongoDBCollTweets.Aggregate(ctx, pipelineRandom)
	if err != nil {
		return nil, err
	}
	defer cursorRandom.Close(ctx)
	for cursorRandom.Next(ctx) {
		var tweetWithUserInfo tweetdomain.TweetGetFollowReq
		if err := cursorRandom.Decode(&tweetWithUserInfo); err != nil {
			return nil, err
		}

		tweetsWithUserInfo = append(tweetsWithUserInfo, tweetWithUserInfo)
	}

	// Verificar si hay tweets originales para recuperar su información
	var originalPostIDs []primitive.ObjectID
	for _, tweet := range tweetsWithUserInfo {
		if tweet.OriginalPost != primitive.NilObjectID {
			originalPostIDs = append(originalPostIDs, tweet.OriginalPost)
		}
	}
	if len(originalPostIDs) > 0 {
		// Crear el pipeline de agregación para obtener información sobre los tweets originales
		originalPostPipeline := bson.A{
			// Etapa de coincidencia para encontrar los tweets originales por sus IDs
			bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: originalPostIDs}}}}}},
			// Etapa para agregar información del usuario
			bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "Users"},
				{Key: "localField", Value: "UserID"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "UserInfo"},
			}}},
			// Desenrollar la información del usuario
			bson.D{{Key: "$unwind", Value: "$UserInfo"}},
			// Proyectar los campos deseados
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
				{Key: "OriginalPost", Value: "$OriginalPost"},
				{Key: "UserInfo.FullName", Value: 1},
				{Key: "UserInfo.Avatar", Value: 1},
				{Key: "UserInfo.NameUser", Value: 1},
			}}},
		}

		// Ejecutar el pipeline de agregación para obtener información sobre los tweets originales
		cursorOriginalPosts, err := GoMongoDBCollTweets.Aggregate(ctx, originalPostPipeline)
		if err != nil {
			return nil, err
		}
		defer cursorOriginalPosts.Close(ctx)

		// Crear un mapa para almacenar los tweets originales por su ID
		var originalPostMap = make(map[primitive.ObjectID]tweetdomain.TweetGetFollowReq)
		for cursorOriginalPosts.Next(ctx) {
			var originalPost tweetdomain.TweetGetFollowReq
			if err := cursorOriginalPosts.Decode(&originalPost); err != nil {
				return nil, err
			}
			originalPostMap[originalPost.ID] = originalPost
		}

		// Asignar la información de los tweets originales a los tweets correspondientes en la lista
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
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

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
			{Key: "OriginalPost", Value: "$OriginalPost"},
			{Key: "UserInfo.FullName", Value: 1},
			{Key: "UserInfo.Avatar", Value: 1},
			{Key: "UserInfo.NameUser", Value: 1},
		}}},
		{{Key: "$skip", Value: int64((page - 1) * limit)}}, // Saltar resultados según la paginación
		{{Key: "$limit", Value: int64(limit)}},             // Limitar resultados según la paginación
	}

	cursor, err := GoMongoDBCollTweets.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var tweets []tweetdomain.TweetGetFollowReq
	if err := cursor.All(context.Background(), &tweets); err != nil {
		return nil, err
	}

	return tweets, nil
}

func (t *TweetRepository) GetTrendsByPrefix(prefix string, limit int) ([]tweetdomain.Trend, error) {
	GoMongoDB := t.mongoClient.Database("PINKKER-BACKEND").Collection("Trends")

	regex := primitive.Regex{Pattern: "^" + prefix, Options: "i"}

	pipeline := []bson.D{
		{{Key: "$match", Value: bson.D{{Key: "hashtag", Value: regex}}}},
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
