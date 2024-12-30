package helpers

import (
	"PINKKER-BACKEND/config"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/discord/lilliput"
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
		basePath := filepath.Join(config.BasePathUpload(), "emotes", typeEmote)

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

		PostImageChanel <- fmt.Sprintf("%s/emotes/%s/%s", config.MediaBaseURL(), typeEmote, fileName)
	} else {
		PostImageChanel <- ""
	}
}

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
func ProcessImageThumbnail(fileHeader *multipart.FileHeader, PostImageChanel chan string, errChanel chan error) {
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			errChanel <- err
			return
		}
		defer file.Close()

		// Ruta base de almacenamiento local
		basePath := filepath.Join(config.BasePathUpload(), "images", "categories")
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

		PostImageChanel <- fmt.Sprintf("%s/images/Thumbnail/%s", config.MediaBaseURL(), fileName)
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
func ProcessImageCategorias(fileHeader *multipart.FileHeader, PostImageChanel chan string, errChanel chan error) {
	if fileHeader == nil {
		PostImageChanel <- ""
		return
	}

	// Abrir el archivo cargado
	file, err := fileHeader.Open()
	if err != nil {
		errChanel <- fmt.Errorf("error opening file: %v", err)
		return
	}
	defer file.Close()

	// Leer los datos del archivo en memoria
	inputBuf, err := ioutil.ReadAll(file)
	if err != nil {
		errChanel <- fmt.Errorf("error reading file: %v", err)
		return
	}

	// Crear un decodificador con lilliput
	decoder, err := lilliput.NewDecoder(inputBuf)
	if err != nil {
		errChanel <- fmt.Errorf("error decoding image: %v", err)
		return
	}
	defer decoder.Close()

	// Crear operaciones de imagen con un buffer máximo de 8192x8192
	ops := lilliput.NewImageOps(8192)
	defer ops.Close()

	// Crear un buffer para almacenar la imagen procesada
	outputImg := make([]byte, 10*1024*1024) // 10MB

	// Ruta base de almacenamiento local
	basePath := filepath.Join(config.BasePathUpload(), "images", "categories")
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		errChanel <- fmt.Errorf("error creating directories: %v", err)
		return
	}

	// Generar el nombre del archivo con extensión .webp
	fileName := sanitizeFileName(basePath, fileHeader.Filename)
	outputFileName := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".webp"
	outputPath := filepath.Join(basePath, outputFileName)

	// Opciones de transformación
	opts := &lilliput.ImageOptions{
		FileType:             ".webp",
		Width:                0,                         // Mantener ancho original
		Height:               0,                         // Mantener altura original
		ResizeMethod:         lilliput.ImageOpsNoResize, // No redimensionar
		NormalizeOrientation: true,
	}

	// Transformar y codificar la imagen
	outputImg, err = ops.Transform(decoder, opts, outputImg)
	if err != nil {
		errChanel <- fmt.Errorf("error transforming image: %v", err)
		return
	}

	// Escribir el archivo procesado al disco
	if err := ioutil.WriteFile(outputPath, outputImg, 0644); err != nil {
		errChanel <- fmt.Errorf("error writing file: %v", err)
		return
	}

	// Enviar la URL generada al canal
	PostImageChanel <- fmt.Sprintf("%s/images/categories/%s", config.MediaBaseURL(), outputFileName)
}
