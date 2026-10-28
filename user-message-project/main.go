package main

import (
	"context"
	"database/sql"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"

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

	e.Start(":5050")
}

// Struct and slice with data
// User struct
type ReqisterData struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	ID       int
}

type UserRequest struct {
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Message struct

type SendMessage struct {
	idUserFrom int
	idUserTo   int
	text       string
}

// Other struct

type DataBase struct {
	DB *sql.DB
}

// Struct for answer

type ResponseForMessages struct {
	Message     string
	YourMessage SendMessage
}

type AnswerToUsers struct {
	YourName string
	YourID   int
	Message  string
}

type AnswerForErrors struct {
	Error   string
	Message string
}

// Main CRUD User func

func (db *DataBase) GETUser(c *echo.Context) error {
	if c.Request().Method != "GET" {
		return c.JSON(http.StatusMethodNotAllowed, AnswerForErrors{Message: "Can't be method exept GET"})
	}
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad id", Error: err.Error()})
	}

	var username string
	rows, err := db.DB.Query(`SELECT username FROM users WHERE id = $1`, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with DB!", Error: err.Error()})
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&username)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with find user in DB", Error: err.Error()})
		}
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, AnswerToUsers{YourName: username, YourID: id})
}

func (db *DataBase) POSTUser(c *echo.Context) error {
	if c.Request().Method != "POST" {
		return c.JSON(http.StatusMethodNotAllowed, AnswerForErrors{Message: "Can't be method exept POST"})
	}
	var regUser ReqisterData
	if err := c.Bind(&regUser); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad JSON", Error: err.Error()})
	}

	id := rand.IntN(99999999)
	regUser.ID = id

	userLog, err := hashPassword(regUser.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Could not get a login", Error: err.Error()})
	}
	_, err = db.DB.Exec(`INSERT INTO users (id, username, login) VALUES ($1,$2,$3)`, regUser.ID, regUser.Name, userLog)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with added user data in DB", Error: err.Error()})
	}
	return c.JSON(http.StatusOK, AnswerToUsers{YourName: regUser.Name, YourID: id, Message: "You were successfully register"})
}

func (db *DataBase) PATCHUserName(c *echo.Context) error {
	if c.Request().Method != "PATCH" {
		return c.JSON(http.StatusMethodNotAllowed, AnswerForErrors{Message: "Can't be method exept PATCH"})
	}
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad id", Error: err.Error()})
	}
	var upgUser UserRequest
	if err := c.Bind(&upgUser); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad JSON", Error: err.Error()})
	}

	if exist := db.checkUserInDB(id); !exist {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No user in DB with required ID,check it!"})
	}

	var userLog string
	err = db.DB.QueryRowContext(context.Background(), `SELECT login FROM users WHERE id = $1`, id).Scan(&userLog)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, AnswerForErrors{Message: "Error with password,check it!"})
	} else if err != nil {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "Error with password,check it!", Error: err.Error()})
	}

	if ok := checkLogin(userLog, upgUser.Password); !ok {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Incorrect password"})
	}

	_, err = db.DB.Exec(`UPDATE users SET username = $1 WHERE id = $2`, upgUser.Name, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with updating name", Error: err.Error()})
	}

	return c.JSON(http.StatusOK, AnswerToUsers{YourName: upgUser.Name, YourID: id, Message: "Name was successfully updated"})
}

func (db *DataBase) DELETEUser(c *echo.Context) error {
	if c.Request().Method != "DELETE" {
		return c.JSON(http.StatusMethodNotAllowed, AnswerForErrors{Message: "Can't be method exept DELETE"})
	}
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad id", Error: err.Error()})
	}

	if exist := db.checkUserInDB(id); !exist {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No user in DB with required ID,check it!"})
	}

	var delUser UserRequest
	if err := c.Bind(&delUser); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad JSON"})
	}

	var userLog string
	err = db.DB.QueryRowContext(context.Background(), `SELECT login FROM users WHERE id = $1`, id).Scan(&userLog)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "Error with DB in selection password!"})
	} else if err != nil {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "Error with DB", Error: err.Error()})
	}

	if ok := checkLogin(userLog, delUser.Password); !ok {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Incorrect password"})
	}

	query := `DELETE FROM users WHERE id = $1`
	if _, err := db.DB.Exec(query, id); err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with deleting name", Error: err.Error()})
	}
	return c.JSON(http.StatusOK, AnswerToUsers{Message: "User was successfully deleted"})
}

//Main Messages func

func (db *DataBase) GETMessages(c *echo.Context) error {
	if c.Request().Method != "GET" {
		return c.JSON(http.StatusMethodNotAllowed, AnswerForErrors{Message: "Can't be method exept GET"})
	}
	return c.JSON(http.StatusOK, AnswerToUsers{YourName: "YOOOOOO"})
}

func (db *DataBase) POSTMessage(c *echo.Context) error {
	if c.Request().Method != "POST" {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Only can be method POST"})
	}

	var newMessage SendMessage
	if err := c.Bind(&newMessage); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad JSON", Error: err.Error()})
	}

	existFrom := db.checkUserInDB(newMessage.idUserFrom)
	existTo := db.checkUserInDB(newMessage.idUserTo)
	if !existFrom || !existTo {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Cannot find users with required id, check it!"})
	}

	_, err := db.DB.Exec(`INSERT INTO messages (idUserFrom, idUserTo, text) VALUES ($1, $2, $3)`, newMessage.idUserFrom, newMessage.idUserTo, newMessage.text)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with added data in DB", Error: err.Error()})
	}

	return c.JSON(http.StatusOK, ResponseForMessages{Message: "Message was successfully sended!", YourMessage: newMessage})
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

func (db *DataBase) checkUserInDB(userID int) bool {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	var exist bool
	err := db.DB.QueryRowContext(context.Background(), query, userID).Scan(&exist)
	if err != nil {
		return false
	}
	return exist
}
