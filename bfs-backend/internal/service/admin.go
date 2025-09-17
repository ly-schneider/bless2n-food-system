package service

type AdminService interface {
}

type adminService struct {
}

func NewAdminService() AdminService {
	return &adminService{}
}
