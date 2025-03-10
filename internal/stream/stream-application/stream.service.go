package streamapplication

import (
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-domain"
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	streaminfrastructure "PINKKER-BACKEND/internal/stream/stream-infrastructure"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"PINKKER-BACKEND/pkg/utils"

	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StreamService struct {
	StreamRepository *streaminfrastructure.StreamRepository
}

func NewStreamService(StreamRepository *streaminfrastructure.StreamRepository) *StreamService {
	return &StreamService{
		StreamRepository: StreamRepository,
	}
}
func (s *StreamService) CategoriesUpdate(req streamdomain.CategoriesUpdate, idUser primitive.ObjectID) error {
	err := s.StreamRepository.CategoriesUpdate(req, idUser)
	return err
}

// get stream by StreamerID
func (s *StreamService) GetStreamById(id primitive.ObjectID) (*streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamById(id)
	return stream, err
}
func (s *StreamService) CommercialInStreamSelectAdvertisements(StreamCategory string, ViewerCount int) (advertisements.Advertisements, error) {
	Advertisements, err := s.StreamRepository.CommercialInStreamSelectAdvertisements(StreamCategory, ViewerCount)
	return Advertisements, err
}
func (s *StreamService) GetStreamSummaryById(id primitive.ObjectID) (*StreamSummarydomain.StreamSummary, error) {
	StreamSummary, err := s.StreamRepository.GetStreamSummaryById(id)
	return StreamSummary, err
}
func (s *StreamService) Get(id primitive.ObjectID) (*streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamById(id)
	return stream, err
}

// get stream by name user
func (s *StreamService) GetStreamByNameUser(nameUser string) (*streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamByNameUser(nameUser)
	return stream, err
}

// get streams by caregories
func (s *StreamService) GetStreamsByCategorie(Category string, page int) ([]streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamsByCategory(Category, page)
	return stream, err
}

func (s *StreamService) GetAllsStreamsOnline(page int) ([]streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetAllsStreamsOnline(page)
	return stream, err
}
func (s *StreamService) GetStreamsMostViewed(page int) ([]streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamsMostViewed(page)
	return stream, err
}
func (s *StreamService) GetAllsStreamsOnlineThatUserFollows(idValueObj primitive.ObjectID) ([]streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetAllStreamsOnlineThatUserFollows(idValueObj)
	return stream, err
}

func (s *StreamService) RecommendationStreams(page int) ([]streamdomain.Stream, error) {
	limit := 10
	stream, err := s.StreamRepository.RecommendStreams(limit, page)
	return stream, err
}

func (s *StreamService) GetStreamsIdsStreamer(idsUsersF []primitive.ObjectID) ([]streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamsIdsStreamer(idsUsersF)
	return stream, err
}

func (s *StreamService) Update_online(Key string, state bool) (primitive.ObjectID, error) {
	LastStreamSummary, err := s.StreamRepository.UpdateOnline(Key, state)
	return LastStreamSummary, err
}

func (s *StreamService) CloseStream(key string) error {
	err := s.StreamRepository.CloseStream(key)
	return err
}
func (s *StreamService) Update_thumbnail(cmt, image string) error {
	err := s.StreamRepository.Update_thumbnail(cmt, image)
	return err
}
func (s *StreamService) GetWebSocketClientsInRoom(roomID string) ([]*websocket.Conn, error) {
	clients, err := utils.NewChatService().GetWebSocketClientsInRoom(roomID)

	return clients, err
}
func (s *StreamService) Streamings_online() (int, error) {
	online, err := s.StreamRepository.Streamings_online()
	return online, err
}

func (s *StreamService) Update_start_date(req streamdomain.Update_start_date) error {
	err := s.StreamRepository.Update_start_date(req)
	return err
}
func (s *StreamService) UpdateStreamInfo(data streamdomain.UpdateStreamInfo, id primitive.ObjectID) error {
	err := s.StreamRepository.UpdateStreamInfo(data, id)
	return err
}
func (s *StreamService) UpdateAntiqueStreamDuration(data streamdomain.UpdateAntiqueStreamDuration, id primitive.ObjectID) error {
	err := s.StreamRepository.UpdateAntiqueStreamDuration(id, data)
	return err
}
func (s *StreamService) UpdateChatRulesStream(data streamdomain.ChatRulesReq, id primitive.ObjectID) error {
	err := s.StreamRepository.UpdateChatRulesStream(id, data)
	return err
}
func (s *StreamService) UpdateModChat(data streamdomain.UpdateModChat, id primitive.ObjectID) error {
	err := s.StreamRepository.UpdateModChat(data, id)
	return err
}
func (s *StreamService) UpdateModChatSlowMode(data streamdomain.UpdateModChatSlowMode, id primitive.ObjectID) error {
	err := s.StreamRepository.UpdateModChatSlowMode(data, id)
	return err
}

func (s *StreamService) Update_Emotes(idUser primitive.ObjectID, date int) error {
	err := s.StreamRepository.Update_Emotes(idUser, date)
	return err
}
func (s *StreamService) GetCategories() ([]streamdomain.Categoria, error) {
	Categorias, err := s.StreamRepository.GetCategories()
	return Categorias, err
}

func (s *StreamService) GetCategoria(Cate string) (streamdomain.Categoria, error) {
	Categorias, err := s.StreamRepository.GetCategia(Cate)
	return Categorias, err
}

func (s *StreamService) ValidateStreamAccess(idUser, idStreamer primitive.ObjectID) (bool, error) {
	return s.StreamRepository.ValidateStreamAccess(idUser, idStreamer)
}
func (s *StreamService) GetInfoUserInRoomBaneados(nameuser string, nameUserToken string) ([]*userdomain.UserInfo, error) {
	return s.StreamRepository.GetInfoUsersInRoom(nameuser, nameUserToken)
}
