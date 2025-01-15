package userapplication

import (
	"PINKKER-BACKEND/config"
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	donationdomain "PINKKER-BACKEND/internal/donation/donation"
	notificationsdomain "PINKKER-BACKEND/internal/notifications/notifications"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	domain "PINKKER-BACKEND/internal/user/user-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	infrastructure "PINKKER-BACKEND/internal/user/user-infrastructure"
	"PINKKER-BACKEND/pkg/authGoogleAuthenticator"
	"PINKKER-BACKEND/pkg/helpers"
	"context"
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

func (u *UserService) GenerateTOTPKey(ctx context.Context, userID primitive.ObjectID, nameUser string) (string, string, error) {
	secret, url, err := authGoogleAuthenticator.GenerateKey(userID.Hex(), nameUser)
	if err != nil {
		return "", "", err
	}
	err = u.roomRepository.SetTOTPSecret(ctx, userID, secret)
	if err != nil {
		return "", "", err
	}
	return secret, url, nil
}

func (u *UserService) ValidateTOTPCode(ctx context.Context, userID primitive.ObjectID, code string) (bool, error) {
	return u.roomRepository.ValidateTOTPCode(ctx, userID, code)
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
		avatarConf := config.FotoPerfilAleatoria()
		modelNewUser.Avatar = avatarConf
	}

	modelNewUser.FullName = newUser.FullName
	modelNewUser.NameUser = newUser.NameUser
	modelNewUser.PasswordHash = passwordHash
	modelNewUser.Banner = "https://www.pinkker.tv/uploads/assets/banner_user.jpg"
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
	fifteenDaysAgo := time.Now().AddDate(0, 0, -15)
	modelNewUser.EditProfiile.Biography = fifteenDaysAgo
	modelNewUser.EditProfiile.NameUser = time.Now()
	modelNewUser.InCommunities = []primitive.ObjectID{}
	modelNewUser.OwnerCommunities = []primitive.ObjectID{}

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
func (u *UserService) RedisGetChangeGoogleAuthenticatorCode(code string) (*domain.User, error) {
	user, err := u.roomRepository.RedisGetChangeGoogleAuthenticatorCode(code)
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

func (u *UserService) IsUserBlocked(NameUser string) (bool, error) {

	user, err := u.roomRepository.IsUserBlocked(NameUser)
	return user, err
}
func (u *UserService) HandleLoginFailure(NameUser string) error {

	return u.roomRepository.HandleLoginFailure(NameUser)

}
func (u *UserService) FindNameUser(NameUser string, Email string) (*domain.User, error) {

	user, err := u.roomRepository.FindNameUser(NameUser, Email)
	return user, err
}
func (u *UserService) FindNameUserInternalOperation(NameUser string, Email string) (*domain.User, error) {

	user, err := u.roomRepository.FindNameUserInternalOperation(NameUser, Email)
	return user, err
}

func (u *UserService) FindNameUserNoSensitiveInformationApli(NameUser string, Email string) (*domain.GetUser, error) {

	user, err := u.roomRepository.FindNameUserNoSensitiveInformation(NameUser, Email)
	return user, err
}
func (u *UserService) GetStreamAndUserData(NameUser string, id primitive.ObjectID, nameUserToken string) (*streamdomain.Stream, *domain.GetUser, *domain.UserInfo, error) {

	return u.roomRepository.GetStreamAndUserData(NameUser, id, nameUserToken)
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
func (u *UserService) GetUserBanInstream(key string) (bool, error) {
	user, err := u.roomRepository.GetUserBanInstream(key)
	return user, err
}
func (u *UserService) GetUserByCmt(key string) (*domain.User, error) {
	user, err := u.roomRepository.GetUserByCmt(key)
	return user, err
}
func (u *UserService) UpdateLastConnection(IdUserTokenP primitive.ObjectID) error {
	return u.roomRepository.UpdateLastConnection(IdUserTokenP)

}

func (u *UserService) GetNotificacionesLastConnection(IdUserTokenP primitive.ObjectID, page int) ([]userdomain.FollowInfoRes, []donationdomain.ResDonation, []subscriptiondomain.ResSubscriber, error) {
	var GetRecentFollows []userdomain.FollowInfoRes
	var AllMyPixelesDonors []donationdomain.ResDonation
	var GetSubsChat []subscriptiondomain.ResSubscriber
	GetRecentFollows, _ = u.roomRepository.GetRecentFollowsLastConnection(IdUserTokenP, page)
	AllMyPixelesDonors, _ = u.roomRepository.AllMyPixelesDonorsLastConnection(IdUserTokenP, page)
	GetSubsChat, err := u.roomRepository.GetSubsChatLastConnection(IdUserTokenP, page)

	if err != nil && err.Error() != "no documents found" {
		return nil, nil, nil, err
	}
	return GetRecentFollows, AllMyPixelesDonors, GetSubsChat, nil
}

func (u *UserService) GetRecentNotis(IdUserTokenP primitive.ObjectID, page int) ([]userdomain.FollowInfoRes, []donationdomain.ResDonation, []subscriptiondomain.ResSubscriber, error) {
	var GetRecentFollows []userdomain.FollowInfoRes
	var AllMyPixelesDonors []donationdomain.ResDonation
	var GetSubsChat []subscriptiondomain.ResSubscriber
	GetRecentFollows, _ = u.roomRepository.GetRecentFollowsBeforeFirstConnection(IdUserTokenP, page)
	AllMyPixelesDonors, _ = u.roomRepository.AllMyPixelesDonorsBeforeFirstConnection(IdUserTokenP, page)
	GetSubsChat, err := u.roomRepository.GetSubsChatBeforeFirstConnection(IdUserTokenP, page)

	if err != nil && err.Error() != "no documents found" {
		return nil, nil, nil, err
	}
	return GetRecentFollows, AllMyPixelesDonors, GetSubsChat, nil
}
func (u *UserService) PurchasePinkkerPrime(IdUser primitive.ObjectID) (bool, error) {
	return u.roomRepository.PurchasePinkkerPrime(IdUser)
}

// follow
func (u *UserService) FollowUser(IdUserTokenP primitive.ObjectID, IdUser primitive.ObjectID) (string, error) {
	avatar, err := u.roomRepository.FollowUser(IdUserTokenP, IdUser)

	return avatar, err
}

func (u *UserService) Unfollow(IdUserTokenP primitive.ObjectID, IdUser primitive.ObjectID) error {
	err := u.roomRepository.UnfollowUser(IdUserTokenP, IdUser)
	if err != nil {
		return err
	}
	return err
}

func (u *UserService) IsFollowing(IdUserTokenP primitive.ObjectID, IdUser primitive.ObjectID) (bool, error) {
	return u.roomRepository.IsFollowing(IdUserTokenP, IdUser)

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
func (u *UserService) EditBanner(avatarUrl string, IdUserTokenP primitive.ObjectID) error {

	err := u.roomRepository.EditBanner(avatarUrl, IdUserTokenP)
	return err
}

func (u *UserService) EditPasswordHast(passwordHash string, id primitive.ObjectID) error {
	err := u.roomRepository.EditPasswordHast(passwordHash, id)
	return err
}
func (u *UserService) DeleteGoogleAuthenticator(id primitive.ObjectID) error {
	err := u.roomRepository.DeleteGoogleAuthenticator(id)
	return err
}

func (u *UserService) RedisSaveAccountRecoveryCode(code string, user domain.User) error {
	err := u.roomRepository.RedisSaveAccountRecoveryCode(code, user)
	return err
}

func (u *UserService) RedisSaveChangeGoogleAuthenticatorCode(code string, user domain.User) error {
	err := u.roomRepository.RedisSaveChangeGoogleAuthenticatorCode(code, user)
	return err
}

func (u *UserService) EditSocialNetworks(SocialNetwork userdomain.SocialNetwork, id primitive.ObjectID) error {
	err := u.roomRepository.EditSocialNetworks(SocialNetwork, id)
	return err
}
func (u *UserService) PanelAdminPinkkerInfoUser(dt userdomain.PanelAdminPinkkerInfoUserReq, id primitive.ObjectID) (*domain.GetUser, streamdomain.Stream, error) {
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
func (u *UserService) ChangeNameUserCodeAdmin(CreateAdmin domain.ChangeNameUser, id primitive.ObjectID) error {
	err := u.roomRepository.ChangeNameUserCodeAdmin(CreateAdmin, id)
	return err
}
func (u *UserService) ChangeNameUser(CreateAdmin domain.ChangeNameUser) error {
	err := u.roomRepository.ChangeNameUser(CreateAdmin)
	return err
}

func (u *UserService) GetRecommendedUsers(idT primitive.ObjectID, excludeIDs []primitive.ObjectID) ([]userdomain.GetUser, error) {
	limit := 5
	Users, err := u.roomRepository.GetRecommendedUsers(idT, excludeIDs, limit)
	return Users, err
}

func (u *UserService) GetAllPendingNameUserAds(page int, IdUser primitive.ObjectID) ([]advertisements.Advertisements, error) {
	ads, err := u.roomRepository.GetAllPendingNameUserAds(page, IdUser)
	if err != nil {
		return ads, err
	}
	return ads, err
}
func (u *UserService) AcceptOrDeleteAdvertisement(IdUserTokenP primitive.ObjectID, ad primitive.ObjectID, action bool) error {
	err := u.roomRepository.AcceptOrDeleteAdvertisement(IdUserTokenP, ad, action)
	if err != nil {
		return err
	}
	return err
}

// aa
func (u *UserService) GetAllAcceptedNameUserAds(page int, IdUser primitive.ObjectID) ([]advertisements.Advertisements, error) {
	return u.roomRepository.GetAllAcceptedNameUserAds(page, IdUser)

}
func (u *UserService) GetActiveAdsByEndAdCommunity(page int, IdUser primitive.ObjectID) ([]advertisements.Advertisements, error) {
	return u.roomRepository.GetActiveAdsByEndAdCommunity(page, IdUser)

}
func (u *UserService) GetAdsByNameUser(page int, IdUser primitive.ObjectID, nameUser string) ([]advertisements.Advertisements, error) {
	return u.roomRepository.GetAdsByNameUser(page, IdUser, nameUser)

}
func (u *UserService) SaveNotification(userID primitive.ObjectID, notification notificationsdomain.Notification) error {
	err := u.roomRepository.SaveNotification(userID, notification)
	return err
}
func (u *UserService) UpdatePinkkerProfitPerMonthRegisterLinkReferent(source string) error {
	if source != "ig" || source != "fb" {
		return nil
	}
	return u.roomRepository.UpdatePinkkerProfitPerMonthRegisterLinkReferent(source)
}
