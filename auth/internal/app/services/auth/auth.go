package auth

import (
	"auth/internal/app/constants"
	"auth/internal/app/models"
	"auth/pkg/cache"
	"auth/pkg/db"
	"auth/pkg/logger"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Service struct {
	AccessSecret  string
	RefreshSecret string
	Pepper        string
}

func New(accessSecret, refreshSecret, pepper string) *Service {
	return &Service{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		Pepper:        pepper,
	}
}

func (s Service) Login(user models.User) (string, string, error) {
	rows, err := db.Conn.Query(`SELECT password, salt FROM users WHERE login = $1`, user.Login)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	var passwordHash, salt string
	if rows.Next() {
		err = rows.Scan(&passwordHash, &salt)
		if err != nil {
			return "", "", err
		}
	}
	if passwordHash == "" {
		return "", "", constants.ErrUserNotFound
	}

	if !s.checkPasswordHash(user.Password, fmt.Sprintf("%s%s%s", user.Password, salt, s.Pepper)) {
		return "", "", errors.New("invalid password")
	}

	data := map[string]interface{}{
		"login":     user.Login,
		"device_id": user.DeviceID,
	}
	accessTokenString, refreshTokenString, err := s.generateAndCacheTokens(data)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (s Service) Register(user models.User) (string, string, error) {
	var exists bool
	err := db.Conn.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE login=$1)`, user.Login).Scan(&exists)
	if err != nil {
		return "", "", err
	} else if exists {
		return "", "", constants.ErrUserExists
	}

	salt, err := s.generateSalt(32)
	if err != nil {
		return "", "", err
	}

	passwordHash, err := s.hashPassword(fmt.Sprintf("%s%s%s", user.Password, salt, s.Pepper))
	if err != nil {
		return "", "", err
	}

	rows, err := db.Conn.Query(`INSERT INTO users (login, password, salt) VALUES ($1, $2, $3)`, user.Login, passwordHash, salt)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	data := map[string]interface{}{
		"login":     user.Login,
		"device_id": user.DeviceID,
	}
	accessTokenString, refreshTokenString, err := s.generateAndCacheTokens(data)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (s Service) ValidateToken(accessTokenString string) (bool, error) {
	accessToken, err := s.parseToken(accessTokenString, s.AccessSecret)
	if err != nil {
		return false, err
	}

	if _, ok := accessToken.Claims.(jwt.MapClaims); ok && accessToken.Valid {
		return true, nil
	}

	return false, errors.New("неверный access токен")
}

// переписать логику, эту хуйню какой то долбаеб под героином писал
func (s Service) RefreshToken(refreshTokenString string) (string, string, error) {
	refreshToken, err := s.parseToken(refreshTokenString, s.RefreshSecret)
	if err != nil {
		return "", "", err
	}

	if claims, ok := refreshToken.Claims.(jwt.MapClaims); ok && refreshToken.Valid {
		refreshHashKey := fmt.Sprintf("token:refresh:%s", refreshTokenString)

		cacheRefreshTokenString, err := cache.ReadCache(refreshHashKey)
		if err != nil || cacheRefreshTokenString == "false" {
			rows, err := db.Conn.Query(`UPDATE refresh_tokens SET expires_at = $1 WHERE device_id = $2`, time.Now(), claims["device_id"].(string))
			if err != nil {
				return "", "", err
			}
			defer rows.Close()

			return "", "", errors.New("refresh токен не совпадает с действительным")
		}

		data := map[string]interface{}{
			"login":     claims["login"],
			"device_id": claims["device_id"],
		}
		accessTokenString, newRefreshTokenString, err := s.generateAndCacheTokens(data)
		if err != nil {
			return "", "", err
		}

		return accessTokenString, newRefreshTokenString, nil
	}

	return "", "", errors.New("неверный refresh токен")
}

func (s Service) parseToken(tokenString, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	return token, err
}

func (s Service) generateToken(data map[string]interface{}, secret string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login":     data["login"],
		"device_id": data["device_id"],
		"exp":       time.Now().Add(duration).Unix(),
	})

	return token.SignedString([]byte(secret))
}

func (s Service) generateAndCacheTokens(data map[string]interface{}) (string, string, error) {
	accessTokenString, err := s.generateToken(data, s.AccessSecret, time.Hour*1)
	if err != nil {
		return "", "", err
	}

	refreshTokenString, err := s.generateToken(data, s.RefreshSecret, time.Hour*720)
	if err != nil {
		return "", "", err
	}

	refreshHashKey := fmt.Sprintf("token:refresh:%s", refreshTokenString)

	err = cache.SaveCache(refreshHashKey, true, time.Hour*720)
	if err != nil {
		logger.Error("Ошибка при сохранении нового Refresh токена в Redis")
	}

	return accessTokenString, refreshTokenString, nil
}

func (s Service) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s Service) checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}

func (s Service) generateSalt(size int) (string, error) {
	salt := make([]byte, size)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(salt), nil
}
