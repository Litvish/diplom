package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Функция для добавления пациента в базу данных
func addPatientHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем информацию о враче из JWT
		claims, _ := c.Get("userInfo")
		userClaims, _ := claims.(*Claims)

		// Структура для приема данных из формы
		var patientData struct {
			Name    string `form:"name"`
			Surname string `form:"surname"`
			Disease string `form:"disease"`
		}

		// Привязываем данные формы к структуре
		if err := c.ShouldBind(&patientData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		attendingDoctor := userClaims.Surname + " " + userClaims.Name

		// SQL запрос для вставки данных
		query := `
			INSERT INTO patients (name, surname, disease, attending_doctor, admission_time)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`

		// Выполняем запрос
		var newPatientID int
		err := db.QueryRow(query, patientData.Name, patientData.Surname, patientData.Disease, attendingDoctor, time.Now()).Scan(&newPatientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении пациента: " + err.Error()})
			return
		}

		// Возвращаем успешный ответ с ID нового пациента
		c.JSON(http.StatusOK, gin.H{"message": "Пациент добавлен", "patient_id": newPatientID})
	}
}
