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

	ClipName := clipRequest.ClipName
	streamer := clipRequest.Streamer

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

	FFmpegPath := config.FFmpegPath()
	cmd := exec.Command(
		FFmpegPath,
		"-i", videoPath,
		"-ss", "0",
		"-t", "6",
		"-c", "copy",
		outputFilePath,
	)

	fmt.Println(outputFilePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error en la ejecución de FFmpeg:")
		fmt.Println(string(out))
		fmt.Println(err)
		return c.Status(fiber.StatusInsufficientStorage).JSON(fiber.Map{
			"message": "Error al recortar el video",
			"data":    err.Error(),
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInsufficientStorage).JSON(fiber.Map{
			"message": "Error al recortar el video",
			"data":    err.Error(),
		})
	}

	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusNetworkAuthenticationRequired).JSON(fiber.Map{
			"message": "StatusNetworkAuthenticationRequired",
			"data":    err.Error(),
		})
	}
	clipCreated, err := clip.ClipService.CreateClip(idValueObj, streamer, ClipName, outputFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "StatusInternalServerError",
		})
	}
	fmt.Println(outputFilePath)
	cloudinaryResponse, err := helpers.UploadVideo(outputFilePath)
	if err != nil {
		fmt.Printf("Error al subir el video a Cloudinary: %v\n", err)
	}
	fmt.Println(cloudinaryResponse)

	if err != nil {
		fmt.Println(err.Error())
	}
	go func() {

		clip.ClipService.UpdateClip(clipCreated, cloudinaryResponse)
		if _, err := os.Stat(videoDir); err == nil {
			err := os.RemoveAll(videoDir)
			if err != nil {
				fmt.Println("Error removing directory:", err.Error())
			} else {
				fmt.Println("Directory removed successfully.")
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
