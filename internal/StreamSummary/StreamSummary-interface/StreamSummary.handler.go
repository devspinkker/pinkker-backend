package StreamSummaryinterfaces

import (
	StreamSummaryapplication "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-application"
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-domain"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StreamSummaryHandler struct {
	StreamSummaryServise *StreamSummaryapplication.StreamSummaryService
}

func NewStreamSummaryService(StreamSummaryServise *StreamSummaryapplication.StreamSummaryService) *StreamSummaryHandler {
	return &StreamSummaryHandler{
		StreamSummaryServise: StreamSummaryServise,
	}
}

func (s *StreamSummaryHandler) GeStreamSummaries(c *fiber.Ctx) error {
	type ReqGetUserByNameUser struct {
		ID primitive.ObjectID `json:"id" query:"id"`
	}

	var Req ReqGetUserByNameUser
	if err := c.QueryParser(&Req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	StreamSummaries, errGetUserBykey := s.StreamSummaryServise.GeStreamSummaries(Req.ID)
	if errGetUserBykey != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    StreamSummaries,
	})
}
func (s *StreamSummaryHandler) GetStreamSummaryByTitle(c *fiber.Ctx) error {
	type ReqGetUserByNameUser struct {
		Title string `json:"title" query:"title"`
	}

	var Req ReqGetUserByNameUser
	if err := c.QueryParser(&Req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	StreamSummaries, errGetUserBykey := s.StreamSummaryServise.GetStreamSummaryByTitle(Req.Title)
	if errGetUserBykey != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    StreamSummaries,
	})
}
func (s *StreamSummaryHandler) GetTopVodsLast48Hours(c *fiber.Ctx) error {

	StreamSummaries, errGetUserBykey := s.StreamSummaryServise.GetTopVodsLast48Hours()
	if errGetUserBykey != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    StreamSummaries,
	})
}

func (s *StreamSummaryHandler) GetStreamSummariesByStreamerIDLast30Days(c *fiber.Ctx) error {
	type ReqGetUserByNameUser struct {
		Streamer primitive.ObjectID `json:"Streamer" query:"Streamer"`
	}

	var Req ReqGetUserByNameUser
	if err := c.QueryParser(&Req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	StreamSummaries, errGetUserBykey := s.StreamSummaryServise.GetStreamSummariesByStreamerIDLast30Days(Req.Streamer)
	if errGetUserBykey != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    StreamSummaries,
	})
}

func (s *StreamSummaryHandler) UpdateStreamSummary(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req StreamSummarydomain.UpdateStreamSummary

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	err := s.StreamSummaryServise.UpdateStreamSummary(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (s *StreamSummaryHandler) AddAds(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req StreamSummarydomain.AddAds

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	err := s.StreamSummaryServise.AddAds(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (s *StreamSummaryHandler) AverageViewers(c *fiber.Ctx) error {
	var req StreamSummarydomain.AverageViewers

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	err := s.StreamSummaryServise.AverageViewers(req.StreamerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (s *StreamSummaryHandler) GetLastSixStreamSummaries(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, err := primitive.ObjectIDFromHex(idValue)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	type request struct {
		Date time.Time `json:"date"`
	}

	var requestBody request
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	LastSixStreamSummaries, err := s.StreamSummaryServise.GetLastSixStreamSummaries(idValueObj, requestBody.Date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    LastSixStreamSummaries,
	})
}
