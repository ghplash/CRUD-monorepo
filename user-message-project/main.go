package main

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Adding PostgreSQL Database
	dsn := "postgres://postgres:647tryGoewers@localhost:5432/user_message_db?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error with configuration of databse: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Error with connection db: %v", err)
	}

	log.Println("Database successfully work")
	// Main SQL query

	// Creating users table

	//_, err = db.Exec(`DROP TABLE users CASCADE`)
	//if err != nil {
	//	log.Fatalf("Error with deleting users table: %v", err)
	//}

	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		login VARCHAR NOT NULL
	)`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Error with creating Users table: %v", err)
	}

	log.Println("Users table was successfully created!")

	// Creating messages table

	//_, err = db.Exec(`DROP TABLE messages`)
	//if err != nil {
	//	log.Fatalf("Error with deleting messages table: %v", err)
	//}

	query = `
	CREATE TABLE IF NOT EXISTS messages (
		messageId SERIAL PRIMARY KEY,
		idUserFrom INT,
		idUserTo INT,
		text TEXT NOT NULL,
		when_send TIMESTAMP DEFAULT NOW(),
		FOREIGN KEY (idUserFrom) REFERENCES users (id) ON DELETE CASCADE,
		FOREIGN KEY (idUserTo) REFERENCES users (id) ON DELETE CASCADE
	)`
	if _, err = db.Exec(query); err != nil {
		log.Fatalf("Error with creating Messages table: %v", err)
	}

	log.Println("Messages table was successfully created!")

	// Processing of connections, Grouping, and Validation
	e := echo.New()
	user_messageDB := &DataBase{DB: db}
	user := e.Group("/users")
	{
		user.GET("/checkProfile/:id", user_messageDB.GETUser)
		user.POST("/createUser", user_messageDB.POSTUser)
		user.PATCH("/update_name/:id", user_messageDB.PATCHUserName)
		user.DELETE("/delete_user/:id", user_messageDB.DELETEUser)
	}

	messages := e.Group("/messages")
	{
		messages.GET("/checkHistory", user_messageDB.CheckMyMessageHistory)
		messages.POST("/sendMessage", user_messageDB.SENDMessage)
		messages.PATCH("/updateMessage/:messageId", user_messageDB.UPGRADEMessage)
		messages.DELETE("/deleteMessage/:messageId", user_messageDB.DELETEMessage)
	}

	e.Start(":5050")
}

// Other func

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

func checkLogin(login string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(login), []byte(password))
	return err == nil
}

func (db *DataBase) checkUserInDB(userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	var exist bool
	err := db.DB.QueryRowContext(context.Background(), query, userID).Scan(&exist)
	if err != nil {
		return false, err
	}
	return exist, nil
}
