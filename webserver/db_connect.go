package main

import (
	"database/sql"
	"fmt"
)

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

// func addPatientHandler(c *gin.Context) {
// 	var patient Patient

// 	// Предполагаем, что данные приходят в формате JSON
// 	if err := c.BindJSON(&patient); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	// Вставка данных в базу данных
// 	query := `INSERT INTO patients (name, attending_doctor, disease) VALUES ($1, $2, $3)`
// 	_, err = db.Exec(query, patient.Name, patient.AttendingDoctor, patient.Disease)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"status": "success"})
// }
