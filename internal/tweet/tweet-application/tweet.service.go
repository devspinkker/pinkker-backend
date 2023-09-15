package tweetapplication

import (
	tweetdomain "PINKKER-BACKEND/internal/tweet/tweet-domain"
	tweetinfrastructure "PINKKER-BACKEND/internal/tweet/tweet-infrastructure"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TweetService struct {
	TweetRepository *tweetinfrastructure.TweetRepository
}

func NewTweetService(TweetRepository *tweetinfrastructure.TweetRepository) *TweetService {
	return &TweetService{
		TweetRepository: TweetRepository,
	}
}

// save
func (ts *TweetService) SaveTweet(status string, img string, user primitive.ObjectID) error {
	var modelNewTweet tweetdomain.Tweet
	modelNewTweet.Status = status
	modelNewTweet.TweetImage = img
	modelNewTweet.UserID = user
	modelNewTweet.Likes = []primitive.ObjectID{}
	modelNewTweet.Comments = []primitive.ObjectID{}
	modelNewTweet.TimeStamp = time.Now()
	err := ts.TweetRepository.TweetSave(modelNewTweet)

	return err
}
func (ts *TweetService) SaveComment(status string, CommentBy primitive.ObjectID, img string, user primitive.ObjectID) error {
	var modelNewTweet tweetdomain.TweetComment
	modelNewTweet.Status = status
	modelNewTweet.TweetImage = img
	modelNewTweet.UserID = user
	modelNewTweet.Likes = []primitive.ObjectID{}
	modelNewTweet.TimeStamp = time.Now()
	modelNewTweet.Comments = []primitive.ObjectID{}
	modelNewTweet.CommentBy = CommentBy
	modelNewTweet.TimeStamp = time.Now()
	err := ts.TweetRepository.SaveComment(&modelNewTweet)

	return err
}

// like
func (ts *TweetService) LikeTweet(idTweet primitive.ObjectID, idValueToken primitive.ObjectID) error {
	err := ts.TweetRepository.LikeTweet(idTweet, idValueToken)
	return err
}

func (ts *TweetService) TweetDislike(idTweet primitive.ObjectID, idValueToken primitive.ObjectID) error {
	err := ts.TweetRepository.TweetDislike(idTweet, idValueToken)
	return err
}

// find
func (ts *TweetService) TweetGetFollow(idValueObj primitive.ObjectID) ([]tweetdomain.TweetGetFollowReq, error) {

	followedUsers, errGetFollowedUsers := ts.TweetRepository.GetFollowedUsers(idValueObj)
	if errGetFollowedUsers != nil {
		return nil, errGetFollowedUsers
	}

	Tweets, errGetTweetsLast24Hours := ts.TweetRepository.GetTweetsLast24Hours(followedUsers.Following)
	return Tweets, errGetTweetsLast24Hours
}

func (ts *TweetService) GetCommentPost(IdPost primitive.ObjectID) ([]tweetdomain.TweetCommentsGetReq, error) {

	followedUsers, errGetFollowedUsers := ts.TweetRepository.GetCommentPosts(IdPost)
	if errGetFollowedUsers != nil {
		return nil, errGetFollowedUsers
	}
	return followedUsers, nil
}
