package helpers

import (
	"PINKKER-BACKEND/config"
	"context"
	"errors"
	"mime/multipart"
	"os"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func ProcessImageEmotes(fileHeader *multipart.FileHeader, PostImageChanel chan string, errChanel chan error) {
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			errChanel <- err
			return
		}
		defer file.Close()

		ctx := context.Background()

		fileSize := fileHeader.Size
		if fileSize > 1<<20 {
			errChanel <- errors.New("el tamaño de la imagen excede 1MB")
			return
		}

		cldService, err := cloudinary.NewFromURL(config.CLOUDINARY_URL())
		if err != nil {
			errChanel <- err
			return
		}

		params := uploader.UploadParams{
			Folder: "emotes/",
		}
		resp, err := cldService.Upload.Upload(ctx, file, params)
		if err != nil {
			errChanel <- err
			return
		}

		if !strings.HasPrefix(resp.SecureURL, "https://") {
			errChanel <- errors.New("la URL de la imagen no tiene un protocolo seguro")
			return
		}

		PostImageChanel <- resp.SecureURL
	} else {
		PostImageChanel <- ""
	}
}
func Processimage(fileHeader *multipart.FileHeader, PostImageChanel chan string, errChanel chan error) {
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			errChanel <- err
			return
		}
		defer file.Close()

		ctx := context.Background()

		cldService, errcloudinary := cloudinary.NewFromURL(config.CLOUDINARY_URL())
		if errcloudinary != nil {
			errChanel <- errcloudinary
		}
		resp, errcldService := cldService.Upload.Upload(ctx, file, uploader.UploadParams{})

		if errcldService != nil || !strings.HasPrefix(resp.SecureURL, "https://") {
			errChanel <- errcldService
		}
		PostImageChanel <- resp.SecureURL
	} else {
		PostImageChanel <- ""
	}
}
func UpdateClipPreviouImage(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	ctx := context.Background()

	cldService, err := cloudinary.NewFromURL(config.CLOUDINARY_URL())
	if err != nil {
		return "", err
	}

	resp, err := cldService.Upload.Upload(ctx, file, uploader.UploadParams{})
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(resp.SecureURL, "https://") {
		return "", errors.New("Invalid secure URL format")
	}

	return resp.SecureURL, nil
}
func UploadVideo(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	ctx := context.Background()

	cldService, err := cloudinary.NewFromURL(config.CLOUDINARY_URL())
	if err != nil {
		return "", err
	}

	resp, err := cldService.Upload.Upload(ctx, file, uploader.UploadParams{})
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(resp.SecureURL, "https://") {
		return "", errors.New("Invalid secure URL format")
	}

	return resp.SecureURL, nil
}
