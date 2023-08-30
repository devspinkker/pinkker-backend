package clipapplication

import (
	"PINKKER-BACKEND/config"
	clipinfrastructure "PINKKER-BACKEND/internal/clip/clip-infrastructure"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type ClipService struct {
	ClipRepository *clipinfrastructure.ClipRepository
}

func NewClipService(ClipRepository *clipinfrastructure.ClipRepository) *ClipService {
	return &ClipService{
		ClipRepository: ClipRepository,
	}
}

func (c *ClipService) CreateClip(streamer *userdomain.User, startTime string, length string, clipName string) error {
	cloudinaryURL := config.CLOUDINARY_URL()
	cld, errcloudinary := cloudinary.NewFromURL(cloudinaryURL)
	if errcloudinary != nil {
		return errcloudinary
	}
	FOLDER_CLIPS_OUTPUT := config.FOLDER_CLIPS_OUTPUT()

	streamKey := streamer.KeyTransmission[4:]

	RMTP_FOLDER_MEDIA := config.RMTP_FOLDER_MEDIA()
	files, err := ioutil.ReadDir(RMTP_FOLDER_MEDIA)
	if err != nil {
		log.Fatalf("Error al leer el directorio: %v", err)
	}

	videoFile := helpers.GetNewestFile(files, "")
	liveURL := filepath.Join(RMTP_FOLDER_MEDIA, streamKey, videoFile)

	outputURL := FOLDER_CLIPS_OUTPUT + clipName + ".mp4"
	startTimeInt, err := strconv.Atoi(startTime)
	if err != nil {
		return fmt.Errorf("invalid start time: %v", err)
	}

	lengthInt, err := strconv.Atoi(length)
	if err != nil {
		return fmt.Errorf("invalid length: %v", err)
	}

	exec.Command("ffmpeg",
		"-i", liveURL,
		"-ss", strconv.Itoa(startTimeInt),
		"-t", strconv.Itoa(lengthInt),
		"-c:v", "copy",
		"-c:a", "copy",
		outputURL,
	)
	EnvCloudUploadFolderClip := config.EnvCloudUploadFolderClip()
	if EnvCloudUploadFolderClip == "" {
		EnvCloudUploadFolderClip = "EnvCloudUploadFolderClip"
	}
	_, errurl := cld.Upload.Upload(context.Background(), outputURL, uploader.UploadParams{Folder: EnvCloudUploadFolderClip})
	if errurl != nil {
		return err
	}
	// newClip := &clipdomain.Clip{
	// 	// StreamerID: ,
	// 	URL: url.URL,
	// }
	// err = c.ClipRepository.SaveClip(newClip)

	return err
}
func (u *ClipService) FindUser(NameUser string) (*userdomain.User, error) {
	findUserInDbExist, errCollUsers := u.ClipRepository.FindUser(NameUser)
	return findUserInDbExist, errCollUsers
}
