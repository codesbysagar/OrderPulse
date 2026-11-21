package auth

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/codesbysagar/OrderPulse/internal/domain"
	"github.com/codesbysagar/OrderPulse/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("username is already taken")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrWeakInput          = errors.New("username must be at least 3 chars and password at least 4 chars")
)

type AuthService struct {
	repo *storage.TicketRepository
}

func NewAuthService(repo *storage.TicketRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) Register(username, password string) (*domain.User, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	if len(username) < 3 || len(password) < 4 {
		return nil, ErrWeakInput
	}

	if existing, _ := s.repo.GetUserByUsername(username); existing != nil {
		return nil, ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(username, password string) (*domain.User, error) {
	username = strings.TrimSpace(strings.ToLower(username))

	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
