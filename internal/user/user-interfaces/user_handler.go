package userinterfaces

import (
	donationdomain "PINKKER-BACKEND/internal/donation/donation"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	application "PINKKER-BACKEND/internal/user/user-application"
	domain "PINKKER-BACKEND/internal/user/user-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	oauth2 "PINKKER-BACKEND/pkg/OAuth2"
	configoauth2 "PINKKER-BACKEND/pkg/OAuth2/configOAuth2"
	"PINKKER-BACKEND/pkg/auth"
	"PINKKER-BACKEND/pkg/helpers"
	"PINKKER-BACKEND/pkg/jwt"
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
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
func (h *UserHandler) GenerateTOTPKey(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)
	nameUser := c.Context().UserValue("nameUser").(string)

	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errinObjectID.Error(),
		})
	}
	secret, url, err := h.userService.GenerateTOTPKey(context.Background(), IdUserTokenP, nameUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
		"secret":  secret,
		"url":     url,
	})
}

func (h *UserHandler) ValidateTOTPCode(c *fiber.Ctx) error {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errinObjectID.Error(),
		})
	}
	valid, err := h.userService.ValidateTOTPCode(context.Background(), IdUserTokenP, req.Code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	if !valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})
}
func (h *UserHandler) PanelAdminPinkkerInfoUser(c *fiber.Ctx) error {
	var PanelAdminPinkkerInfoUserReq domain.PanelAdminPinkkerInfoUserReq
	if err := c.BodyParser(&PanelAdminPinkkerInfoUserReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errinObjectID.Error(),
		})
	}
	user, stream, errUpdateUserFollow := h.userService.PanelAdminPinkkerInfoUser(PanelAdminPinkkerInfoUserReq, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errUpdateUserFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
		"user":    user,
		"stream":  stream,
	})

}

func (h *UserHandler) UpdateLastConnection(idToken primitive.ObjectID) error {
	return h.userService.UpdateLastConnection(idToken)

}
func (h *UserHandler) Pinker_notifications(c *websocket.Conn) error {

	sub := h.userService.SubscribeToRoom("pinker_notifications")

	for {
		go func() {
			for {
				if c == nil {
					fmt.Println("WebSocket connection is closed.")
					break
				}
				_, _, err := c.ReadMessage()
				if err != nil {
					h.userService.CloseSubscription(sub)
					c.Close()
					return
				}
			}
		}()

		message, err := sub.ReceiveMessage(context.Background())
		if err != nil {
			h.userService.CloseSubscription(sub)
			return err
		}

		err = c.WriteMessage(websocket.TextMessage, []byte(message.Payload))
		if err != nil {
			h.userService.CloseSubscription(sub)
			return err
		}
	}
}
func (h *UserHandler) PanelAdminPinkkerPartnerUser(c *fiber.Ctx) error {
	var PanelAdminPinkkerPartnerUser domain.PanelAdminPinkkerInfoUserReq
	if err := c.BodyParser(&PanelAdminPinkkerPartnerUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errinObjectID.Error(),
		})
	}
	errUpdateUserFollow := h.userService.PanelAdminPinkkerPartnerUser(PanelAdminPinkkerPartnerUser, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errUpdateUserFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})

}
func (h *UserHandler) ChangeNameUserCodeAdmin(c *fiber.Ctx) error {
	var CreateAdmin domain.ChangeNameUser
	if err := c.BodyParser(&CreateAdmin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errinObjectID.Error(),
		})
	}
	if err := CreateAdmin.ValidateUser(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err,
		})
	}
	errUpdateUserFollow := h.userService.ChangeNameUserCodeAdmin(CreateAdmin, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errUpdateUserFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})

}
func (h *UserHandler) ChangeNameUser(c *fiber.Ctx) error {
	var CreateAdmin domain.ChangeNameUser
	if err := c.BodyParser(&CreateAdmin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	nameuser := c.Context().UserValue("nameUser").(string)
	CreateAdmin.NameUserRemove = nameuser

	if err := CreateAdmin.ValidateUser(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err,
		})
	}
	errUpdateUserFollow := h.userService.ChangeNameUser(CreateAdmin)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errUpdateUserFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})

}
func (h *UserHandler) CreateAdmin(c *fiber.Ctx) error {
	var CreateAdmin domain.CreateAdmin
	if err := c.BodyParser(&CreateAdmin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errinObjectID.Error(),
		})
	}
	errUpdateUserFollow := h.userService.CreateAdmin(CreateAdmin, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errUpdateUserFollow.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})

}
func (h *UserHandler) PanelAdminPinkkerbanStreamer(c *fiber.Ctx) error {
	var PanelAdminPinkkerInfoUserReq domain.PanelAdminPinkkerInfoUserReq
	if err := c.BodyParser(&PanelAdminPinkkerInfoUserReq); err != nil {
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
	errUpdateUserFollow := h.userService.PanelAdminPinkkerbanStreamer(PanelAdminPinkkerInfoUserReq, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})
}
func (h *UserHandler) PanelAdminRemoveBanStreamer(c *fiber.Ctx) error {
	var PanelAdminPinkkerInfoUserReq domain.PanelAdminPinkkerInfoUserReq
	if err := c.BodyParser(&PanelAdminPinkkerInfoUserReq); err != nil {
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
	errUpdateUserFollow := h.userService.PanelAdminRemoveBanStreamer(PanelAdminPinkkerInfoUserReq, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})
}
func (h *UserHandler) SignupSaveUserRedis(c *fiber.Ctx) error {
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
					userDomaion := h.userService.UserDomaionUpdata(&newUser, avatarUrl, passwordHash)
					code, err := h.userService.SaveUserRedis(userDomaion)
					if err != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"message": "Internal Server Error",
							"err":     err,
						})
					}
					err = helpers.ResendConfirmMail(code, userDomaion.Email)
					if err != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"message": "Internal Server Error",
							"err":     err,
						})
					}
					return c.Status(fiber.StatusOK).JSON(fiber.Map{
						"message": "email to confirm",
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

type ReqCodeInRedisSignup struct {
	Code string `json:"code"`
}

func (h *UserHandler) SaveUserCodeConfirm(c *fiber.Ctx) error {
	var newUser ReqCodeInRedisSignup
	if err := c.BodyParser(&newUser); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"messages": "Bad Request",
		})
	}
	user, errGetUserinRedis := h.userService.GetUserinRedis(newUser.Code)
	if errGetUserinRedis != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"messages": "StatusNotFound",
			"data":     "not found code or not exist",
		})
	}
	streamID, err := h.userService.SaveUser(user)
	user.ID = streamID
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"messages": "StatusInternalServerError",
		})
	}
	err = h.userService.CreateStream(user, streamID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"messages": "Stream Create error",
		})
	}
	tokenRequest, err := jwt.CreateToken(user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "token error",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "token",
		"data":            tokenRequest,
		"_id":             user.ID,
		"avatar":          user.Avatar,
		"keyTransmission": user.KeyTransmission,
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
	user, errGoMongoDBCollUsers := h.userService.FindNameUserInternalOperation(DataForLogin.NameUser, "")
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

	IsUserBlocked, err := h.userService.IsUserBlocked(DataForLogin.NameUser)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "login failed",
		})
	}
	if IsUserBlocked {

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Too many failed login attempts. Please try again late",
		})
	}
	if err := helpers.DecodePassword(user.PasswordHash, DataForLogin.Password); err != nil {

		h.userService.HandleLoginFailure(DataForLogin.NameUser)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "login failed",
		})
	}
	if user.TOTPSecret != "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "TOTPSecret",
		})
	}
	token, err := jwt.CreateToken(user)
	if err != nil {
		log.Fatal("Login,CreateTokenError", err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "token",
		"data":            token,
		"_id":             user.ID,
		"avatar":          user.Avatar,
		"keyTransmission": user.KeyTransmission,
	})
}
func (h *UserHandler) LoginTOTPSecret(c *fiber.Ctx) error {
	var DataForLogin domain.LoginTOTPSecret

	if err := c.BodyParser(&DataForLogin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}
	if err := DataForLogin.LoginTOTPSecret(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   err.Error(),
		})
	}
	user, errGoMongoDBCollUsers := h.userService.FindNameUserInternalOperation(DataForLogin.NameUser, "")
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

	IsUserBlocked, err := h.userService.IsUserBlocked(DataForLogin.NameUser)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "login failed in block",
		})
	}
	if IsUserBlocked {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Too many failed login attempts. Please try again late",
		})
	}
	if err := helpers.DecodePassword(user.PasswordHash, DataForLogin.Password); err != nil {
		h.userService.HandleLoginFailure(DataForLogin.NameUser)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "login failed",
		})
	}
	valid, err := auth.TOTPAutheLogin(DataForLogin.Totpcode, user.TOTPSecret)
	if !valid || err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid",
			"data":    err,
		})
	}
	token, err := jwt.CreateToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "CreateToken err",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "token",
		"data":            token,
		"_id":             user.ID,
		"avatar":          user.Avatar,
		"keyTransmission": user.KeyTransmission,
	})
}

func (h *UserHandler) Get_Recover_lost_password(c *fiber.Ctx) error {
	var Get_new_password domain.Req_Recover_lost_password

	if err := c.BodyParser(&Get_new_password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}

	user, errGoMongoDBCollUsers := h.userService.FindNameUser("", Get_new_password.Mail)

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
	passerdGenerateCodecharset := helpers.GenerateCodecharset(10)
	err := helpers.ResendRecoverPassword(passerdGenerateCodecharset, Get_new_password.Mail)
	err = h.userService.RedisSaveAccountRecoveryCode(passerdGenerateCodecharset, *user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"data":    err,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})
}

func (h *UserHandler) RestorePassword(c *fiber.Ctx) error {
	var Get_new_password domain.ReqRestorePassword

	if err := c.BodyParser(&Get_new_password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}
	user, errGetUserinRedis := h.userService.GetUserinRedis(Get_new_password.Code)
	if errGetUserinRedis != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}
	user, errGoMongoDBCollUsers := h.userService.FindNameUser("", user.Email)

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
	passwordHashChan := make(chan string)
	go helpers.HashPassword(Get_new_password.Password, passwordHashChan)
	passwordHash := <-passwordHashChan

	err := h.userService.EditPasswordHast(passwordHash, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"data":    err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})
}
func (h *UserHandler) ChangeGoogleAuthenticator(c *fiber.Ctx) error {
	nameuser := c.Context().UserValue("nameUser").(string)

	user, errGoMongoDBCollUsers := h.userService.FindNameUser(nameuser, "")

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
	passerdGenerateCodecharset := helpers.GenerateCodecharset(5)
	err := helpers.ChangeGoogleAuthenticator(passerdGenerateCodecharset, user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"data":    err,
		})
	}
	err = h.userService.RedisSaveChangeGoogleAuthenticatorCode(passerdGenerateCodecharset, *user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"data":    err,
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})
}
func (h *UserHandler) DeleteGoogleAuthenticator(c *fiber.Ctx) error {
	var Get_new_password domain.ReqRestorePassword

	if err := c.BodyParser(&Get_new_password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}
	user, errGetUserinRedis := h.userService.RedisGetChangeGoogleAuthenticatorCode(Get_new_password.Code)
	if errGetUserinRedis != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
		})
	}
	user, errGoMongoDBCollUsers := h.userService.FindNameUser("", user.Email)

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

	err := h.userService.DeleteGoogleAuthenticator(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"data":    err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})
}

// get User By Id
func (h *UserHandler) GetUserByIdTheToken(c *fiber.Ctx) error {

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

func (h *UserHandler) GetNotificacionesLastConnection(c *fiber.Ctx) error {
	// Obtener el ID del token del usuario desde el contexto
	IdUserToken := c.Context().UserValue("_id").(string)
	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	// Convertir el ID del token del usuario a ObjectID
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}

	// Obtener las notificaciones llamando al servicio de usuario
	FollowInfo, ResDonation, Subscription, errUpdateUserFollow := h.userService.GetNotificacionesLastConnection(IdUserTokenP, page)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errUpdateUserFollow.Error(),
		})
	}

	// Estructura de los datos a devolver
	type dataStruct struct {
		FollowInfo   []userdomain.FollowInfoRes         `json:"FollowInfo"`
		ResDonation  []donationdomain.ResDonation       `json:"ResDonation"`
		Subscription []subscriptiondomain.ResSubscriber `json:"Subscription"`
	}

	// Responder con los datos de notificaciones
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data": dataStruct{
			FollowInfo:   FollowInfo,
			ResDonation:  ResDonation,
			Subscription: Subscription,
		},
	})
}
func (h *UserHandler) GetRecentNotis(c *fiber.Ctx) error {
	// Obtener el ID del token del usuario desde el contexto
	IdUserToken := c.Context().UserValue("_id").(string)
	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	// Convertir el ID del token del usuario a ObjectID
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}

	// Obtener las notificaciones llamando al servicio de usuario
	FollowInfo, ResDonation, Subscription, errUpdateUserFollow := h.userService.GetRecentNotis(IdUserTokenP, page)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errUpdateUserFollow.Error(),
		})
	}

	// Estructura de los datos a devolver
	type dataStruct struct {
		FollowInfo   []userdomain.FollowInfoRes         `json:"FollowInfo"`
		ResDonation  []donationdomain.ResDonation       `json:"ResDonation"`
		Subscription []subscriptiondomain.ResSubscriber `json:"Subscription"`
	}

	// Responder con los datos de notificaciones
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data": dataStruct{
			FollowInfo:   FollowInfo,
			ResDonation:  ResDonation,
			Subscription: Subscription,
		},
	})
}
func (h *UserHandler) PurchasePinkkerPrime(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}

	_, errUpdateUserFollow := h.userService.PurchasePinkkerPrime(IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errUpdateUserFollow,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

// follow
func (h *UserHandler) Follow(c *fiber.Ctx) error {

	var idReq IdReq
	err := c.BodyParser(&idReq)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	IdUser, err := primitive.ObjectIDFromHex(idReq.IdUser)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	IdUserToken := c.Context().UserValue("_id").(string)
	nameUser := c.Context().UserValue("nameUser").(string)

	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}

	if IdUserTokenP == IdUser {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    "toUser !== ",
		})
	}
	avatar, errUpdateUserFollow := h.userService.FollowUser(IdUserTokenP, IdUser)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errUpdateUserFollow,
		})
	}
	errdeleteUser := h.userService.DeleteRedisUserChatInOneRoom(IdUserTokenP, IdUser)
	if errdeleteUser != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errdeleteUser,
		})
	}
	h.NotifyActivityFeed(IdUser.Hex()+"ActivityFeed", nameUser, avatar)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Follow",
	})
}
func (h *UserHandler) NotifyActivityFeed(room, user, Avatar string) error {
	clients, err := h.userService.GetWebSocketActivityFeed(room)
	if err != nil {
		return err
	}

	notification := map[string]interface{}{
		"Type":     "follow",
		"Nameuser": user,
		"Avatar":   Avatar,
	}

	for _, client := range clients {
		err = client.WriteJSON(notification)
		if err != nil {
			return err
		}
	}

	return nil
}
func (h *UserHandler) Unfollow(c *fiber.Ctx) error {

	var idReq IdReq
	err := c.BodyParser(&idReq)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
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
	errdeleteUser := h.userService.DeleteRedisUserChatInOneRoom(IdUserTokenP, IdUser)
	if errdeleteUser != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errdeleteUser,
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
	Key string `json:"key" query:"key"`
}

func (h *UserHandler) GetUserBykey(c *fiber.Ctx) error {

	var Req ReqGetUserBykey
	if err := c.QueryParser(&Req); err != nil {
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
func (h *UserHandler) GetUserBanInstream(c *fiber.Ctx) error {

	var Req ReqGetUserBykey
	if err := c.QueryParser(&Req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	user, errGetUserBykey := h.userService.GetUserBanInstream(Req.Key)
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
func (h *UserHandler) GetUserByCmt(c *fiber.Ctx) error {
	type ReqGetUserByCmt struct {
		Cmt string `json:"Cmt" query:"Cmt"`
	}

	var Req ReqGetUserByCmt
	if err := c.QueryParser(&Req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	user, errGetUserBykey := h.userService.GetUserByCmt(Req.Cmt)
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

type ReqGetUserByNameUser struct {
	NameUser string `json:"nameUser" query:"nameUser"`
}

func (h *UserHandler) GetUserByNameUser(c *fiber.Ctx) error {

	var Req ReqGetUserByNameUser
	if err := c.QueryParser(&Req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	user, errGetUserBykey := h.userService.FindNameUserNoSensitiveInformationApli(Req.NameUser, "")
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
func (h *UserHandler) GetUserByNameUserIndex(c *fiber.Ctx) error {

	var Req ReqGetUserByNameUser
	if err := c.QueryParser(&Req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	user, errGetUserBykey := h.userService.GetUserByNameUserIndex(Req.NameUser)
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

// login google
func (h *UserHandler) GoogleLogin(c *fiber.Ctx) error {
	randomstate := helpers.GenerateStateOauthCookie(c)

	googleConfig := configoauth2.LoadConfig()
	url := googleConfig.AuthCodeURL(randomstate)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "statusOk",
		"redirect": url,
	})
}

func (h *UserHandler) Google_callback(c *fiber.Ctx) error {
	code := c.Query("code")
	googleConfig := configoauth2.LoadConfig()
	token, err := googleConfig.Exchange(context.TODO(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}

	userInfo, err := oauth2.GetUserInfoFromGoogle(token)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	user, existUser := h.userService.FindNameUser(userInfo.Name, userInfo.Email)
	if existUser != nil {
		if existUser == mongo.ErrNoDocuments {
			newUser := &userdomain.UserModelValidator{
				FullName: userInfo.Name,
				NameUser: "",
				Password: "",
				Pais:     "",
				Ciudad:   "",
				Email:    userInfo.Email,
			}

			userDomaion := h.userService.UserDomaionUpdata(newUser, userInfo.Picture, "")
			idInsert, errSaveUser := h.userService.SaveUser(userDomaion)
			if errSaveUser != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "StatusInternalServerError",
					"data":    errSaveUser.Error(),
				})
			}
			err = h.userService.CreateStream(userDomaion, idInsert)

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "redirect to complete user",
				"data":    userInfo.Email,
			})
		}

	}

	if user.NameUser == "" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "redirect to complete user",
			"data":    userInfo.Email,
		})
	}
	tokenRequest, err := jwt.CreateToken(user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "token",
		"data":            tokenRequest,
		"_id":             user.ID,
		"avatar":          user.Avatar,
		"keyTransmission": user.KeyTransmission,
	})
}

func (h *UserHandler) Google_callback_Complete_Profile_And_Username(c *fiber.Ctx) error {

	var req domain.Google_callback_Complete_Profile_And_Username
	err := c.BodyParser(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	if err = req.ValidateUser(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	passwordHashChan := make(chan string)
	go helpers.HashPassword(req.Password, passwordHashChan)
	passwordHash := <-passwordHashChan
	req.Password = passwordHash
	user, err := h.userService.FindEmailForOauth2Updata(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	user.NameUser = req.NameUser

	tokenRequest, err := jwt.CreateToken(user)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "token error",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "token",
		"data":            tokenRequest,
		"_id":             user.ID,
		"avatar":          user.Avatar,
		"keyTransmission": user.KeyTransmission,
	})

}

func (h *UserHandler) EditProfile(c *fiber.Ctx) error {

	var EditProfile domain.EditProfile
	if err := c.BodyParser(&EditProfile); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	if err := EditProfile.ValidateEditProfile(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err,
		})
	}
	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	errUpdateUserFollow := h.userService.EditProfile(EditProfile, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (h *UserHandler) EditAvatar(c *fiber.Ctx) error {
	fileHeader, _ := c.FormFile("avatar")
	PostImageChanel := make(chan string)
	errChanel := make(chan error)

	go helpers.Processimage(fileHeader, PostImageChanel, errChanel)

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	for {
		select {
		case avatarUrl := <-PostImageChanel:
			errUpdateUserFollow := h.userService.EditAvatar(avatarUrl, IdUserTokenP)
			if errUpdateUserFollow != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "StatusInternalServerError",
				})
			}
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "StatusOK",
				"avatar":  avatarUrl,
			})
		case <-errChanel:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "avatarUrl error",
			})
		}

	}

}
func (h *UserHandler) EditBanner(c *fiber.Ctx) error {
	fileHeader, _ := c.FormFile("Banner")
	PostImageChanel := make(chan string)
	errChanel := make(chan error)

	go helpers.Processimage(fileHeader, PostImageChanel, errChanel)

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	for {
		select {
		case Banner := <-PostImageChanel:
			errUpdateUserFollow := h.userService.EditBanner(Banner, IdUserTokenP)
			if errUpdateUserFollow != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": "StatusInternalServerError",
				})
			}
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "StatusOK",
				"Banner":  Banner,
			})
		case <-errChanel:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Banner error",
			})
		}

	}

}
func (h *UserHandler) EditSocialNetworks(c *fiber.Ctx) error {
	var SocialNetwork domain.SocialNetwork
	if err := c.BodyParser(&SocialNetwork); err != nil {
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
	errUpdateUserFollow := h.userService.EditSocialNetworks(SocialNetwork, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})

}

func (h *UserHandler) GetRecommendedUsers(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req userdomain.GetRecommended
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
		})
	}
	User, err := h.userService.GetRecommendedUsers(idValueObj, req.ExcludeIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    User,
	})
}
func (h *UserHandler) GetStreamAndUserData(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	nameUserToken := c.Context().UserValue("nameUser").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var Req ReqGetUserByNameUser
	if err := c.QueryParser(&Req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	stream, User, UserInfoRoom, err := h.userService.GetStreamAndUserData(Req.NameUser, idValueObj, nameUserToken)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}

	response := domain.ReqGetUserByNameUser{
		User:     User,
		Stream:   stream,
		UserInfo: UserInfoRoom,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    response,
	})
}

func (h *UserHandler) GetAllPendingNameUserAds(c *fiber.Ctx) error {

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	ads, errUpdateUserFollow := h.userService.GetAllPendingNameUserAds(page, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
		"data":    ads,
	})

}

func (h *UserHandler) GetAllAcceptedNameUserAds(c *fiber.Ctx) error {

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	ads, errUpdateUserFollow := h.userService.GetAllAcceptedNameUserAds(page, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
		"data":    ads,
	})

}
func (h *UserHandler) GetActiveAdsByEndAdCommunity(c *fiber.Ctx) error {

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	ads, errUpdateUserFollow := h.userService.GetActiveAdsByEndAdCommunity(page, IdUserTokenP)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
		"data":    ads,
	})

}
func (h *UserHandler) GetAdsByNameUser(c *fiber.Ctx) error {

	IdUserToken := c.Context().UserValue("_id").(string)
	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	page, errpage := strconv.Atoi(c.Query("page", "1"))
	if errpage != nil || page < 1 {
		page = 1
	}
	nameAd := c.Query("name", "")

	ads, errUpdateUserFollow := h.userService.GetAdsByNameUser(page, IdUserTokenP, nameAd)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
		"data":    ads,
	})

}

func (h *UserHandler) AcceptOrDeleteAdvertisement(c *fiber.Ctx) error {
	var data struct {
		Action bool               `json:"action"`
		AdId   primitive.ObjectID `json:"AdId"`
	}
	req := data
	if err := c.BodyParser(&req); err != nil {
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
	errUpdateUserFollow := h.userService.AcceptOrDeleteAdvertisement(IdUserTokenP, req.AdId, req.Action)
	if errUpdateUserFollow != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})

}
