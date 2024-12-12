package helpers

import (
	"PINKKER-BACKEND/config"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
)

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

		filePath := filepath.Join(basePath, fileHeader.Filename)
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

		PostImageChanel <- fmt.Sprintf("%s/emotes/%s/%s/%s", config.MediaBaseURL(), nameUser, typeEmote, fileHeader.Filename)
	} else {
		PostImageChanel <- ""
	}
}

func ProcessImage(fileHeader *multipart.FileHeader, PostImageChanel chan string, errChanel chan error) {
	fmt.Println("Iniciando ProcessImage")
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			errChanel <- err
			return
		}
		defer file.Close()

		// Ruta base de almacenamiento local
		basePath := fmt.Sprintf("%s/images", config.BasePathUpload())
		fmt.Println("Base path:", basePath)

		// Crear el directorio si no existe
		if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
			fmt.Printf("Error creando directorio %s: %v\n", basePath, err)
			errChanel <- err
			return
		}

		// Ruta completa del archivo
		filePath := filepath.Join(basePath, fileHeader.Filename)
		fmt.Println("File path:", filePath)

		// Crear archivo
		out, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creando archivo %s: %v\n", filePath, err)
			errChanel <- err
			return
		}
		defer out.Close()

		// Escribir datos en el archivo
		if _, err := file.Seek(0, 0); err != nil {
			errChanel <- err
			return
		}

		bytesWritten, err := out.ReadFrom(file)
		if err != nil {
			fmt.Printf("Error escribiendo archivo %s: %v\n", filePath, err)
			errChanel <- err
			return
		}

		fmt.Printf("Bytes escritos: %d\n", bytesWritten)
		PostImageChanel <- fmt.Sprintf("%s/images/%s", config.MediaBaseURL(), fileHeader.Filename)
	} else {
		PostImageChanel <- ""
	}
}

// UpdateClipPreviouImage guarda un archivo existente en una nueva ubicación
func UpdateClipPreviouImage(filePath string) (string, error) {
	newPath := filepath.Join(config.BasePathUpload(), "clips", filepath.Base(filePath))

	if err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm); err != nil {
		return "", err
	}

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

	return fmt.Sprintf("%s/clips/%s", config.MediaBaseURL(), filepath.Base(newPath)), nil
}

// UploadVideo guarda videos en el servidor local
func UploadVideo(filePath string) (string, error) {
	newPath := filepath.Join(config.BasePathUpload(), "videos", filepath.Base(filePath))

	if err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm); err != nil {
		return "", err
	}

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

	return fmt.Sprintf("%s/videos/%s", config.MediaBaseURL(), filepath.Base(newPath)), nil
}
