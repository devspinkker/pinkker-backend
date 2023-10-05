package streaminterfaces

import (
	streamapplication "PINKKER-BACKEND/internal/stream/stream-application"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StreamHandler struct {
	StreamServise *streamapplication.StreamService
}

func NewStreamService(StreamServise *streamapplication.StreamService) *StreamHandler {
	return &StreamHandler{
		StreamServise: StreamServise,
	}
}

type IDStream struct {
	IdStream primitive.ObjectID `json:"IdStream"`
}

// get stream by id
func (s *StreamHandler) GetStreamById(c *fiber.Ctx) error {
	var idStream IDStream
	if err := c.BodyParser(&idStream); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	stream, err := s.StreamServise.GetStreamById(idStream.IdStream)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Stream not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    stream,
	})
}

type Streamer struct {
	Streamer string `json:"Streamer"`
}

// get stream by name user
func (s *StreamHandler) GetStreamByNameUser(c *fiber.Ctx) error {
	var StreamerReq Streamer
	if err := c.BodyParser(&StreamerReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	stream, err := s.StreamServise.GetStreamByNameUser(StreamerReq.Streamer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Stream not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    stream,
	})
}

type Categorie struct {
	Categorie string `json:"Categorie"`
}

// get streams by caregories
func (s *StreamHandler) GetStreamsByCategorie(c *fiber.Ctx) error {
	var CategorierReq Categorie
	if err := c.BodyParser(&CategorierReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	stream, err := s.StreamServise.GetStreamsByCategorie(CategorierReq.Categorie, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    stream,
	})
}

type StremesIFollow struct {
	FollowingIds []primitive.ObjectID `json:"FollowingIds"`
}

func (s *StreamHandler) GetStreamsIdsStreamer(c *fiber.Ctx) error {
	var StremesIFollowReq StremesIFollow
	if err := c.BodyParser(&StremesIFollowReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	streams, errGetStreams := s.StreamServise.GetStreamsIdsStreamer(StremesIFollowReq.FollowingIds)
	if errGetStreams != nil {
		if errGetStreams.Error() == "no se encontraron streams que esten Online" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": errGetStreams.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": errGetStreams.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    streams,
	})
}

type ReqUpdate_online struct {
	State bool   `json:"State"`
	Key   string `json:"Key"`
}

func (s *StreamHandler) Update_online(c *fiber.Ctx) error {

	var req ReqUpdate_online
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if err := s.StreamServise.Update_online(req.Key, req.State); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (s *StreamHandler) CloseStream(c *fiber.Ctx) error {

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if err := s.StreamServise.CloseStream(IdUserTokenP); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

type Update_thumbnail struct {
	Image string `json:"image"`
}

func (s *StreamHandler) Update_thumbnail(c *fiber.Ctx) error {

	var requestBody Update_thumbnail
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if err := s.StreamServise.Update_thumbnail(IdUserTokenP, requestBody.Image); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (s *StreamHandler) Streamings_online(c *fiber.Ctx) error {

	online, err := s.StreamServise.Streamings_online()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    online,
	})
}

type Update_start_date struct {
	Date int `json:"date"`
}

func (s *StreamHandler) Update_start_date(c *fiber.Ctx) error {

	var requestBody Update_start_date
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)

	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if err := s.StreamServise.Update_start_date(IdUserTokenP, requestBody.Date); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (s *StreamHandler) UpdateStreamInfo(c *fiber.Ctx) error {

	var requestBody streamdomain.UpdateStreamInfo
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	if err := requestBody.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if err := s.StreamServise.UpdateStreamInfo(IdUserTokenP, requestBody); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (s *StreamHandler) Update_Emotes(c *fiber.Ctx) error {

	var requestBody Update_start_date // #############
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)

	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if err := s.StreamServise.Update_Emotes(IdUserTokenP, requestBody.Date); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
