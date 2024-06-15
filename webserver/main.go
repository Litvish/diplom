package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Patient struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Surname         string `json:"surname"`
	Disease         string `json:"disease"`
	AdmissionTime   string `json:"admission_time"` // Теперь это строка
	AttendingDoctor string `json:"attending_doctor"`
}

func main() {
	connStr := "host=localhost port=5433 user=admin dbname=webapp password=a123123	 sslmode=disable"
	db, err := initDB(connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := gin.Default()

	// Загрузка шаблонов
	router.LoadHTMLGlob("html/*.html")

	router.GET("/", AuthMiddleware(), func(c *gin.Context) {
		Claims := c.MustGet("userInfo").(*Claims)

		patients, err := getPatients(db)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "template.html", gin.H{
			"UserSurname": Claims.Surname,
			"UserName":    Claims.Name,
			"Patients":    patients,
			"content":     "main",
			"Title":       "Dashboard",
		})
	})
	router.GET("/auth", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth.html", nil)
	})

	router.POST("/auth", func(c *gin.Context) {
		var loginCreds struct {
			Surname  string `form:"surname"`
			Name     string `form:"name"`
			Password string `form:"password"`
		}
		if err := c.ShouldBind(&loginCreds); err != nil {
			c.HTML(http.StatusBadRequest, "auth.html", gin.H{"error": "Invalid form input"})
			return
		}

		// Проверяем учетные данные пользователя
		match, err := CheckUserPassword(db, loginCreds.Surname, loginCreds.Name, loginCreds.Password)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "auth.html", gin.H{"error": "Неверные данные"})

			return
		}

		if !match {
			c.HTML(http.StatusUnauthorized, "auth.html", gin.H{"error": "Неверные данные"})
			return
		}

		// Учетные данные верны, генерируем JWT токен
		tokenString, err := GenerateJWT(loginCreds.Surname, loginCreds.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Устанавливаем токен в куки (или заголовок, если нужно)
		c.SetCookie("token", tokenString, 3600, "/", "", false, true)

		// Перенаправляем пользователя на главную страницу
		c.Redirect(http.StatusFound, "/")
	})
	router.GET("/game", AuthMiddleware(), func(c *gin.Context) {

		c.HTML(http.StatusOK, "template.html", gin.H{
			"Title":   "Game Page",
			"Content": "game",
		})
	})
	router.GET("/patient", AuthMiddleware(), func(c *gin.Context) {
		patients, err := getPatients(db)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "template.html", gin.H{
			"Title":    "patients",
			"Content":  "table",
			"Patients": patients,
		})
	})
	router.GET("/add-patient", AuthMiddleware(), func(c *gin.Context) {

		c.HTML(http.StatusOK, "template.html", gin.H{
			"Title":   "добавить пациента",
			"Content": "add",
		})
	})
	router.POST("/add-patient", AuthMiddleware(), addPatientHandler(db))

	// Обработчик для статических файлов
	router.Static("/static", "./static")

	// Запуск веб-сервера
	log.Println("Server is running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
