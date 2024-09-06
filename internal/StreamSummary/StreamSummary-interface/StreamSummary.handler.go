package StreamSummaryinterfaces

import (
	StreamSummaryapplication "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-application"
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-domain"
	"strconv"
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
func (h *StreamSummaryHandler) DeleteStreamSummaryByIDAndStreamerID(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    errorID.Error(),
		})
	}

	Idvod := c.Query("Idvod", "")
	if Idvod == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "IdClip parameter is required",
		})
	}

	IdvodPr, err := primitive.ObjectIDFromHex(Idvod)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid IdClip parameter",
		})
	}

	err = h.StreamSummaryServise.DeleteStreamSummaryByIDAndStreamerID(IdvodPr, idValueToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (h *StreamSummaryHandler) UpdateStreamSummaryByIDAndStreamerID(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    errorID.Error(),
		})
	}

	title := c.Query("title", "")
	if title == "" || len(title) > 100 || len(title) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "title bad request",
		})
	}

	Idvodstr := c.Query("Idvod", "")
	if Idvodstr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "IdClip parameter is required",
		})
	}

	IdvodPr, err := primitive.ObjectIDFromHex(Idvodstr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid IdClip parameter",
		})
	}

	err = h.StreamSummaryServise.UpdateStreamSummaryByIDAndStreamerID(IdvodPr, idValueToken, title)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (h *StreamSummaryHandler) GetEarningsByRange(c *fiber.Ctx) error {
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	idValue := c.Context().UserValue("_id").(string)

	streamerID, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start date"})
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid end date"})
	}

	earnings, err := h.StreamSummaryServise.GetEarningsByRange(streamerID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get earnings"})
	}

	return c.JSON(earnings)
}

func (h *StreamSummaryHandler) GetEarningsByDay(c *fiber.Ctx) error {
	dayStr := c.Query("day")
	idValue := c.Context().UserValue("_id").(string)

	streamerID, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	// Parse the day from the query parameter
	day, err := time.Parse("2006-01-02", dayStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid day format, use YYYY-MM-DD",
		})
	}

	// Get earnings by day from the service
	earnings, err := h.StreamSummaryServise.GetEarningsByDay(streamerID, day)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get earnings by day",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    earnings,
	})
}

// GetEarningsByWeek handles the request to get earnings for a specific week
func (h *StreamSummaryHandler) GetEarningsByWeek(c *fiber.Ctx) error {
	weekStr := c.Query("week")
	idValue := c.Context().UserValue("_id").(string)
	streamerID, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	// Parse the week start date from the query parameter
	week, err := time.Parse("2006-01-02", weekStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid week format, use YYYY-MM-DD",
		})
	}

	// Get earnings by week from the service
	earnings, err := h.StreamSummaryServise.GetEarningsByWeek(streamerID, week)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get earnings by week",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    earnings,
	})
}

// GetEarningsByMonth handles the request to get earnings for a specific month
func (h *StreamSummaryHandler) GetEarningsByMonth(c *fiber.Ctx) error {
	monthStr := c.Query("month")

	idValue := c.Context().UserValue("_id").(string)

	streamerID, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	// Parse the month from the query parameter
	month, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid month format, use YYYY-MM",
		})
	}

	// Get earnings by month from the service
	earnings, err := h.StreamSummaryServise.GetEarningsByMonth(streamerID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get earnings by month",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    earnings,
	})
}
func (h *StreamSummaryHandler) GetDailyEarningsForMonth(c *fiber.Ctx) error {
	monthStr := c.Query("month")

	idValue := c.Context().UserValue("_id").(string)

	streamerID, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	month, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid month format, use YYYY-MM",
		})
	}

	// Get earnings by month from the service
	earnings, err := h.StreamSummaryServise.GetDailyEarningsForMonth(streamerID, month)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get earnings by month",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    earnings,
	})
}

// GetEarningsByYear handles the request to get earnings for a specific year
func (h *StreamSummaryHandler) GetEarningsByYear(c *fiber.Ctx) error {
	yearStr := c.Query("year")

	idValue := c.Context().UserValue("_id").(string)

	streamerID, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	// Parse the year from the query parameter
	year, err := time.Parse("2006", yearStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid year format, use YYYY",
		})
	}

	// Get earnings by year from the service
	earnings, err := h.StreamSummaryServise.GetEarningsByYear(streamerID, year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get earnings by year",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    earnings,
	})

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
func (s *StreamSummaryHandler) AWeekOfStreaming(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, err := primitive.ObjectIDFromHex(idValue)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user ID",
		})
	}

	page, err := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid page number",
		})
	}

	LastWeekStreamSummaries, err := s.StreamSummaryServise.AWeekOfStreaming(idValueObj, int(page))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    LastWeekStreamSummaries,
	})
}
