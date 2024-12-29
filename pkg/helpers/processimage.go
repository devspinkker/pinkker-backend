package helpers

import (
	"PINKKER-BACKEND/config"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"mime/multipart"
	"os"
	"os/exec"
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
		basePath := filepath.Join(config.BasePathUpload(), "images", "Thumbnail")
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

func ProcessImageCategorias(fileHeader *multipart.FileHeader, postImageChannel chan string, errChannel chan error) {
	if fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			errChannel <- err
			return
		}
		defer file.Close()

		// Decodificar la imagen para validación y conversión a bytes
		img, _, err := image.Decode(file)
		if err != nil {
			errChannel <- fmt.Errorf("error decoding image: %v", err)
			return
		}

		// Convertir la imagen a un buffer JPEG temporal para que `cwebp` pueda procesarla
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, nil); err != nil {
			errChannel <- fmt.Errorf("error encoding image to JPEG: %v", err)
			return
		}

		// Ruta base de almacenamiento local
		basePath := filepath.Join(config.BasePathUpload(), "images", "categories")
		if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
			errChannel <- fmt.Errorf("error creating directories: %v", err)
			return
		}

		// Generar nombre de archivo con extensión .webp
		fileName := sanitizeFileName(basePath, fileHeader.Filename)
		outputFileName := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ".webp"
		outputPath := filepath.Join(basePath, outputFileName)

		// Crear archivo temporal para que `cwebp` lo procese
		tempInputPath := filepath.Join(basePath, "temp_input.jpg")
		if err := os.WriteFile(tempInputPath, buf.Bytes(), 0644); err != nil {
			errChannel <- fmt.Errorf("error creating temporary input file: %v", err)
			return
		}
		defer os.Remove(tempInputPath) // Eliminar archivo temporal después de usarlo

		// Ejecutar `cwebp` para convertir el archivo a WebP
		cmd := exec.Command("cwebp", tempInputPath, "-o", outputPath, "-q", "80") // -q: calidad (0-100)
		if err := cmd.Run(); err != nil {
			errChannel <- fmt.Errorf("error converting image to WebP: %v", err)
			return
		}

		// Enviar la ruta generada al canal
		postImageChannel <- fmt.Sprintf("%s/images/categories/%s", config.MediaBaseURL(), outputFileName)
	} else {
		postImageChannel <- ""
	}
}
