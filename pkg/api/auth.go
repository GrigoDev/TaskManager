package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SignInRequest struct {
	Password string `json:"password"`
}

type SignInResponse struct {
	Token string `json:"token"`
}

type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

func signInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": "Неверный формат данных"})
		return
	}

	expectedPassword := os.Getenv("TODO_PASSWORD")
	if expectedPassword == "" {
		writeJSON(w, map[string]string{"error": "Аутентификация не настроена"})
		return
	}

	if req.Password != expectedPassword {
		writeJSON(w, map[string]string{"error": "Неверный пароль"})
		return
	}

	// Создаем хеш пароля
	hash := sha256.Sum256([]byte(expectedPassword))
	passwordHash := hex.EncodeToString(hash[:])

	// Создаем JWT токен
	claims := Claims{
		PasswordHash: passwordHash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(expectedPassword))
	if err != nil {
		writeJSON(w, map[string]string{"error": "Ошибка создания токена"})
		return
	}

	writeJSON(w, SignInResponse{Token: tokenString})
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, установлен ли пароль
		password := os.Getenv("TODO_PASSWORD")
		if len(password) == 0 {
			next(w, r)
			return
		}

		// Получаем токен из cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
			return
		}

		// Парсим и проверяем токен
		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(password), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
			return
		}

		// Проверяем хеш пароля
		claims, ok := token.Claims.(*Claims)
		if !ok {
			http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
			return
		}

		hash := sha256.Sum256([]byte(password))
		passwordHash := hex.EncodeToString(hash[:])
		if claims.PasswordHash != passwordHash {
			http.Error(w, "Требуется аутентификация", http.StatusUnauthorized)
			return
		}

		next(w, r)
	})
}
