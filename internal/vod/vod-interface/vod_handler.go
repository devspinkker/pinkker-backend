package vodinterfaces

import (
	vodsapplication "PINKKER-BACKEND/internal/vod/vod-application"

	"github.com/gofiber/fiber/v2"
)

type VodHandler struct {
	VodServise *vodsapplication.VodService
}

func NewVodService(VodServise *vodsapplication.VodService) *VodHandler {
	return &VodHandler{
		VodServise: VodServise,
	}
}

func (v *VodHandler) GetVodtreamer(c *fiber.Ctx) error {
	streamer := c.Query("streamer")
	limit := c.Query("limit")
	sort := c.Query("sort")
	vod, err := v.VodServise.GetVodByStreamer(streamer, limit, sort)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"messages": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"messages": "GetVodtreamer",
		"data":     vod,
	})
}
func (v *VodHandler) GetVodWithId(c *fiber.Ctx) error {
	vodId := c.Query("vodId")

	vod, err := v.VodServise.GetVodWithId(vodId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"messages": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"messages": "GetVodtreamer",
		"data":     vod,
	})
}
func (v *VodHandler) CreateVod(c *fiber.Ctx) error {
	var reqBody struct {
		URL       string `json:"url"`
		StreamKey string `json:"stream_key"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"messages": "Invalid request body",
		})
	}

	err := v.VodServise.CreateVod(reqBody.URL, reqBody.StreamKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"messages": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"messages": "Created Vod success!",
	})
}
