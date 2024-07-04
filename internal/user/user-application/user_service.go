package userapplication

import (
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	domain "PINKKER-BACKEND/internal/user/user-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	infrastructure "PINKKER-BACKEND/internal/user/user-infrastructure"
	"PINKKER-BACKEND/pkg/helpers"
	"strings"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	roomRepository *infrastructure.UserRepository
}

func NewChatService(roomRepository *infrastructure.UserRepository) *UserService {
	return &UserService{
		roomRepository: roomRepository,
	}
}

func (u *UserService) SubscribeToRoom(roomID string) *redis.PubSub {
	sub := u.roomRepository.SubscribeToRoom(roomID)
	return sub
}
func (u *UserService) CloseSubscription(sub *redis.PubSub) error {
	return u.roomRepository.CloseSubscription(sub)
}

// signup
func (u *UserService) UserDomaionUpdata(newUser *domain.UserModelValidator, avatarUrl string, passwordHash string) *userdomain.User {
	var modelNewUser domain.User
	KeyTransmission := helpers.KeyTransmission(45)
	cmt := helpers.KeyTransmission(32)

	modelNewUser.Avatar = avatarUrl
	if modelNewUser.Avatar == "" {
		modelNewUser.Avatar = "https://res.cloudinary.com/pinkker/image/upload/v1676850748/avatar/u0wa6m0xqrzceuopdawi.jpg"
	}
	modelNewUser.FullName = newUser.FullName
	modelNewUser.NameUser = newUser.NameUser
	modelNewUser.PasswordHash = passwordHash
	modelNewUser.Pais = newUser.Pais
	modelNewUser.Ciudad = newUser.Ciudad
	modelNewUser.Email = newUser.Email
	modelNewUser.KeyTransmission = "live" + KeyTransmission
	modelNewUser.Cmt = cmt
	modelNewUser.Timestamp = time.Now()
	modelNewUser.Likes = []primitive.ObjectID{}
	modelNewUser.Followers = make(map[primitive.ObjectID]domain.FollowInfo)
	modelNewUser.Following = make(map[primitive.ObjectID]domain.FollowInfo)
	modelNewUser.ClipsLikes = []primitive.ObjectID{}
	modelNewUser.Verified = false
	modelNewUser.Wallet = newUser.Wallet
	modelNewUser.Subscribers = []primitive.ObjectID{}
	modelNewUser.Subscriptions = []primitive.ObjectID{}
	modelNewUser.BirthDate = newUser.BirthDateTime
	modelNewUser.Clips = []primitive.ObjectID{}
	modelNewUser.Online = false
	modelNewUser.PanelAdminPinkker.Level = 0
	modelNewUser.PanelAdminPinkker.Asset = false
	modelNewUser.PanelAdminPinkker.Date = time.Now()
	modelNewUser.PanelAdminPinkker.Code = ""
	modelNewUser.Banned = false
	modelNewUser.CategoryPreferences = make(map[string]float64)

	return &modelNewUser
}

func (u *UserService) SaveUserRedis(newUser *domain.User) (string, error) {
	code, err := u.roomRepository.SaveUserRedis(newUser)
	return code, err
}
func (u *UserService) GetUserinRedis(code string) (*domain.User, error) {
	user, err := u.roomRepository.GetUserByCodeFromRedis(code)
	return user, err
}
func (u *UserService) SaveUser(newUser *domain.User) (primitive.ObjectID, error) {
	id, err := u.roomRepository.SaveUser(newUser)
	return id, err
}
func (u *UserService) CreateStream(newUser *domain.User, ID primitive.ObjectID) error {
	err := u.roomRepository.CreateStreamUser(newUser, ID)
	return err
}
func (u *UserService) SendConfirmationEmail(Email string, Token string) error {
	err := u.roomRepository.SendConfirmationEmail(Email, Token)
	return err
}

func (u *UserService) ConfirmationEmailToken(nameUser string) error {
	user, errFindUser := u.roomRepository.FindNameUser(nameUser, "")
	if errFindUser != nil {
		return errFindUser
	}
	errUpdateConfirmationEmailToken := u.roomRepository.UpdateConfirmationEmailToken(user)
	return errUpdateConfirmationEmailToken
}

// find
func (u *UserService) FindNameUser(NameUser string, Email string) (*domain.User, error) {

	user, err := u.roomRepository.FindNameUser(NameUser, Email)
	return user, err
}
func (u *UserService) FindNameUserNoSensitiveInformationApli(NameUser string, Email string) (*domain.GetUser, error) {

	user, err := u.roomRepository.FindNameUserNoSensitiveInformation(NameUser, Email)
	return user, err
}
func (u *UserService) GetUserByNameUserIndex(NameUser string) ([]*domain.GetUser, error) {
	NameUserLower := strings.ToLower(NameUser)

	user, err := u.roomRepository.GetUserByNameUserIndex(NameUserLower)
	return user, err
}
func (u *UserService) FindUserById(id primitive.ObjectID) (*domain.User, error) {
	user, err := u.roomRepository.FindUserById(id)
	return user, err
}

func (u *UserService) GetUserBykey(key string) (*domain.GetUser, error) {
	user, err := u.roomRepository.GetUserBykey(key)
	return user, err
}
func (u *UserService) GetUserByCmt(key string) (*domain.User, error) {
	user, err := u.roomRepository.GetUserByCmt(key)
	return user, err
}

// follow
func (u *UserService) FollowUser(IdUserTokenP primitive.ObjectID, IdUser primitive.ObjectID) error {
	err := u.roomRepository.FollowUser(IdUserTokenP, IdUser)

	return err
}
func (u *UserService) Unfollow(IdUserTokenP primitive.ObjectID, IdUser primitive.ObjectID) error {
	err := u.roomRepository.UnfollowUser(IdUserTokenP, IdUser)
	if err != nil {
		return err
	}
	return err
}
func (u *UserService) DeleteRedisUserChatInOneRoom(userToDelete primitive.ObjectID, IdRoom primitive.ObjectID) error {
	err := u.roomRepository.DeleteRedisUserChatInOneRoom(userToDelete, IdRoom)

	return err
}
func (u *UserService) GetWebSocketActivityFeed(user string) ([]*websocket.Conn, error) {
	client, err := u.roomRepository.GetWebSocketClientsInRoom(user)
	return client, err
}

// oauth2
func (u *UserService) FindEmailForOauth2Updata(user *domain.Google_callback_Complete_Profile_And_Username) (*domain.User, error) {
	userFind, err := u.roomRepository.FindEmailForOauth2Updata(user)
	return userFind, err
}

func (u *UserService) EditProfile(Profile domain.EditProfile, IdUserTokenP primitive.ObjectID) error {
	err := u.roomRepository.EditProfile(Profile, IdUserTokenP)
	return err
}
func (u *UserService) EditAvatar(avatarUrl string, IdUserTokenP primitive.ObjectID) error {

	err := u.roomRepository.EditAvatar(avatarUrl, IdUserTokenP)
	return err
}

func (u *UserService) EditPasswordHast(passwordHash string, id primitive.ObjectID) error {
	err := u.roomRepository.EditPasswordHast(passwordHash, id)
	return err
}

func (u *UserService) RedisSaveAccountRecoveryCode(code string, user domain.User) error {
	err := u.roomRepository.RedisSaveAccountRecoveryCode(code, user)
	return err
}
func (u *UserService) EditSocialNetworks(SocialNetwork userdomain.SocialNetwork, id primitive.ObjectID) error {
	err := u.roomRepository.EditSocialNetworks(SocialNetwork, id)
	return err
}
func (u *UserService) PanelAdminPinkkerInfoUser(dt userdomain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) (domain.User, streamdomain.Stream, error) {
	user, stream, err := u.roomRepository.PanelAdminPinkkerInfoUser(dt, id)
	return user, stream, err
}

func (u *UserService) PanelAdminPinkkerPartnerUser(dt userdomain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) error {
	err := u.roomRepository.PanelAdminPinkkerPartnerUser(dt, id)
	return err
}
func (u *UserService) PanelAdminPinkkerbanStreamer(dt userdomain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) error {
	err := u.roomRepository.PanelAdminPinkkerbanStreamer(dt, id)
	return err
}

func (u *UserService) PanelAdminRemoveBanStreamer(dt userdomain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) error {
	err := u.roomRepository.PanelAdminRemoveBanStreamer(dt, id)
	return err
}

func (u *UserService) CreateAdmin(CreateAdmin domain.CreateAdmin, id primitive.ObjectID) error {
	err := u.roomRepository.CreateAdmin(CreateAdmin, id)
	return err
}
