package clipinterface

import (
	"PINKKER-BACKEND/config"
	clipapplication "PINKKER-BACKEND/internal/clip/clip-application"
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// Handler para obtener clips filtrados
func (clip *ClipHandler) GetClipsByNameUserIDOrdenación(c *fiber.Ctx) error {
	UserIDStr := c.Query("UserID", "")
	filter := c.Query("filter", "recent")
	dateRange := c.Query("dateRange", "")
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	UserID, err := primitive.ObjectIDFromHex(UserIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid StreamerID",
		})
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid page parameter",
		})
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid limit parameter",
		})
	}

	if dateRange == "" {
		var startDate time.Time

		if filter == "week" {
			startDate = time.Now().AddDate(0, 0, -7)
		} else if filter == "month" {
			startDate = time.Now().AddDate(0, -1, 0)
		} else if filter == "year" {
			startDate = time.Now().AddDate(-1, 0, 0)
		}

		if !startDate.IsZero() {
			dateRange = startDate.Format("2006-01-02")
		}
	}

	clips, err := clip.ClipService.GetClipsByNameUserIDOrdenación(UserID, filter, dateRange, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error fetching clips",
			"error":   err.Error(),
		})
	}

	// Devuelve los clips en la respuesta
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Clips fetched successfully",
		"data":    clips,
	})
}

func (clip *ClipHandler) DeleteClipByIDAndUserID(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    errorID.Error(),
		})
	}

	IdClipStr := c.Query("IdClip", "")
	if IdClipStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "IdClip parameter is required",
		})
	}

	IdClip, err := primitive.ObjectIDFromHex(IdClipStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid IdClip parameter",
		})
	}

	err = clip.ClipService.DeleteClipByIDAndUserID(IdClip, idValueToken)
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

func (clip *ClipHandler) UpdateClipTitle(c *fiber.Ctx) error {
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

	IdClipStr := c.Query("IdClip", "")
	if IdClipStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "IdClip parameter is required",
		})
	}

	IdClip, err := primitive.ObjectIDFromHex(IdClipStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid IdClip parameter",
		})
	}

	err = clip.ClipService.UpdateClipTitle(IdClip, idValueToken, title)
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
func (clip *ClipHandler) TimeOutClipCreate(c *fiber.Ctx) error {

	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusNetworkAuthenticationRequired).JSON(fiber.Map{
			"message": "StatusNetworkAuthenticationRequired",
			"data":    errorID,
		})
	}
	err := clip.ClipService.TimeOutClipCreate(idValueObj)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "err",
			"data":    err,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})
}
func (clip *ClipHandler) GetClipsByTitle(c *fiber.Ctx) error {
	title := c.Query("title", "")

	clipsGet, err := clip.ClipService.GetClipsByTitle(title)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "err",
			"data":    err,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
		"data":    clipsGet,
	})
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
func (clip *ClipHandler) GetClipIdLogueado(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusNetworkAuthenticationRequired",
			"data":    errorID,
		})
	}
	clipIDStr := c.Query("clipId")
	clipID, err := primitive.ObjectIDFromHex(clipIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid clipId format",
			"data":    err.Error(),
		})
	}
	clipGet, err := clip.ClipService.GetClipIdLogueado(clipID, idValueObj)
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

	idValue := c.Context().UserValue("_id").(string)
	nameUser := c.Context().UserValue("nameUser").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusNetworkAuthenticationRequired).JSON(fiber.Map{
			"message": "StatusNetworkAuthenticationRequired",
			"data":    errorID,
		})
	}

	errTimeOutClipCreate := clip.ClipService.TimeOutClipCreate(idValueObj)
	if errTimeOutClipCreate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    errTimeOutClipCreate.Error(),
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
	concatenatedFilePath := filepath.Join(videoDir, "input.ts")
	outputFilePath := filepath.Join(videoDir, "output.mp4")

	if err := os.MkdirAll(videoDir, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInsufficientStorage).JSON(fiber.Map{
			"message": "Error al crear directorio de video",
			"data":    err.Error(),
		})
	}

	// Obtener los datos de video
	videoData := clipRequest.TsUrls
	if len(videoData) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "No video data provided",
		})
	}

	// Descargar y concatenar segmentos .ts
	concatenatedFile, err := os.Create(concatenatedFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error al crear archivo concatenado",
			"data":    err.Error(),
		})
	}
	defer concatenatedFile.Close()

	for _, tsURL := range videoData {
		resp, err := http.Get(tsURL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error al descargar segmento de video",
				"data":    err.Error(),
			})
		}
		defer resp.Body.Close()

		_, err = io.Copy(concatenatedFile, resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error al escribir segmento de video",
				"data":    err.Error(),
			})
		}
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

	// Convertir el archivo .ts concatenado a .mp4 usando FFmpeg
	FFmpegPath := config.FFmpegPath()

	cmd := exec.Command(
		FFmpegPath,
		"-i", concatenatedFilePath,
		"-t", "70",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-strict", "experimental",
		"-b:a", "192k",
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-y", // overwrite output files without asking
		outputFilePath,
	)
	_, err = cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInsufficientStorage).JSON(fiber.Map{
			"message": "Error al recortar el video",
			"data":    err.Error(),
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
func (clip *ClipHandler) GetClipsWeightedByDate(c *fiber.Ctx) error {

	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}

	Clips, errClipsGetFollow := clip.ClipService.GetClipsWeightedByDate(page)
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
	idValue := c.Context().UserValue("_id").(string)
	idTPrimitive, err := primitive.ObjectIDFromHex(idValue)
	if err != nil {
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
	errLike := clip.ClipService.MoreViewOfTheClip(IDClip, idTPrimitive)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (clip *ClipHandler) ClipsRecommended(c *fiber.Ctx) error {

	var req clipdomain.GetRecommended
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusNetworkAuthenticationRequired).JSON(fiber.Map{
			"message": "StatusNetworkAuthenticationRequired",
			"data":    errorID,
		})
	}
	clips, errLike := clip.ClipService.ClipsRecommended(idValueObj, req.ExcludeIDs)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}

	ClipAd, err := clip.ClipService.GetClipTheAd(clips)
	if err != nil {
		fmt.Println(err)
	}
	if ClipAd.URL != "" {
		clips = append(clips, ClipAd)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    clips,
	})
}

func (clip *ClipHandler) CommentClip(c *fiber.Ctx) error {

	var req clipdomain.CommentClip
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}

	idValue := c.Context().UserValue("_id").(string)
	nameuser := c.Context().UserValue("nameUser").(string)

	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    errorID.Error(),
		})
	}

	comment, errLike := clip.ClipService.CommentClip(req.IdClip, idValueToken, nameuser, req.CommentClip)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "commented",
		"data":    comment,
	})
}
func (clip *ClipHandler) DeleteComment(c *fiber.Ctx) error {

	var req clipdomain.CommentClipId
	if err := c.BodyParser(&req); err != nil {
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

	errLike := clip.ClipService.DeleteComment(req.IdClip, idValueToken)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "commented delete",
	})
}
func (clip *ClipHandler) LikeCommentClip(c *fiber.Ctx) error {

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

	errLike := clip.ClipService.LikeCommentClip(IDClip, idValueToken)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "LikeCommentClip",
	})
}

func (clip *ClipHandler) UnlikeComment(c *fiber.Ctx) error {

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

	errLike := clip.ClipService.UnlikeComment(IDClip, idValueToken)
	if errLike != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errLike.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "LikeCommentClip",
	})
}

func (clip *ClipHandler) GetClipComments(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	IdClipStr := c.Query("IdClip", "")
	if IdClipStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "IdClip parameter is required",
		})
	}

	IdClip, err := primitive.ObjectIDFromHex(IdClipStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid IdClip parameter",
		})
	}

	Clips, err := clip.ClipService.GetClipComments(IdClip, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Clips,
	})
}
func (clip *ClipHandler) GetClipCommentsLoguedo(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueToken, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    errorID.Error(),
		})
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	IdClipStr := c.Query("IdClip", "")
	if IdClipStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "IdClip parameter is required",
		})
	}

	IdClip, err := primitive.ObjectIDFromHex(IdClipStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid IdClip parameter",
		})
	}

	Clips, err := clip.ClipService.GetClipCommentsLoguedo(IdClip, page, idValueToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Clips,
	})
}
