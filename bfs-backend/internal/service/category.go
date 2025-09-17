package service

type CategoryService interface {
}

type categoryService struct {
}

func NewCategoryService() CategoryService {
	return &categoryService{}
}
