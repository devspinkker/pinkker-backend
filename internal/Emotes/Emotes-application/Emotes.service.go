package Emotesapplication

import (
	Emotesdomain "PINKKER-BACKEND/internal/Emotes/Emotes"
	Emotesinfrastructure "PINKKER-BACKEND/internal/Emotes/Emotes-infrastructure"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmotesService struct {
	EmotesRepository *Emotesinfrastructure.EmotesRepository
}

func NewEmotesService(EmotesRepository *Emotesinfrastructure.EmotesRepository) *EmotesService {
	return &EmotesService{
		EmotesRepository: EmotesRepository,
	}
}

func (s *EmotesService) CreateEmote(emote Emotesdomain.Emote) (*Emotesdomain.Emote, error) {
	return s.EmotesRepository.CreateEmote(emote)
}

func (s *EmotesService) UpdateEmoteAut(emote Emotesdomain.EmoteUpdate, id primitive.ObjectID) (*Emotesdomain.Emote, error) {

	err := s.EmotesRepository.AutCode(id, emote.Code)
	if err != nil {
		return nil, err
	}
	return s.EmotesRepository.UpdateEmote(emote)
}

func (s *EmotesService) DeleteEmote(emoteID primitive.ObjectID) error {
	return s.EmotesRepository.DeleteEmote(emoteID)
}

func (s *EmotesService) GetEmote(emoteID primitive.ObjectID) (*Emotesdomain.Emote, error) {
	return s.EmotesRepository.GetEmote(emoteID)
}

func (s *EmotesService) GetAllEmotes() ([]Emotesdomain.Emote, error) {
	return s.EmotesRepository.GetAllEmotes()
}

func (s *EmotesService) ChangeEmoteTypeToGlobal(emoteID primitive.ObjectID) (*Emotesdomain.Emote, error) {

	return s.EmotesRepository.ChangeEmoteTypeToGlobal(emoteID)
}

func (s *EmotesService) ChangeEmoteTypeToPinkker(emoteID primitive.ObjectID) (*Emotesdomain.Emote, error) {
	return s.EmotesRepository.ChangeEmoteTypeToPinkker(emoteID)
}
func (s *EmotesService) GetPinkkerEmotes() ([]Emotesdomain.Emote, error) {
	return s.EmotesRepository.GetEmotesByType("pinkker")
}

// GetGlobalEmotes obtiene los emotes de tipo "global"
func (s *EmotesService) GetGlobalEmotes() ([]Emotesdomain.Emote, error) {
	return s.EmotesRepository.GetEmotesByType("global")
}
