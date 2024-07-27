package donationtinterfaces

import (
	donationdomain "PINKKER-BACKEND/internal/donation/donation"
	donationapplication "PINKKER-BACKEND/internal/donation/donation-application"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	banned, err := d.VodServise.StateTheUserInChat(idReq.ToUser, FromUser)
	if err != nil || banned {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "StatusConflict",
			"data":    "baneado",
		})
	}

	err = d.VodServise.UserHasNumberPikels(FromUser, idReq.Pixeles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	errdonatePixels := d.VodServise.DonatePixels(FromUser, idReq.ToUser, idReq.Pixeles, idReq.Text)
	if errdonatePixels != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": errdonatePixels.Error(),
		})
	}
	d.NotifyActivityFeed(idReq.ToUser.Hex()+"ActivityFeed", nameUser, idReq.Pixeles, idReq.Text)
	d.NotifyActivityToChat(idReq.ToUser, nameUser, idReq.Pixeles, idReq.Text)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (d *DonationHandler) NotifyActivityFeed(room, user string, Pixeles float64, text string) error {
	clients, err := d.VodServise.GetWebSocketActivityFeed(room)
	if err != nil {
		return err
	}

	notification := map[string]interface{}{
		"action":  "DonatePixels",
		"Pixeles": Pixeles,
		"data":    user,
		"text":    text,
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
