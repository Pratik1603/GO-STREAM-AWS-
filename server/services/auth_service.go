package services

import (
	"errors"
	"os"
	"time"

	"server/myproject/models"
	"server/myproject/store"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getEnvDefault("JWT_SECRET", "your-secret-key-change-in-production"))

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type AuthService struct {
	store *store.Store
}

func NewAuthService(s *store.Store) *AuthService {
	return &AuthService{store: s}
}

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID int, role string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func (s *AuthService) Register(email, password, name, adminSecret string) (*models.User, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	role := "user"
	// Check if the provided secret matches the one in environment variables
	expectedSecret := os.Getenv("ADMIN_REGISTRATION_SECRET")
	if expectedSecret != "" && adminSecret == expectedSecret {
		role = "admin"
	}

	user, err := s.store.CreateUserWithRole(email, string(hash), name, role)
	if err != nil {
		return nil, "", err
	}

	token, err := GenerateJWT(user.ID, user.Role)
	return user, token, err
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
	user, err := s.store.GetUserByEmail(email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := GenerateJWT(user.ID, user.Role)
	return user, token, err
}

func (s *AuthService) ListUsers() ([]models.User, error) {
	return s.store.ListUsers()
}

func (s *AuthService) UpdateUserRole(id int, role string) error {
	return s.store.UpdateUserRole(id, role)
}

func (s *AuthService) DeleteUser(id int) error {
	return s.store.DeleteUser(id)
}
