package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Patient struct {
	ID              int
	Name            string
	AdmissionTime   time.Time
	GamesPlayed     []string
	AttendingDoctor string
	Disease         string
	AvatarURL       string
}

func initDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении к базе данных: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}
	var version string
	if err = db.QueryRow("SELECT version()").Scan(&version); err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	fmt.Println("Успешное подключение к базе данных! Версия сервера PostgreSQL:", version)
	return db, nil
}

func getPatients(db *sql.DB) ([]Patient, error) {
	var patients []Patient
	query := `SELECT id, name, disease, avatar_url FROM patients`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Patient
		if err := rows.Scan(&p.ID, &p.Name, &p.Disease, &p.AvatarURL); err != nil {
			return nil, err
		}
		patients = append(patients, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return patients, nil
}

func main() {
	connStr := "host=localhost port=5433 user=admin dbname=webapp password=a123123 sslmode=disable"
	db, err := initDB(connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := gin.Default()

	// Загрузка шаблонов
	router.LoadHTMLGlob("html/*.html")

	router.GET("/", func(c *gin.Context) {
		patients, err := getPatients(db)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "template.html", gin.H{
			"Patients": patients,
		})
	})

	// Обработчик для статических файлов
	router.Static("/static", "./static")

	// Запуск веб-сервера
	log.Println("Server is running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}