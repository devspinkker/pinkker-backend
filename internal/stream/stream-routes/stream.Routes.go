package streamroutes

import (
	streamapplication "PINKKER-BACKEND/internal/stream/stream-application"
	streaminfrastructure "PINKKER-BACKEND/internal/stream/stream-infrastructure"
	streaminterfaces "PINKKER-BACKEND/internal/stream/stream-interface"
	"PINKKER-BACKEND/pkg/middleware"
	"PINKKER-BACKEND/pkg/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func StreamsRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	streamRepository := streaminfrastructure.NewStreamRepository(redisClient, newMongoDB)
	streamService := streamapplication.NewStreamService(streamRepository)
	streamHandler := streaminterfaces.NewStreamService(streamService)

	App.Post("/stream/getStreamById", streamHandler.GetStreamById)
	App.Get("/stream/getStreamByNameUser", streamHandler.GetStreamByNameUser)
	App.Get("/stream/getStreamsByCategorie", streamHandler.GetStreamsByCategorie)

	App.Get("/stream/GetAllsStreamsOnline", streamHandler.GetAllsStreamsOnline)
	App.Get("/stream/GetStreamsMostViewed", streamHandler.GetStreamsMostViewed)
	App.Get("/stream/GetAllsStreamsOnlineThatUserFollows", middleware.UseExtractor(), streamHandler.GetAllsStreamsOnlineThatUserFollows)
	App.Get("/stream/RecommendationStreams", streamHandler.RecommendationStreams)

	App.Get("/stream/Recommendation", middleware.UseExtractor(), streamHandler.GetAllsStreamsOnlineThatUserFollows)

	App.Post("/stream/getStreamsIdsStreamer", middleware.UseExtractor(), streamHandler.GetStreamsIdsStreamer)

	App.Post("/stream/update_online", streamHandler.Update_online)
	App.Post("/stream/closeStream", streamHandler.CloseStream)
	App.Post("/stream/update_thumbnail", streamHandler.Update_thumbnail)
	App.Post("/stream/update_start_date", streamHandler.Update_start_date)

	App.Post("/stream/update_stream_info", middleware.UseExtractor(), streamHandler.UpdateStreamInfo)
	App.Post("/stream/updateModChat", middleware.UseExtractor(), streamHandler.UpdateModChat)
	App.Post("/stream/updateModChatSlowMode", middleware.UseExtractor(), streamHandler.UpdateModChatSlowMode)
	// claudinary, push modificar el map, request
	App.Post("/stream/update_Emotes", streamHandler.Update_Emotes)

	App.Get("/stream/get_streamings_online", streamHandler.Streamings_online)

	App.Post("/stream/commercialInStream", middleware.UseExtractor(), streamHandler.CommercialInStream)
	App.Get("/ws/commercialInStream/:roomID", websocket.New(func(c *websocket.Conn) {
		roomID := c.Params("roomID")
		chatService := utils.NewChatService()
		client := &utils.Client{Connection: c}
		chatService.AddClientToRoom(roomID, client)

		// Asegurar cierre correcto de recursos
		defer func() {
			chatService.RemoveClientFromRoom(roomID, client)
			if err := c.Close(); err != nil {
				fmt.Printf("Error closing WebSocket connection: %v\n", err)
			}
			fmt.Println("WebSocket connection closed")
		}()

		// Bucle para leer mensajes
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					fmt.Printf("WebSocket closed by client: %v\n", err)
					break
				} else if websocket.IsUnexpectedCloseError(err) {
					fmt.Printf("Unexpected WebSocket close: %v\n", err)
				} else {
					fmt.Printf("WebSocket read error: %v\n", err)
				}
				break // Salir del bucle si hay error
			}
		}
	}))

	// esto se tiene que mover a una carpeta especifica
	App.Get("/categorie/GetCategories", streamHandler.GetCategories)
	App.Get("/categorie/GetCategoria", streamHandler.GetCategoria)
	App.Post("/categorie/update", middleware.UseExtractor(), streamHandler.CategoriesUpdate)

}
