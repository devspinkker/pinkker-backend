package withdrawroutes

import (
	withdrawapplication "PINKKER-BACKEND/internal/withdraw/withdraw-application"
	withdrawtinterfaces "PINKKER-BACKEND/internal/withdraw/withdraw-interface"
	withdrawalstinfrastructure "PINKKER-BACKEND/internal/withdraw/withdrawals-infrastructure"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func Withdrawroutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	withdrawRepository := withdrawalstinfrastructure.NewwithdrawalsRepository(redisClient, newMongoDB)
	withdrawService := withdrawapplication.NewwithdrawalsService(withdrawRepository)
	withdrawHandler := withdrawtinterfaces.NewwithdrawService(withdrawService)

	App.Post("/Withdraw/WithdrawalRequest", middleware.UseExtractor(), withdrawHandler.WithdrawalRequest)
	App.Post("/Withdraw/GetWithdrawalRequest", middleware.UseExtractor(), withdrawHandler.GetWithdrawalRequest)
	App.Post("/Withdraw/AcceptWithdrawal", middleware.UseExtractor(), withdrawHandler.AcceptWithdrawal)

	App.Post("/Withdraw/RejectWithdrawal", middleware.UseExtractor(), withdrawHandler.RejectWithdrawal)
	App.Get("/Withdraw/GetWithdrawalToken", middleware.UseExtractor(), withdrawHandler.AcceptWithdrawal)
}
