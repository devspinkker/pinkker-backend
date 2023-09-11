package userapplication

import (
	domain "PINKKER-BACKEND/internal/user/user-domain"
	infrastructure "PINKKER-BACKEND/internal/user/user-infrastructure"
	"PINKKER-BACKEND/pkg/helpers"
	"strings"
	"time"

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

// signup
func (u *UserService) SaveUser(newUser *domain.UserModelValidator, avatarUrl string, passwordHash string) error {
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
	modelNewUser.Followers = []primitive.ObjectID{}
	modelNewUser.Following = []primitive.ObjectID{}
	modelNewUser.Verified = false
	modelNewUser.Wallet = newUser.Wallet
	modelNewUser.Subscribers = []domain.Subscriber{}
	modelNewUser.Subscriptions = []domain.Subscription{}

	id, err := u.roomRepository.SaveUserDB(&modelNewUser)
	err = u.roomRepository.CreateStreamUser(&modelNewUser, id)
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
	NameUserLower := strings.ToLower(NameUser)

	user, err := u.roomRepository.FindNameUser(NameUserLower, Email)
	return user, err
}
func (u *UserService) FindUserById(id primitive.ObjectID) (*domain.User, error) {
	user, err := u.roomRepository.FindUserById(id)
	return user, err
}

func (u *UserService) GetUserBykey(key string) (*domain.User, error) {
	user, err := u.roomRepository.GetUserBykey(key)
	return user, err
}

// follow
func (u *UserService) FollowUser(IdUserTokenP primitive.ObjectID, IdUser primitive.ObjectID) error {
	err := u.roomRepository.FollowUser(IdUserTokenP, IdUser)
	return err
}
func (u *UserService) Unfollow(IdUserTokenP primitive.ObjectID, IdUser primitive.ObjectID) error {
	err := u.roomRepository.UnfollowUser(IdUserTokenP, IdUser)
	return err
}
