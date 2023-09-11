package userinterfaces

import (
	application "PINKKER-BACKEND/internal/user/user-application"
	domain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"PINKKER-BACKEND/pkg/jwt"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserHandler struct {
	userService *application.UserService
}

func NewUserHandler(chatService *application.UserService) *UserHandler {
	return &UserHandler{
		userService: chatService,
	}
}

// signup
func (h *UserHandler) Signup(c *fiber.Ctx) error {
	var newUser domain.UserModelValidator
	fileHeader, _ := c.FormFile("avatar")
	PostImageChanel := make(chan string)
	errChanel := make(chan error)
	go helpers.Processimage(fileHeader, PostImageChanel, errChanel)

	if err := c.BodyParser(&newUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"messages": "Bad Request",
		})
	}
	if err := newUser.ValidateUser(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   err.Error(),
		})
	}

	// password
	passwordHashChan := make(chan string)
	go helpers.HashPassword(newUser.Password, passwordHashChan)

	_, existUser := h.userService.FindNameUser(newUser.NameUser, newUser.Email)
	if existUser != nil {
		// si no exiaste crealo
		if existUser == mongo.ErrNoDocuments {
			passwordHash := <-passwordHashChan
			if passwordHash == "error" {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "Internal Server Error hash",
				})
			}
			for {
				select {
				case avatarUrl := <-PostImageChanel:
					err := h.userService.SaveUser(&newUser, avatarUrl, passwordHash)
					if err != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"message": "Internal Server Error",
							"err":     err,
						})
					}

					tokenEmailConfirmation, ErrtokenEmailConfirmation := jwt.CreateTokenEmailConfirmation(&newUser)

					if ErrtokenEmailConfirmation != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"message": "Err token Email Confirmation",
						})
					}
					errSendConfirmationEmail := h.userService.SendConfirmationEmail(newUser.Email, tokenEmailConfirmation)
					if errSendConfirmationEmail != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"message": "err Send Confirmation Email",
						})
					}
					return c.Status(fiber.StatusOK).JSON(fiber.Map{
						"message": "save user",
					})

				case <-errChanel:
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"message": "avatarUrl error",
					})
				}

			}
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	} else {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "exist NameUser or Email",
		})
	}
}
func (h *UserHandler) ConfirmEmailSignup(c *fiber.Ctx) error {
	token := c.Params("token")

	nameUser, err := jwt.ExtractDataFromTokenConfirmEmail(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "StatusUnauthorized",
		})
	}

	errConfirmationEmailToken := h.userService.ConfirmationEmailToken(nameUser)
	if errConfirmationEmailToken != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "StatusCreated",
	})

}

// login
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var DataForLogin domain.LoginValidatorStruct

	if err := c.BodyParser(&DataForLogin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}
	if err := DataForLogin.LoginValidator(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   err.Error(),
		})
	}
	user, errGoMongoDBCollUsers := h.userService.FindNameUser(DataForLogin.NameUser, "")
	if errGoMongoDBCollUsers != nil {
		if errGoMongoDBCollUsers == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal Server Error",
			})
		}
	}
	if err := helpers.DecodePassword(user.PasswordHash, DataForLogin.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	token, err := jwt.CreateToken(user)
	if err != nil {
		log.Fatal("Login,CreateTokenError", err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Token created",
		"data":    token,
	})
}

// get User By Id
func (h *UserHandler) GetUserById(c *fiber.Ctx) error {

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	user, err := h.userService.FindUserById(IdUserTokenP)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    user,
	})
}

type IdReq struct {
	IdUser string `json:"IdUser"`
}

// follow
func (h *UserHandler) Follow(c *fiber.Ctx) error {

	var idReq IdReq
	c.BodyParser(&idReq)
	IdUser, err := primitive.ObjectIDFromHex(idReq.IdUser)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	errUpdateUserFollow := h.userService.FollowUser(IdUserTokenP, IdUser)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Follow",
	})
}

func (h *UserHandler) Unfollow(c *fiber.Ctx) error {

	var idReq IdReq
	c.BodyParser(&idReq)
	IdUser, err := primitive.ObjectIDFromHex(idReq.IdUser)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	errUpdateUserFollow := h.userService.Unfollow(IdUserTokenP, IdUser)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Unfollow",
	})
}

// buy pixeless

func (h *UserHandler) BuyPixeles(c *fiber.Ctx) error {

	return nil
}

type ReqGetUserBykey struct {
	Key string `json:"key"`
}

// follow
func (h *UserHandler) GetUserBykey(c *fiber.Ctx) error {

	var Req ReqGetUserBykey
	err := c.BodyParser(&Req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	user, errGetUserBykey := h.userService.GetUserBykey(Req.Key)
	if errGetUserBykey != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    user,
	})
}
