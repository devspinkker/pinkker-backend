package advertisementsinterface

import (
	"PINKKER-BACKEND/config"
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	advertisementsapplication "PINKKER-BACKEND/internal/advertisements/advertisements-application"
	"PINKKER-BACKEND/pkg/helpers"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdvertisementsRepository struct {
	Servise *advertisementsapplication.AdvertisementsService
}

func NewwithdrawService(Servise *advertisementsapplication.AdvertisementsService) *AdvertisementsRepository {
	return &AdvertisementsRepository{
		Servise: Servise,
	}
}

func (s *AdvertisementsRepository) BuyadCreate(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	nameUser := c.Context().UserValue("nameUser").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.UpdateAdvertisement

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	req.NameUser = nameUser
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	advertisementsGet, err := s.Servise.BuyadCreate(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}

func (s *AdvertisementsRepository) GetAdvertisements(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req advertisements.AdvertisementGet

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	page, err := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "InvalidPageNumber",
		})
	}

	advertisementsGet, err := s.Servise.GetAdvertisements(idValueObj, req, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}
func (s *AdvertisementsRepository) AcceptPendingAds(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req advertisements.AcceptPendingAds

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	err := s.Servise.AcceptPendingAds(idValueObj, req)
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
func (s *AdvertisementsRepository) GetAllPendingAds(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req advertisements.AdvertisementGet

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	page, err := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "InvalidPageNumber",
		})
	}

	pendings, err := s.Servise.GetAllPendingAds(idValueObj, req, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    pendings,
	})
}

func (s *AdvertisementsRepository) GetAdsUserPendingCode(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req advertisements.AcceptPendingAds

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	page, err := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "InvalidPageNumber",
		})
	}

	pendings, err := s.Servise.GetAllPendingNameUserAds(idValueObj, req, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    pendings,
	})
}
func (s *AdvertisementsRepository) RemovePendingAds(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req advertisements.AcceptPendingAds

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	err := s.Servise.RemovePendingAds(idValueObj, req)
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

func (s *AdvertisementsRepository) GetAdsUser(c *fiber.Ctx) error {
	nameUser := c.Context().UserValue("nameUser").(string)

	advertisementsGet, err := s.Servise.GetAdsUser(nameUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}
func (s *AdvertisementsRepository) GetAdsUserCode(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.GetAdsUserCode

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	advertisementsGet, err := s.Servise.GetAdsUserCode(req, idValueObj)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}
func (s *AdvertisementsRepository) CreateAdvertisement(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.UpdateAdvertisement

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	advertisementsGet, err := s.Servise.CreateAdvertisement(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}

func (s *AdvertisementsRepository) IdOfTheUsersWhoClicked(c *fiber.Ctx) error {
	type request struct {
		IdAdvertisements primitive.ObjectID `json:"idAdvertisements" `
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req request

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	err := s.Servise.IdOfTheUsersWhoClicked(idValueObj, req.IdAdvertisements)
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

func (s *AdvertisementsRepository) UpdateAdvertisement(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.UpdateAdvertisement

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	advertisementsGet, err := s.Servise.UpdateAdvertisement(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}
func (s *AdvertisementsRepository) DeleteAdvertisement(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.DeleteAdvertisement

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	err := s.Servise.DeleteAdvertisement(idValueObj, req)
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
func (s *AdvertisementsRepository) CreateAdsClips(c *fiber.Ctx) error {
	var req advertisements.ClipAdsCreate

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	file, err := c.FormFile("video")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "No video file provided",
			"data":    err.Error(),
		})
	}

	idValue := c.Context().UserValue("_id").(string)
	nameUser := c.Context().UserValue("nameUser").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusNetworkAuthenticationRequired).JSON(fiber.Map{
			"message": "StatusNetworkAuthenticationRequired",
			"data":    errorID,
		})
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error al obtener el directorio actual",
			"data":    err.Error(),
		})
	}

	baseDir := filepath.Join(currentDir, "clips_CreateAds")
	videoDir := filepath.Join(baseDir, uuid.NewString())
	if err := os.MkdirAll(videoDir, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error al crear directorio de video",
			"data":    err.Error(),
		})
	}

	videoPath := filepath.Join(videoDir, file.Filename)
	if err := c.SaveFile(file, videoPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error al guardar el archivo de video",
			"data":    err.Error(),
		})
	}

	outputFilePath := filepath.Join(videoDir, "output.mp4")
	FFmpegPath := config.FFmpegPath()

	// Usar FFmpeg para procesar el video recibido
	cmd := exec.Command(
		FFmpegPath,
		"-i", videoPath,
		"-t", "60",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-strict", "experimental",
		"-b:a", "192k",
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-y",
		outputFilePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("FFmpeg error: %s\n", string(output))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error al procesar el video",
			"data":    err.Error(),
		})
	}

	// Crear el clip en el servicio
	clipCreated, err := s.Servise.CreateClipForAds(idValueObj, nameUser, req.ClipTitle, outputFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error creando el clip",
			"data":    err.Error(),
		})
	}
	ads, err := s.Servise.BuyadClipCreate(idValueObj, req, clipCreated.ID)
	if err != nil {
		fmt.Println(ads)
		fmt.Println(err.Error())
		fmt.Println(req)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	cloudinaryResponse, err := helpers.UploadVideo(outputFilePath)
	if err != nil {
		fmt.Printf("Error al subir el video a Cloudinary: %v\n", err)
	}

	go func() {
		s.Servise.UpdateClip(clipCreated, cloudinaryResponse)
	}()

	defer os.RemoveAll(videoDir)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "ad clip creado con exito",
		"data":    ads,
	})
}
