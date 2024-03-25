package streamapplication

import (
	streamdomain "PINKKER-BACKEND/internal/stream/stream-domain"
	streaminfrastructure "PINKKER-BACKEND/internal/stream/stream-infrastructure"

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

// get stream by id
func (s *StreamService) GetStreamById(id primitive.ObjectID) (*streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamById(id)
	return stream, err
}

// get stream by name user
func (s *StreamService) GetStreamByNameUser(nameUser string) (*streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamByNameUser(nameUser)
	return stream, err
}

// get streams by caregories
func (s *StreamService) GetStreamsByCategorie(Categorie string, page int) ([]streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamsByCategorie(Categorie, page)
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
func (s *StreamService) GetStreamsIdsStreamer(idsUsersF []primitive.ObjectID) ([]streamdomain.Stream, error) {
	stream, err := s.StreamRepository.GetStreamsIdsStreamer(idsUsersF)
	return stream, err
}

func (s *StreamService) Update_online(Key string, state bool) error {
	err := s.StreamRepository.UpdateOnline(Key, state)
	return err
}

func (s *StreamService) CloseStream(key string) error {
	err := s.StreamRepository.CloseStream(key)
	return err
}
func (s *StreamService) Update_thumbnail(cmt, image string) error {
	err := s.StreamRepository.Update_thumbnail(cmt, image)
	return err
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

func (s *StreamService) UpdateModChat(data streamdomain.UpdateModChat, id primitive.ObjectID) error {
	err := s.StreamRepository.UpdateModChat(data, id)
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
