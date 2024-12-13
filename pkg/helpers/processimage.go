package helpers

import (
	"PINKKER-BACKEND/config"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// sanitizeFileName reemplaza espacios y asegura que el nombre del archivo no se repita.
func sanitizeFileName(basePath, originalName string) string {
	// Reemplaza espacios por guiones bajos
	name := strings.ReplaceAll(originalName, " ", "_")

	// Asegura que no exista un archivo con el mismo nombre
	finalName := name
	count := 1
	for {
		if _, err := os.Stat(filepath.Join(basePath, finalName)); os.IsNotExist(err) {
			break
		}
		// Agregar un sufijo al nombre del archivo
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		finalName = fmt.Sprintf("%s_%d%s", base, count, ext)
		count++
	}

	return finalName
}

// ProcessImageEmotes guarda imágenes de emotes en el servidor local
func ProcessImageEmotes(fileHeader *multipart.FileHeader, PostImageChanel chan string, errChanel chan error, nameUser, typeEmote string) {
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			errChanel <- err
			return
		}
		defer file.Close()

		fileSize := fileHeader.Size
		if fileSize > 1<<20 {
			errChanel <- errors.New("el tamaño de la imagen excede 1MB")
			return
		}

		// Definir la ruta de almacenamiento local
		basePath := filepath.Join(config.BasePathUpload(), "emotes", nameUser, typeEmote)

		if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
			errChanel <- err
			return
		}

		// Sanitizar el nombre del archivo
		fileName := sanitizeFileName(basePath, fileHeader.Filename)
		filePath := filepath.Join(basePath, fileName)

		out, err := os.Create(filePath)
		if err != nil {
			errChanel <- err
			return
		}
		defer out.Close()

		if _, err := file.Seek(0, 0); err != nil {
			errChanel <- err
			return
		}

		if _, err := out.ReadFrom(file); err != nil {
			errChanel <- err
			return
		}

		PostImageChanel <- fmt.Sprintf("%s/emotes/%s/%s/%s", config.MediaBaseURL(), nameUser, typeEmote, fileName)
	} else {
		PostImageChanel <- ""
	}
}

// ProcessImage guarda imágenes generales en el servidor local
func ProcessImage(fileHeader *multipart.FileHeader, PostImageChanel chan string, errChanel chan error) {
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			errChanel <- err
			return
		}
		defer file.Close()

		// Ruta base de almacenamiento local
		basePath := filepath.Join(config.BasePathUpload(), "images")
		if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
			errChanel <- err
			return
		}

		// Sanitizar el nombre del archivo
		fileName := sanitizeFileName(basePath, fileHeader.Filename)
		filePath := filepath.Join(basePath, fileName)

		out, err := os.Create(filePath)
		if err != nil {
			errChanel <- err
			return
		}
		defer out.Close()

		if _, err := file.Seek(0, 0); err != nil {
			errChanel <- err
			return
		}

		if _, err := out.ReadFrom(file); err != nil {
			errChanel <- err
			return
		}

		PostImageChanel <- fmt.Sprintf("%s/images/%s", config.MediaBaseURL(), fileName)
	} else {
		PostImageChanel <- ""
	}
}

// UpdateClipPreviouImage guarda un archivo existente en una nueva ubicación
func UpdateClipPreviouImage(filePath string) (string, error) {
	basePath := filepath.Join(config.BasePathUpload(), "clips")
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		return "", err
	}

	fileName := sanitizeFileName(basePath, filepath.Base(filePath))
	newPath := filepath.Join(basePath, fileName)

	input, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer input.Close()

	output, err := os.Create(newPath)
	if err != nil {
		return "", err
	}
	defer output.Close()

	if _, err := input.Seek(0, 0); err != nil {
		return "", err
	}

	if _, err := output.ReadFrom(input); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/clips/%s", config.MediaBaseURL(), fileName), nil
}

// UploadVideo guarda videos en el servidor local
func UploadVideo(filePath string) (string, error) {
	// Ruta base para almacenar videos
	basePath := filepath.Join(config.BasePathUpload(), "videos")
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		return "", err
	}

	uniqueID := uuid.New().String()
	ext := filepath.Ext(filePath)

	fileName := fmt.Sprintf("%s%s", uniqueID, ext)
	newPath := filepath.Join(basePath, fileName)

	input, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer input.Close()

	output, err := os.Create(newPath)
	if err != nil {
		return "", err
	}
	defer output.Close()

	if _, err := input.Seek(0, 0); err != nil {
		return "", err
	}
	if _, err := output.ReadFrom(input); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/videos/%s", config.MediaBaseURL(), fileName), nil
}
