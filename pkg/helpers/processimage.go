package helpers

import (
	"PINKKER-BACKEND/config"
	"context"
	"mime/multipart"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

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
		// si fileHeader esta vacio quiero que devuelva esto
		PostImageChanel <- ""
	}
}
