package config

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func PORT() string {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println(err.Error())

		log.Fatal("godotenv.Load error PORT")
	}
	return os.Getenv("PORT")
}

func PASSWORDREDIS() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error PASSWORDREDIS")
	}
	return os.Getenv("PASSWORDREDIS")
}

func ADDRREDIS() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error ADDRREDIS")
	}
	return os.Getenv("ADDRREDIS")
}
func CLOUDINARY_URL() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error CLOUDINARY_URL")
	}
	return os.Getenv("CLOUDINARY_URL")
}
func MONGODB_URI() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error URI mongo")
	}
	return os.Getenv("MONGODB_URI")

}

func TOKENPASSWORD() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("TOKENPASSWORD")
}

func SENDGRIDAPIKEY() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("SENDGRIDAPIKEY")
}
func RMTP_FOLDER_MEDIA() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("RMTP_FOLDER_MEDIA")
}
func FOLDER_CLIPS_OUTPUT() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("FOLDER_CLIPS_OUTPUT")
}
func EnvCloudUploadFolderClip() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("EnvCloudUploadFolderClip")
}
func GOOGLE_CLIENT_ID() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("GOOGLE_CLIENT_ID")
}
func GOOGLE_CLIENT_SECRET() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("GOOGLE_CLIENT_SECRET")
}
func FFmpegPath() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("FFMPEG_PATH")
}
func ResendApi() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("RESENDAPI")
}
func ResendDominio() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("RESENDDOMINIO")
}

func AdvertisementsPayPerPrint() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("AdvertisementsPayPerPrint")
}
func AdvertisementsPayClicks() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("AdvertisementsPayClicks")
}
func SubsPayPerPrint() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("SubsPayPerPrint")
}
func PinkkerPrimeCost() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("PinkkerPrimeCost")
}
func CostToCreateCommunity() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("CostToCreateCommunity")
}
func FotoPerfilAleatoria() string {
	// Cargar las variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error cargando el archivo .env")
	}

	// Lista de IDs de perfiles disponibles
	ids := []string{"1", "2", "3", "4", "5"}

	// Semilla para generar un número aleatorio
	rand.Seed(time.Now().UnixNano())
	// Elegir un índice aleatorio de la lista
	randomID := ids[rand.Intn(len(ids))]

	// Retornar la URL correspondiente al ID seleccionado
	return os.Getenv("FOTO_PERFIL_" + randomID)
}
func BANNER() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("BANNER")
}
func BasePathUpload() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("BasePathUpload")
}
func MediaBaseURL() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("MediaBaseURL")
}
func KeyEmotes() string {
	if err := godotenv.Load(); err != nil {
		log.Fatal("godotenv.Load error")
	}
	return os.Getenv("KeyEmotes")
}
