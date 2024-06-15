package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("123123") // Замените на ваш секретный ключ

type Claims struct {
	Surname string `json:"surname"`
	Name    string `json:"name"`
	jwt.StandardClaims
}

// HashPassword принимает пароль в виде строки и возвращает его хеш
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash принимает пароль и хеш и возвращает true, если они совпадают
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		// Логируем ошибку, если пароль не совпадает или произошла ошибка при сравнении
		log.Printf("bcrypt.CompareHashAndPassword failed: %v", err)
		return false
	}
	// Логируем успешное совпадение пароля
	log.Println("Password matched")
	return true
}

// Функция для проверки пароля пользователя
func CheckUserPassword(db *sql.DB, surname, name, password string) (bool, error) {
	var hashedPassword string

	// Выбираем хешированный пароль из базы данных
	query := `SELECT password FROM users WHERE surname = $1 AND name = $2`
	err := db.QueryRow(query, surname, name).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			// Пользователь не найден
			log.Printf("User with surname '%s' and name '%s' not found", surname, name)
			return false, nil
		}
		// Ошибка выполнения запроса
		log.Printf("Error querying user password: %v", err)
		return false, err
	}

	// Сравниваем предоставленный пароль с хешированным паролем из базы данных
	//log.Printf("Hashed password from DB: '%s'", hashedPassword)
	//log.Printf("Password being checked: '%s'", password)
	match := CheckPasswordHash(password, hashedPassword)
	if !match {
		// Логируем неудачную попытку входа
		log.Printf("Password does not match for user with surname '%s' and name '%s'", surname, name)
	} else {
		// Логируем успешную аутентификацию
		log.Printf("User with surname '%s' and name '%s' successfully authenticated", surname, name)
	}
	return match, nil
}

// GenerateJWT создает и подписывает новый JWT токен для пользователя
func GenerateJWT(surname, name string) (string, error) {
	// Устанавливаем время истечения токена
	expirationTime := time.Now().Add(50 * time.Minute)
	//log.Printf("Generating JWT for user: %s %s with expiration time: %v", surname, name, expirationTime)

	// Создаем новые утверждения (claims) с информацией о пользователе и временем истечения
	claims := &Claims{
		Surname: surname,
		Name:    name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Создаем новый JWT токен с утверждениями
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен с использованием секретного ключа
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		return "", err
	}

	// Возвращаем подписанный токен
	//log.Printf("JWT token generated: %s", tokenString)
	return tokenString, nil
}
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем токен из куки
		tokenString, err := c.Cookie("token")
		if err != nil {
			log.Println("Token cookie is missing: ", err)
			// Вместо отправки ошибки перенаправляем на страницу аутентификации
			c.Redirect(http.StatusTemporaryRedirect, "/auth")
			c.Abort() // Останавливаем цепочку обработчиков
			return
		}

		//log.Printf("Token string from cookie: %s", tokenString)

		// Инициализируем структуру для хранения данных токена
		claims := &Claims{}

		// Парсим и валидируем JWT
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		// Проверяем наличие ошибок при парсинге или недействительности токена
		if err != nil || !token.Valid {
			log.Printf("Error parsing token or token is invalid: %v", err)
			// Вместо отправки ошибки перенаправляем на страницу аутентификации
			c.Redirect(http.StatusTemporaryRedirect, "/auth")
			c.Abort() // Останавливаем цепочку обработчиков
			return
		}

		// Токен действителен, сохраняем информацию о пользователе в кxонтексте
		log.Printf("Authenticated user: %s %s", claims.Surname, claims.Name)
		c.Set("userInfo", claims)

		// Передаем управление следующему обработчику
		c.Next()
	}
}
