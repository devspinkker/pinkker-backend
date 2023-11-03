package userinterfaces

import (
	application "PINKKER-BACKEND/internal/user/user-application"
	domain "PINKKER-BACKEND/internal/user/user-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	oauth2 "PINKKER-BACKEND/pkg/OAuth2"
	configoauth2 "PINKKER-BACKEND/pkg/OAuth2/configOAuth2"
	"PINKKER-BACKEND/pkg/helpers"
	"PINKKER-BACKEND/pkg/jwt"
	"context"
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
					userDomaion := h.userService.UserDomaionUpdata(&newUser, avatarUrl, passwordHash)
					idInsert, err := h.userService.SaveUser(userDomaion)
					if err != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"message": "Internal Server Error",
							"err":     err,
						})
					}
					err = h.userService.CreateStream(userDomaion, idInsert)
					if err != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"message": "Internal Server Error",
							"err":     err,
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
		"_id":     user.ID,
		"avatar":  user.Avatar,
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

	user, errGetUserBykey := h.userService.FindNameUser(Req.NameUser, "")
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
		"message": "token",
		"data":    tokenRequest,
		"_id":     user.ID,
		"avatar":  user.Avatar,
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
	user, err := h.userService.FindEmailForOauth2Updata(&req)
	if err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    err.Error(),
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
		"message": "token",
		"data":    tokenRequest,
		"_id":     user.ID,
		"avatar":  user.Avatar,
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
			})
		case <-errChanel:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "avatarUrl error",
			})
		}

	}

}
