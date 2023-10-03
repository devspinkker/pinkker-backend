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
func (t *TweetRepository) GetLatestPosts() ([]tweetdomain.TweetGetFollowReq, error) {

	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

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

func (t *TweetRepository) GetCommentPosts(tweetID primitive.ObjectID) ([]tweetdomain.TweetCommentsGetReq, error) {
	GoMongoDBCollTweets := t.mongoClient.Database("PINKKER-BACKEND").Collection("Post")

	var tweet tweetdomain.Post
	if err := GoMongoDBCollTweets.FindOne(context.Background(), bson.M{"_id": tweetID}).Decode(&tweet); err != nil {
		return nil, err
	}

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
			{Key: "UserInfo.FullName", Value: "$UserInfo.FullName"},
			{Key: "UserInfo.Avatar", Value: "$UserInfo.Avatar"},
			{Key: "UserInfo.NameUser", Value: "$UserInfo.NameUser"},
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
