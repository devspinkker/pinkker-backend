package StreamSummaryapplication

import (
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary.repository/StreamSummary-domain"
	StreamSummaryinfrastructure "PINKKER-BACKEND/internal/StreamSummary.repository/StreamSummary-infrastructure"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StreamSummaryService struct {
	StreamSummaryRepository *StreamSummaryinfrastructure.StreamSummaryRepository
}

func NewStreaSummaryService(StreamSummaryRepository *StreamSummaryinfrastructure.StreamSummaryRepository) *StreamSummaryService {
	return &StreamSummaryService{
		StreamSummaryRepository: StreamSummaryRepository,
	}
}
func (s *StreamSummaryService) UpdateStreamSummary(StreamerID primitive.ObjectID, data StreamSummarydomain.UpdateStreamSummary) error {
	err := s.StreamSummaryRepository.UpdateStreamSummary(StreamerID, data)
	return err
}
