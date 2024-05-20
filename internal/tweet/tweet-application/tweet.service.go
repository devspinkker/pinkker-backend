package tweetapplication

import (
	tweetdomain "PINKKER-BACKEND/internal/tweet/tweet-domain"
	tweetinfrastructure "PINKKER-BACKEND/internal/tweet/tweet-infrastructure"
	"strings"
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
func (ts *TweetService) SaveTweet(status string, img string, user primitive.ObjectID) (primitive.ObjectID, error) {
	var modelNewTweet tweetdomain.Post
	modelNewTweet.Status = status
	modelNewTweet.PostImage = img
	modelNewTweet.UserID = user
	modelNewTweet.Likes = []primitive.ObjectID{}
	modelNewTweet.Comments = []primitive.ObjectID{}
	modelNewTweet.TimeStamp = time.Now()
	modelNewTweet.RePosts = []primitive.ObjectID{}
	modelNewTweet.Type = "Post"
	Hashtags := extractHashtags(status)
	modelNewTweet.Hashtags = Hashtags

	idTweet, err := ts.TweetRepository.TweetSave(modelNewTweet)
	if err != nil {
		return idTweet, err

	}
	if len(Hashtags) > 0 {
		err = ts.TweetRepository.UpdateTrends(Hashtags)
	}
	return idTweet, err
}

func extractHashtags(status string) []string {
	hashtags := []string{}
	words := strings.Fields(status)
	for _, word := range words {
		if strings.HasPrefix(word, "#") {
			hashtag := strings.ToLower(strings.Trim(word, "#"))
			hashtags = append(hashtags, hashtag)
		}
	}
	return hashtags
}
func (ts *TweetService) SaveComment(status string, CommentTo primitive.ObjectID, img string, user primitive.ObjectID) error {
	var modelNewTweet tweetdomain.PostComment
	modelNewTweet.Status = status
	modelNewTweet.PostImage = img
	modelNewTweet.UserID = user
	modelNewTweet.Likes = []primitive.ObjectID{}
	modelNewTweet.TimeStamp = time.Now()
	modelNewTweet.Comments = []primitive.ObjectID{}
	modelNewTweet.OriginalPost = CommentTo
	modelNewTweet.TimeStamp = time.Now()
	modelNewTweet.RePosts = []primitive.ObjectID{}
	modelNewTweet.Type = "PostComment"

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
func (ts *TweetService) TweetGetFollow(idValueObj primitive.ObjectID, page int) ([]tweetdomain.TweetGetFollowReq, error) {
	// Obtener los usuarios seguidos por el usuario específico
	followedUsers, errGetFollowedUsers := ts.TweetRepository.GetFollowedUsers(idValueObj)
	if errGetFollowedUsers != nil {
		return nil, errGetFollowedUsers
	}

	var followingIDs []primitive.ObjectID

	for ids := range followedUsers.Following {
		followingIDs = append(followingIDs, ids)
	}

	Tweets, errGetTweetsLast24Hours := ts.TweetRepository.GetTweetsLast24HoursFollow(followingIDs, page)
	return Tweets, errGetTweetsLast24Hours
}
func (ts *TweetService) GetTweetsRecommended(idT primitive.ObjectID, excludeIDs []primitive.ObjectID) ([]tweetdomain.TweetGetFollowReq, error) {
	limit := 15
	Tweets, err := ts.TweetRepository.GetTweetsRecommended(idT, excludeIDs, limit)
	return Tweets, err
}
func (ts *TweetService) GetPost(page int) ([]tweetdomain.TweetGetFollowReq, error) {

	Tweets, errGetFollowedUsers := ts.TweetRepository.GetPost(page)
	return Tweets, errGetFollowedUsers
}
func (ts *TweetService) GetPostId(id primitive.ObjectID) (tweetdomain.TweetGetFollowReq, error) {

	Tweet, errGetFollowedUsers := ts.TweetRepository.GetPostId(id)
	return Tweet, errGetFollowedUsers
}

func (ts *TweetService) GetPostuser(page int, id primitive.ObjectID, limit int) ([]tweetdomain.TweetGetFollowReq, error) {

	Tweets, errGetFollowedUsers := ts.TweetRepository.GetPostuser(page, id, limit)
	return Tweets, errGetFollowedUsers
}

func (ts *TweetService) GetCommentPost(IdPost primitive.ObjectID, page int) ([]tweetdomain.TweetCommentsGetReq, error) {

	comments, err := ts.TweetRepository.GetCommentPosts(IdPost, page)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (ts *TweetService) RePost(userid primitive.ObjectID, IdPost primitive.ObjectID) error {

	var Repost tweetdomain.RePost
	Repost.UserID = userid
	Repost.OriginalPost = IdPost
	Repost.TimeStamp = time.Now()
	Repost.Type = "RePost"

	err := ts.TweetRepository.RePost(&Repost)
	return err
}
func (ts *TweetService) CitaPost(userid primitive.ObjectID, IdPost primitive.ObjectID, status string, image string) error {

	var CitaPost tweetdomain.CitaPost
	CitaPost.UserID = userid
	CitaPost.PostImage = image
	CitaPost.OriginalPost = IdPost
	CitaPost.Status = status
	CitaPost.TimeStamp = time.Now()
	CitaPost.Likes = []primitive.ObjectID{}
	CitaPost.RePosts = []primitive.ObjectID{}
	CitaPost.Comments = []primitive.ObjectID{}
	CitaPost.Type = "CitaPost"

	err := ts.TweetRepository.CitaPost(&CitaPost)
	return err
}
func (t *TweetService) GetTrends(page int) ([]tweetdomain.Trend, error) {
	Trend, err := t.TweetRepository.GetTrends(page, 10)
	return Trend, err
}
func (t *TweetService) GetTweetsByHashtag(page int, hashtag string) ([]tweetdomain.TweetGetFollowReq, error) {
	post, err := t.TweetRepository.GetTweetsByHashtag(hashtag, page, 10)
	return post, err

}

func (t *TweetService) GetTrendsByPrefix(prefix string) ([]tweetdomain.Trend, error) {
	post, err := t.TweetRepository.GetTrendsByPrefix(prefix, 10)
	return post, err

}
