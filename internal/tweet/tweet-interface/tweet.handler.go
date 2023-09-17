package tweetinterfaces

import (
	tweetapplication "PINKKER-BACKEND/internal/tweet/tweet-application"
	tweetdomain "PINKKER-BACKEND/internal/tweet/tweet-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"fmt"

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
	go helpers.Processimage(fileHeader, PostImageChanel, errChanel)

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

	for {
		select {
		case PostImage := <-PostImageChanel:
			err := th.TweetServise.SaveTweet(newTweet.Status, PostImage, idValueObj)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": err.Error(),
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
	Tweets, errTweetGetFollow := th.TweetServise.TweetGetFollow(idValueObj)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": Tweets,
	})

}

func (th *TweetHandler) CommentPost(c *fiber.Ctx) error {

	fileHeader, _ := c.FormFile("imgPost")
	PostImageChanel := make(chan string)
	errChanel := make(chan error)
	go helpers.Processimage(fileHeader, PostImageChanel, errChanel)

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

	for {
		select {
		case PostImage := <-PostImageChanel:
			err := th.TweetServise.SaveComment(TweetComment.Status, TweetComment.CommentTo, PostImage, idValueObj)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": err.Error(),
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
func (th *TweetHandler) TweetGetCommentPostGetFollow(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	Tweets, errTweetGetFollow := th.TweetServise.TweetGetFollow(idValueObj)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": Tweets,
	})

}

type reqCommentPost struct {
	IdPost primitive.ObjectID `json:"IdPost"`
}

func (th *TweetHandler) GetCommentPost(c *fiber.Ctx) error {
	var req reqCommentPost
	err := c.BodyParser(&req)
	if err != nil {
		fmt.Println("asa")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	Tweets, errTweetGetFollow := th.TweetServise.GetCommentPost(req.IdPost)
	if errTweetGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errTweetGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": Tweets,
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
	errRePost := th.TweetServise.CitaPost(idValueObj, req.OriginalPost, req.Status)
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
