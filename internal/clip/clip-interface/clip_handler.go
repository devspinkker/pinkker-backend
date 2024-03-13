package clipinterface

import (
	"PINKKER-BACKEND/config"
	clipapplication "PINKKER-BACKEND/internal/clip/clip-application"
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ClipHandler struct {
	ClipService *clipapplication.ClipService
}

func NewClipHandler(ClipService *clipapplication.ClipService) *ClipHandler {
	return &ClipHandler{
		ClipService: ClipService,
	}
}

func (clip *ClipHandler) GetClipId(c *fiber.Ctx) error {
	clipIDStr := c.Query("clipId")
	clipID, err := primitive.ObjectIDFromHex(clipIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid clipId format",
			"data":    err.Error(),
		})
	}
	clipGet, err := clip.ClipService.GetClipId(clipID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	if strings.HasPrefix(clipGet.URL, "https://") {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":  "StatusOK",
			"dataClip": clipGet,
			"videoURL": true,
			"video":    false,
		})
	}
	localVideoPath := clipGet.URL
	localVideoContent, err := ioutil.ReadFile(localVideoPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error retrieving local video content",
			"data":    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "StatusOK",
		"dataClip": clipGet,
		"videoURL": false,
		"video":    localVideoContent,
	})
}

func (clip *ClipHandler) CreateClips(c *fiber.Ctx) error {
	var clipRequest clipdomain.ClipRequest
	if err := c.BodyParser(&clipRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}

	if err := clipRequest.ValidateClipRequest(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	ClipTitle := clipRequest.ClipTitle
	totalKey := clipRequest.TotalKey

	currentDir, err := os.Getwd()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error al crear directorio de video",
			"data":    err.Error(),
		})
	}

	baseDir := filepath.Join(currentDir, "clips_recortes")
	videoDir := filepath.Join(baseDir, uuid.NewString())
	videoPath := filepath.Join(videoDir, "input.mp4")
	outputFilePath := filepath.Join(videoDir, "output.mp4")

	if err := os.MkdirAll(videoDir, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInsufficientStorage).JSON(fiber.Map{
			"message": "Error al crear directorio de video",
			"data":    err.Error(),
		})
	}

	if err := ioutil.WriteFile(videoPath, clipRequest.Video, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInsufficientStorage).JSON(fiber.Map{
			"message": "Error al escribir el archivo de video",
			"data":    err.Error(),
		})
	}
	defer func() {

		if _, err := os.Stat(videoDir); err == nil {
			err := os.RemoveAll(videoDir)
			if err != nil {
				fmt.Println("Error removing directory:", err.Error())
			}
		} else if !os.IsNotExist(err) {
			fmt.Println("Error checking directory:", err.Error())
		}
	}()
	FFmpegPath := config.FFmpegPath()
	cmdSS := strconv.Itoa(clipRequest.Start)
	cmdT := strconv.Itoa(clipRequest.End - clipRequest.Start)

	cmd := exec.Command(
		FFmpegPath,
		"-i", videoPath,
		"-ss", cmdSS,
		"-t", cmdT,
		"-c", "copy",
		outputFilePath,
	)

	_, err = cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInsufficientStorage).JSON(fiber.Map{
			"message": "Error al recortar el video",
			"data":    err.Error(),
		})
	}

	idValue := c.Context().UserValue("_id").(string)
	nameUser := c.Context().UserValue("nameUser").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusNetworkAuthenticationRequired).JSON(fiber.Map{
			"message": "StatusNetworkAuthenticationRequired",
			"data":    err,
		})
	}
	clipCreated, err := clip.ClipService.CreateClip(idValueObj, totalKey, nameUser, ClipTitle, outputFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	cloudinaryResponse, err := helpers.UploadVideo(outputFilePath)
	if err != nil {
		fmt.Printf("Error al subir el video a Cloudinary: %v\n", err)
	}
	go func() {

		clip.ClipService.UpdateClip(clipCreated, cloudinaryResponse)
		if _, err := os.Stat(videoDir); err == nil {
			err := os.RemoveAll(videoDir)
			if err != nil {
				fmt.Println("Error removing directory:", err.Error())
			}
		} else if !os.IsNotExist(err) {
			fmt.Println("Error checking directory:", err.Error())
		}
	}()
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "StatusCreated",
		"data":    clipCreated.ID,
	})
}
func (clip *ClipHandler) GetClipsNameUser(c *fiber.Ctx) error {

	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	NameUser := c.Query("NameUser")

	Clips, errClipsGetFollow := clip.ClipService.GetClipsNameUser(page, NameUser)
	if errClipsGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errClipsGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Clips,
	})
}
func (clip *ClipHandler) GetClipsCategory(c *fiber.Ctx) error {

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	Category := c.Query("Category")
	lastClip := c.Query("lastClip")
	var lastClipId primitive.ObjectID
	if lastClip == "" {
		lastClipId = primitive.ObjectID{}
	} else {
		lastClipId, err = primitive.ObjectIDFromHex(lastClip)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "StatusBadRequest",
				"data":    err.Error(),
			})
		}
	}
	Clips, errClipsGetFollow := clip.ClipService.GetClipsCategory(page, Category, lastClipId)
	if errClipsGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errClipsGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Clips,
	})
}
func (clip *ClipHandler) GetClipsMostViewed(c *fiber.Ctx) error {

	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}

	Clips, errClipsGetFollow := clip.ClipService.GetClipsMostViewed(page)
	if errClipsGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errClipsGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Clips,
	})
}
func (clip *ClipHandler) GetClipsMostViewedLast48Hours(c *fiber.Ctx) error {

	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}

	Clips, errClipsGetFollow := clip.ClipService.GetClipsMostViewedLast48Hours(page)
	if errClipsGetFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errClipsGetFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Clips,
	})
}

type IDClipT struct {
	IDClip string `json:"ClipId"`
}

func (clip *ClipHandler) CliptLike(c *fiber.Ctx) error {

	var IDClipReq IDClipT
	if err := c.BodyParser(&IDClipReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}

	IDClip, err := primitive.ObjectIDFromHex(IDClipReq.IDClip)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    errorID.Error(),
		})
	}

	errLike := clip.ClipService.LikeClip(IDClip, idValueToken)
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
func (clip *ClipHandler) ClipDislike(c *fiber.Ctx) error {

	var IDClipReq IDClipT
	if err := c.BodyParser(&IDClipReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}

	IDClip, err := primitive.ObjectIDFromHex(IDClipReq.IDClip)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    errorID.Error(),
		})
	}

	errLike := clip.ClipService.ClipDislike(IDClip, idValueToken)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Dislike",
	})
}
func (clip *ClipHandler) MoreViewOfTheClip(c *fiber.Ctx) error {

	var IDClipReq IDClipT
	if err := c.BodyParser(&IDClipReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}

	IDClip, err := primitive.ObjectIDFromHex(IDClipReq.IDClip)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	errLike := clip.ClipService.MoreViewOfTheClip(IDClip)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Dislike",
	})
}
