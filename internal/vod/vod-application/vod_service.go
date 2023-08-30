package vodsapplication

import (
	voddomain "PINKKER-BACKEND/internal/vod/vod-domain"
	vodinfrastructure "PINKKER-BACKEND/internal/vod/vod-infrastructure"
)

type VodService struct {
	VodRepository *vodinfrastructure.VodRepository
}

func NewVodService(VodRepository *vodinfrastructure.VodRepository) *VodService {
	return &VodService{
		VodRepository: VodRepository,
	}
}

func (v *VodService) GetVodByStreamer(streamer string, limit string, sort string) ([]*voddomain.Vod, error) {
	Vods, err := v.VodRepository.GetVodsByStreamer(streamer, limit, sort)
	return Vods, err
}
func (v *VodService) GetVodWithId(vodId string) (*voddomain.Vod, error) {
	Vod, err := v.VodRepository.GetVodWithId(vodId)
	return Vod, err
}

func (v *VodService) CreateVod(url, streamKey string) error {
	user, err := v.VodRepository.GetUserByStreamKey(streamKey)
	if err != nil {
		return err
	}

	stream, err := v.VodRepository.GetStreamByStreamer(user.NameUser)
	if err != nil {
		return err
	}

	newVod := &voddomain.Vod{
		StreamerId:         user.ID,
		Streamer:           user.NameUser,
		URL:                url,
		StreamTitle:        stream.StreamTitle,
		StreamCategory:     stream.StreamCategory,
		StreamNotification: stream.StreamNotification,
		StreamTag:          stream.StreamTag,
		StreamIdiom:        stream.StreamIdiom,
		StreamThumbnail:    stream.StreamThumbnail,
		StartDate:          stream.StartDate,
	}

	return v.VodRepository.CreateVod(newVod)
}
