package StreamSummaryapplication

import (
	StreamSummarydomain "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-domain"
	StreamSummaryinfrastructure "PINKKER-BACKEND/internal/StreamSummary/StreamSummary-infrastructure"
	"time"

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

func (s *StreamSummaryService) GetDailyEarningsForMonth(streamerID primitive.ObjectID, Time time.Time) ([]StreamSummarydomain.EarningsPerDay, error) {
	GetEarningsByRange, err := s.StreamSummaryRepository.GetDailyEarningsForMonth(streamerID, Time)
	return GetEarningsByRange, err
}
func (s *StreamSummaryService) GetEarningsByRange(streamerID primitive.ObjectID, startDate time.Time, endDate time.Time) (StreamSummarydomain.Earnings, error) {
	GetEarningsByRange, err := s.StreamSummaryRepository.GetEarningsByRange(streamerID, startDate, endDate)
	return GetEarningsByRange, err
}
func (s *StreamSummaryService) GetEarningsByDay(streamerID primitive.ObjectID, Time time.Time) (StreamSummarydomain.Earnings, error) {
	GetEarningsByDay, err := s.StreamSummaryRepository.GetEarningsByDay(streamerID, Time)
	return GetEarningsByDay, err
}
func (s *StreamSummaryService) GetEarningsByWeek(streamerID primitive.ObjectID, Time time.Time) (StreamSummarydomain.Earnings, error) {
	GetEarningsByWeek, err := s.StreamSummaryRepository.GetEarningsByWeek(streamerID, Time)
	return GetEarningsByWeek, err
}
func (s *StreamSummaryService) GetEarningsByMonth(streamerID primitive.ObjectID, Time time.Time) (StreamSummarydomain.Earnings, error) {
	GetEarningsByMonth, err := s.StreamSummaryRepository.GetEarningsByMonth(streamerID, Time)
	return GetEarningsByMonth, err
}
func (s *StreamSummaryService) GetEarningsByYear(streamerID primitive.ObjectID, Time time.Time) (StreamSummarydomain.Earnings, error) {
	GetEarningsByYear, err := s.StreamSummaryRepository.GetEarningsByYear(streamerID, Time)
	return GetEarningsByYear, err
}
func (s *StreamSummaryService) UpdateStreamSummary(StreamerID primitive.ObjectID, data StreamSummarydomain.UpdateStreamSummary) error {
	err := s.StreamSummaryRepository.UpdateStreamSummary(StreamerID, data)
	return err
}
func (s *StreamSummaryService) AddAds(idValueObj primitive.ObjectID, AddAds StreamSummarydomain.AddAds) error {
	err := s.StreamSummaryRepository.AddAds(idValueObj, AddAds)
	return err
}
func (s *StreamSummaryService) AverageViewers(Streamer primitive.ObjectID) error {
	err := s.StreamSummaryRepository.UpdateInfoStreamSummary(Streamer)
	return err
}
func (s *StreamSummaryService) GetLastSixStreamSummaries(Streamer primitive.ObjectID, date time.Time) ([]StreamSummarydomain.StreamSummary, error) {
	StreamSummarydomain, err := s.StreamSummaryRepository.GetLastSixStreamSummariesBeforeDate(Streamer, date)
	return StreamSummarydomain, err
}
func (s *StreamSummaryService) AWeekOfStreaming(Streamer primitive.ObjectID, page int) ([]StreamSummarydomain.StreamSummary, error) {
	StreamSummarydomain, err := s.StreamSummaryRepository.AWeekOfStreaming(Streamer, page)
	return StreamSummarydomain, err
}
func (s *StreamSummaryService) GeStreamSummaries(Streamer primitive.ObjectID) (*StreamSummarydomain.StreamSummaryGet, error) {
	GetStreamSummaryByID, err := s.StreamSummaryRepository.GetStreamSummaryByID(Streamer)
	return GetStreamSummaryByID, err
}

func (s *StreamSummaryService) GetStreamSummaryByTitle(title string) ([]StreamSummarydomain.StreamSummaryGet, error) {
	GetStreamSummaryByID, err := s.StreamSummaryRepository.GetStreamSummaryByTitle(title)
	return GetStreamSummaryByID, err
}
func (s *StreamSummaryService) GetStreamSummariesByStreamerIDLast30Days(Streamer primitive.ObjectID) ([]StreamSummarydomain.StreamSummaryGet, error) {
	GetStreamSummaryByID, err := s.StreamSummaryRepository.GetStreamSummariesByStreamerIDLast30Days(Streamer)
	return GetStreamSummaryByID, err
}
func (s *StreamSummaryService) GetTopVodsLast48Hours() ([]StreamSummarydomain.StreamSummaryGet, error) {
	GetStreamSummaryByID, err := s.StreamSummaryRepository.GetTopVodsLast48Hours()
	return GetStreamSummaryByID, err
}
