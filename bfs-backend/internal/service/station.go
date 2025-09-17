package service

type StationService interface {
}

type stationService struct {
}

func NewStationService() StationService {
	return &stationService{}
}
