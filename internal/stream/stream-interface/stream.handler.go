package streaminterfaces

import (
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	streamapplication "PINKKER-BACKEND/internal/stream/stream-application"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"fmt"
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
	IdStream primitive.ObjectID `json:"IdStream" `
}

func (s *StreamHandler) CategoriesUpdate(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)
	idUser, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)

	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var requestBody streamdomain.CategoriesUpdate
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	fmt.Println(requestBody)
	fileHeader, err := c.FormFile("avatar")
	if err == nil {
		PostImageChanel := make(chan string)
		errChanel := make(chan error)
		go helpers.Processimage(fileHeader, PostImageChanel, errChanel)

		select {
		case avatarUrl := <-PostImageChanel:
			requestBody.Img = avatarUrl
		case <-errChanel:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error processing image",
			})
		}
	}

	// Actualizar la categoría
	if err := s.StreamServise.CategoriesUpdate(requestBody, idUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (s *StreamHandler) CommercialInStream(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)
	idUser, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)

	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var requestBody streamdomain.CommercialInStream
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

	Stream, err := s.StreamServise.GetStreamById(idUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	Commercial, err := s.StreamServise.CommercialInStreamSelectAdvertisements(Stream.StreamCategory)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "StatusNotFound",
			"data":    err.Error(),
		})
	}
	err = s.NotifyCommercialInStreamToRoomClients(Stream.ID.Hex(), Commercial)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (s *StreamHandler) NotifyCommercialInStreamToRoomClients(roomID string, Commercial advertisements.Advertisements) error {
	clients, err := s.StreamServise.GetWebSocketClientsInRoom(roomID)
	if err != nil {
		return err
	}

	notification := map[string]interface{}{
		"_id":           Commercial.ID,
		"UrlVideo":      Commercial.UrlVideo,
		"LinkReference": Commercial.ReferenceLink,
	}
	for _, client := range clients {
		err = client.WriteJSON(notification)
		if err != nil {
			return err
		}
	}

	return nil
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
	Streamer string `json:"Streamer" query:"Streamer"`
}

// get stream by name user
func (s *StreamHandler) GetStreamByNameUser(c *fiber.Ctx) error {
	var StreamerReq Streamer
	if err := c.QueryParser(&StreamerReq); err != nil {
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
	if err := c.QueryParser(&CategorierReq); err != nil {
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

func (s *StreamHandler) GetAllsStreamsOnline(c *fiber.Ctx) error {

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	stream, err := s.StreamServise.GetAllsStreamsOnline(page)
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
func (s *StreamHandler) GetStreamsMostViewed(c *fiber.Ctx) error {

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	stream, err := s.StreamServise.GetStreamsMostViewed(page)
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

func (s *StreamHandler) GetAllsStreamsOnlineThatUserFollows(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	stream, err := s.StreamServise.GetAllsStreamsOnlineThatUserFollows(idValueObj)
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
	LastStreamSummary, err := s.StreamServise.Update_online(req.Key, req.State)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    LastStreamSummary,
	})
}

type ReqUpdate_CloseStream struct {
	Key string `json:"keyTransmission"`
}

func (s *StreamHandler) CloseStream(c *fiber.Ctx) error {

	var req ReqUpdate_CloseStream
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if err := s.StreamServise.CloseStream(req.Key); err != nil {
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
	Cmt   string `json:"cmt"`
}

func (s *StreamHandler) Update_thumbnail(c *fiber.Ctx) error {

	var requestBody Update_thumbnail
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	if err := s.StreamServise.Update_thumbnail(requestBody.Cmt, requestBody.Image); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
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
	Date int    `json:"date"`
	Key  string `json:"keyTransmission"`
}

func (s *StreamHandler) Update_start_date(c *fiber.Ctx) error {

	var requestBody streamdomain.Update_start_date
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	if err := s.StreamServise.Update_start_date(requestBody); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (s *StreamHandler) UpdateStreamInfo(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)

	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
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

	if err := s.StreamServise.UpdateStreamInfo(requestBody, IdUserTokenP); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (s *StreamHandler) UpdateModChat(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)

	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var requestBody streamdomain.UpdateModChat
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

	if err := s.StreamServise.UpdateModChat(requestBody, IdUserTokenP); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (s *StreamHandler) UpdateModChatSlowMode(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)

	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var requestBody streamdomain.UpdateModChatSlowMode
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

	if err := s.StreamServise.UpdateModChatSlowMode(requestBody, IdUserTokenP); err != nil {
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

func (s *StreamHandler) GetCategories(c *fiber.Ctx) error {

	Categorias, err := s.StreamServise.GetCategories()
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Categorias,
	})

}

type Categoria struct {
	Categoria string `json:"Categoria" query:"Categoria"`
}

func (s *StreamHandler) GetCategoria(c *fiber.Ctx) error {
	var CategoriaReq Categoria
	if err := c.QueryParser(&CategoriaReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	Categorias, err := s.StreamServise.GetCategoria(CategoriaReq.Categoria)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Categorias,
	})
}
