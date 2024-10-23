package tweetinterfaces

import (
	tweetapplication "PINKKER-BACKEND/internal/tweet/tweet-application"
	tweetdomain "PINKKER-BACKEND/internal/tweet/tweet-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TweetHandler struct {
	TweetServise *tweetapplication.TweetService
}

func NewTweetService(TweetServise *tweetapplication.TweetService) *TweetHandler {
	return &TweetHandler{
		TweetServise: TweetServise,
	}
}

// Create
func (th *TweetHandler) CreatePost(c *fiber.Ctx) error {

	fileHeader, _ := c.FormFile("imgPost")

	PostImageChanel := make(chan string)
	errChanel := make(chan error)
	var newTweet tweetdomain.TweetModelValidator
	if err := c.BodyParser(&newTweet); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"messages": "Bad Request",
		})
	}
	if err := newTweet.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   err.Error(),
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var IsPrivate bool = false
	if newTweet.CommunityID != primitive.NilObjectID {
		member, Private, _ := th.TweetServise.IsUserMemberOfCommunity(newTweet.CommunityID, idValueObj)

		if !member {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "no member",
			})
		}
		IsPrivate = Private
	}
	go helpers.Processimage(fileHeader, PostImageChanel, errChanel)

	for {
		select {
		case PostImage := <-PostImageChanel:
			idTweet, err := th.TweetServise.SaveTweet(newTweet.Status, newTweet.CommunityID, PostImage, idValueObj, IsPrivate)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": err.Error(),
				})
			}
			Tweet, errTweetGetFollow := th.TweetServise.GetPostId(idTweet)
			if errTweetGetFollow != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "StatusInternalServerError",
					"data":    errTweetGetFollow.Error(),
				})
			}
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"message": "StatusCreated",
				"post":    Tweet,
			})
		case <-errChanel:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "avatarUrl error",
			})
		}
	}
}

type IDTweet struct {
	IDTweet string `json:"idPost"`
}

// like
func (th *TweetHandler) PostLike(c *fiber.Ctx) error {
	var idTweetReq IDTweet
	if err := c.BodyParser(&idTweetReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	IdTweet, err := primitive.ObjectIDFromHex(idTweetReq.IDTweet)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	errLike := th.TweetServise.LikeTweet(IdTweet, idValueToken)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Like",
	})
}
func (th *TweetHandler) PostDislike(c *fiber.Ctx) error {
	var idTweetReq IDTweet
	if err := c.BodyParser(&idTweetReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	IdTweet, err := primitive.ObjectIDFromHex(idTweetReq.IDTweet)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	errLike := th.TweetServise.TweetDislike(IdTweet, idValueToken)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "TweetDislike",
	})
}

// find
func (th *TweetHandler) TweetGetFollow(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	Tweets, errTweetGetFollow := th.TweetServise.TweetGetFollow(idValueObj, page)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": Tweets,
	})
}

func (th *TweetHandler) PostGets(c *fiber.Ctx) error {

	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	Tweets, errTweetGetFollow := th.TweetServise.GetPost(page)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweets,
	})
}
func (th *TweetHandler) GetPostId(c *fiber.Ctx) error {

	idStr := c.Query("id", "")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    "name error",
		})
	}

	Tweet, errTweetGetFollow := th.TweetServise.GetPostId(id)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweet,
	})
}
func (th *TweetHandler) GetPostIdLogueado(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	idStr := c.Query("id", "")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    "name error",
		})
	}

	Tweet, errTweetGetFollow := th.TweetServise.GetPostIdLogueado(id, idValueObj)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweet,
	})
}

func (th *TweetHandler) GetTweetsRecommended(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req tweetdomain.GetRecommended
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	Tweets, errTweetGetFollow := th.TweetServise.GetTweetsRecommended(idValueObj, req.ExcludeIDs)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}

	PostAds, err := th.TweetServise.GetAdsMuroAndPost()
	if err != nil {
		fmt.Println(err)
	}
	if PostAds.ReferenceLink != "" {
		var combinedData []interface{}

		for _, tweet := range Tweets {
			combinedData = append(combinedData, tweet)
		}

		combinedData = append(combinedData, PostAds)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "ok",
			"data":    combinedData,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweets,
	})
}

func (th *TweetHandler) GetRandomPostcommunities(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req tweetdomain.GetRecommended
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	Tweets, errTweetGetFollow := th.TweetServise.GetRandomPostcommunities(idValueObj, req.ExcludeIDs)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}

	PostAds, err := th.TweetServise.GetAdsMuroAndPost()
	if err != nil {
		fmt.Println(err)
	}
	if PostAds.ReferenceLink != "" {
		var combinedData []interface{}

		for _, tweet := range Tweets {
			combinedData = append(combinedData, tweet)
		}

		combinedData = append(combinedData, PostAds)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "ok",
			"data":    combinedData,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweets,
	})
}
func (th *TweetHandler) GetPostCommunitiesFromUser(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req tweetdomain.GetRecommended
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	Tweets, errTweetGetFollow := th.TweetServise.GetPostCommunitiesFromUser(idValueObj, req.ExcludeIDs)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}

	PostAds, err := th.TweetServise.GetAdsMuroAndPost()
	if err != nil {
		fmt.Println(err)
	}
	if PostAds.ReferenceLink != "" {
		var combinedData []interface{}

		for _, tweet := range Tweets {
			combinedData = append(combinedData, tweet)
		}

		combinedData = append(combinedData, PostAds)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "ok",
			"data":    combinedData,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweets,
	})
}
func (th *TweetHandler) GetPostuser(c *fiber.Ctx) error {

	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	idStr := c.Query("id", "")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    "name error",
		})
	}

	Tweets, errTweetGetFollow := th.TweetServise.GetPostuser(page, id)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweets,
	})
}

func (th *TweetHandler) GetPostsWithImages(c *fiber.Ctx) error {

	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	idStr := c.Query("id", "")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    "name error",
		})
	}

	Tweets, errTweetGetFollow := th.TweetServise.GetPostsWithImages(page, id)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweets,
	})
}

func (th *TweetHandler) GetPostuserLogueado(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	idStr := c.Query("id", "")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    "name error",
		})
	}

	Tweets, errTweetGetFollow := th.TweetServise.GetPostuserLogueado(page, id, idValueObj)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweets,
	})
}
func (th *TweetHandler) CommentPost(c *fiber.Ctx) error {

	fileHeader, _ := c.FormFile("imgPost")
	PostImageChanel := make(chan string)
	errChanel := make(chan error)
	var TweetComment tweetdomain.TweetCommentModelValidator
	if err := c.BodyParser(&TweetComment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"messages": "Bad Request",
		})
	}
	if err := TweetComment.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   err.Error(),
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var IsPrivate bool = false
	if TweetComment.CommunityID != primitive.NilObjectID {
		member, private, _ := th.TweetServise.IsUserMemberOfCommunity(TweetComment.CommunityID, idValueObj)

		if !member {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "no member",
			})
		}
		IsPrivate = private
	}
	go helpers.Processimage(fileHeader, PostImageChanel, errChanel)

	for {
		select {
		case PostImage := <-PostImageChanel:
			insertedID, err := th.TweetServise.SaveComment(TweetComment.Status, TweetComment.OriginalPost, PostImage, idValueObj, IsPrivate)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": err.Error(),
				})
			}
			Tweet, errTweetGetFollow := th.TweetServise.GetPostId(insertedID)
			if errTweetGetFollow != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "StatusInternalServerError",
					"data":    errTweetGetFollow.Error(),
				})
			}
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"message": "StatusCreated",
				"post":    Tweet,
			})
		case <-errChanel:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "avatarUrl error",
			})
		}
	}
}

// func (th *TweetHandler) TweetGetCommentPostGetFollow(c *fiber.Ctx) error {
// 	idValue := c.Context().UserValue("_id").(string)
// 	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
// 	if errorID != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"message": "StatusBadRequest",
// 		})
// 	}
// 	Tweets, errTweetGetFollow := th.TweetServise.TweetGetFollow(idValueObj)
// 	if errTweetGetFollow != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"message": "StatusBadRequest",
// 		})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"message": Tweets,
// 	})
// }

func (th *TweetHandler) GetCommentPost(c *fiber.Ctx) error {
	idStr := c.Query("id", "")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    "name error",
		})
	}
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	GetCommentPost, errTweetGetFollow := th.TweetServise.GetCommentPost(id, page, idValueObj)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    GetCommentPost,
	})

}

type RePostIdPost struct {
	IDPost primitive.ObjectID `json:"idPost"`
}

func (th *TweetHandler) RePost(c *fiber.Ctx) error {
	var req RePostIdPost
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	errRePost := th.TweetServise.RePost(idValueObj, req.IDPost)
	if errRePost != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errRePost.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (th *TweetHandler) CitaPost(c *fiber.Ctx) error {
	fileHeader, _ := c.FormFile("imgPost")
	PostImageChanel := make(chan string)
	errChanel := make(chan error)
	go helpers.Processimage(fileHeader, PostImageChanel, errChanel)

	var req tweetdomain.CitaPostModelValidator
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err,
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	for {
		select {
		case PostImage := <-PostImageChanel:
			errRePost := th.TweetServise.CitaPost(idValueObj, req.OriginalPost, req.Status, PostImage)
			if errRePost != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": errRePost.Error(),
				})
			}
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"message": "StatusCreated",
			})
		case <-errChanel:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "avatarUrl error",
			})
		}
	}

}

func (th *TweetHandler) GetTrends(c *fiber.Ctx) error {

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	Tweet, errTweetGetFollow := th.TweetServise.GetTrends(page)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweet,
	})
}

func (th *TweetHandler) GetTweetsByHashtag(c *fiber.Ctx) error {

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	hashtag := c.Query("hashtag", "")

	Tweet, errTweetGetFollow := th.TweetServise.GetTweetsByHashtag(page, hashtag)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Tweet,
	})
}

func (th *TweetHandler) GetTrendsByPrefix(c *fiber.Ctx) error {
	hashtag := c.Query("hashtag", "")

	Trend, err := th.TweetServise.GetTrendsByPrefix(hashtag)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Trend,
	})
}

func (h *TweetHandler) GetCommunityPosts(c *fiber.Ctx) error {
	var req struct {
		CommunityIDs     primitive.ObjectID   `json:"community_ids"`
		ExcludeFilterIDs []primitive.ObjectID `json:"ExcludeFilterIDs"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing request",
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	// Llamar al servicio para obtener los posts de las comunidades
	posts, err := h.TweetServise.GetCommunityPosts(c.Context(), req.CommunityIDs, req.ExcludeFilterIDs, idValueToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching community posts",
			"error":   err.Error(),
		})
	}
	PostAds, err := h.TweetServise.GetAdsMuroByCommunityId(req.CommunityIDs)
	if err != nil {
		fmt.Println(err)
	}
	if PostAds.ReferenceLink != "" {
		var combinedData []interface{}

		for _, tweet := range posts {
			combinedData = append(combinedData, tweet)
		}

		combinedData = append(combinedData, PostAds)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "ok",
			"posts":   combinedData,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Posts fetched successfully",
		"posts":   posts,
	})
}
