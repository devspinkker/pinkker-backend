package donationtinterfaces

import (
	donationdomain "PINKKER-BACKEND/internal/donation/donation"
	donationapplication "PINKKER-BACKEND/internal/donation/donation-application"
	"PINKKER-BACKEND/pkg/helpers"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DonationHandler struct {
	VodServise *donationapplication.DonationService
}

func NewDonationService(VodServise *donationapplication.DonationService) *DonationHandler {
	return &DonationHandler{
		VodServise: VodServise,
	}
}

func (d *DonationHandler) Donate(c *fiber.Ctx) error {
	var idReq donationdomain.ReqCreateDonation
	if err := c.BodyParser(&idReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if errValidate := idReq.ValidateReqCreateDonation(); errValidate != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    errValidate.Error(),
		})
	}

	nameUser := c.Context().UserValue("nameUser").(string)
	IdUserToken := c.Context().UserValue("_id").(string)

	FromUser, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	if FromUser == idReq.ToUser {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    "toUser !== ",
		})
	}
	banned, avatar, err := d.VodServise.StateTheUserInChat(idReq.ToUser, FromUser)
	if err != nil || banned {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "StatusConflict",
			"data":    "baneado",
		})
	}

	err = d.VodServise.UserHasNumberPikels(FromUser, idReq.Pixeles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "FromUser no tiene pixeles",
		})
	}
	errdonatePixels := d.VodServise.DonatePixels(FromUser, idReq.ToUser, idReq.Pixeles, idReq.Text)
	if errdonatePixels != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "errdonatePixels",
		})
	}

	LatestStreamSummaryByUpdate := d.VodServise.LatestStreamSummaryByUpdateDonations(idReq.ToUser, idReq.Pixeles)
	if errors.Is(LatestStreamSummaryByUpdate, mongo.ErrNoDocuments) {
		fmt.Println(LatestStreamSummaryByUpdate)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error update summary",
		})
	}
	IsFollowing, _ := d.VodServise.IsFollowing(FromUser, idReq.ToUser)
	d.NotifyActivityFeed(idReq.ToUser.Hex()+"ActivityFeed", nameUser, avatar, idReq.Pixeles, idReq.Text, FromUser, IsFollowing)
	d.NotifyActivityToChat(idReq.ToUser, nameUser, idReq.Pixeles, idReq.Text)

	Notification := helpers.CreateNotification("DonatePixels", nameUser, avatar, idReq.Text, idReq.Pixeles, FromUser)
	err = d.VodServise.SaveNotification(idReq.ToUser, Notification)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "ok",
			"data":    "SaveNotification error",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (d *DonationHandler) NotifyActivityFeed(room, user, Avatar string, Pixeles float64, text string, id primitive.ObjectID, IsFollowing bool) error {
	clients, err := d.VodServise.GetWebSocketActivityFeed(room)
	if err != nil {
		return err
	}

	notification := map[string]interface{}{
		"type":       "DonatePixels",
		"nameuser":   user,
		"Text":       text,
		"avatar":     Avatar,
		"Pixeles":    Pixeles,
		"idUser":     id,
		"isFollowed": IsFollowing,
		"timestamp":  time.Now(),
	}

	for _, client := range clients {
		err = client.WriteJSON(notification)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DonationHandler) NotifyActivityToChat(idReq primitive.ObjectID, user string, Pixeles float64, text string) error {

	notification := map[string]interface{}{
		"action":  "DonatePixels",
		"Pixeles": Pixeles,
		"data":    user,
		"text":    text,
	}
	return d.VodServise.PublishNotification(idReq, notification)

}
func (d *DonationHandler) Mydonors(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)

	id, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	donations, err := d.VodServise.MyPixelesdonors(id)
	if err != nil {
		if err.Error() == "no documents found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	errUpdateDonationsNotifiedStatus := d.VodServise.UpdateDonationsNotifiedStatus(donations)
	if errUpdateDonationsNotifiedStatus != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": errUpdateDonationsNotifiedStatus.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    donations,
	})
}

func (d *DonationHandler) AllMyPixelesDonors(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)
	id, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	donations, err := d.VodServise.AllMyPixelesDonors(id)
	if err != nil {
		if err.Error() == "no documents found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    donations,
	})
}

type ReqChatDonations struct {
	Id primitive.ObjectID `json:"-" query:"Toid"`
}

func (d *DonationHandler) GetPixelesDonationsChat(c *fiber.Ctx) error {
	var req ReqChatDonations
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}
	donations, err := d.VodServise.GetPixelesDonationsChat(req.Id)
	if err != nil {
		if err.Error() == "no documents found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    donations,
	})
}
