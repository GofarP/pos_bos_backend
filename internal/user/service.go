package user

import (
	"context"
	"errors"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/validation"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type userService struct {
	repository domain.UserRepository
	validate   *validator.Validate
}

func NewUserService(repository domain.UserRepository) domain.UserService {
	return &userService{
		repository: repository,
		validate:   validator.New(),
	}
}

func (service *userService) AddUser(ctx context.Context, request domain.AddUserRequest) (*domain.User, error) {
	if err := service.validate.Struct(request); err != nil {
		return nil, validation.FormatValidationError(err)
	}

	existingUser, err := service.repository.GetByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &domain.User{
		Name:     request.Name,
		Email:    request.Email,
		Password: string(hashedPassword),
		Photo:    request.Photo,
	}

	err = service.repository.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *userService) GetByID(ctx context.Context, id int) (*domain.User, error) {
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (service *userService) GetAllUsers(ctx context.Context, request domain.PaginationRequest) (domain.PaginationResponse, error) {
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.Limit <= 0 {
		request.Limit = 10
	}

	users, total, err := service.repository.GetAll(ctx, request)
	if err != nil {
		return domain.PaginationResponse{}, err
	}

	totalPages := total / request.Limit
	if total%request.Limit > 0 {
		totalPages++
	}

	return domain.PaginationResponse{
		Data: users,
		Meta: domain.PaginationMeta{
			Page:         request.Page,
			Limit:        request.Limit,
			TotalRecords: total,
			TotalPages:   totalPages,
		},
	}, nil
}

func (service *userService) UpdateUser(ctx context.Context, id int, request domain.UpdateUserRequest) (*domain.User, error) {
	if err := service.validate.Struct(request); err != nil {
		return nil, validation.FormatValidationError(err)
	}

	user := &domain.User{
		Name:  request.Name,
		Email: request.Email,
		Photo: request.Photo,
	}

	if request.Password != nil && *request.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*request.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	err := service.repository.Update(ctx, id, user)
	if err != nil {
		return nil, err
	}

	user.ID = id
	return user, nil
}

func (service *userService) DeleteUser(ctx context.Context, id int) error {
	return service.repository.Delete(ctx, id)
}
