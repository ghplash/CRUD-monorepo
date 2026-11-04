package main

import (
	"database/sql"
	"math/rand/v2"
	"net/http"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v5"
)

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

	return c.JSON(http.StatusOK, AnswerToUsers{YourName: username, YourID: id, Message: "Here is your profile"})
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

	if len(regUser.Password) <= 10 || len(regUser.Password) >= 20 {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "The password length cannot be less than 12 or more than 20"})
	}
	userLog, err := hashPassword(regUser.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Could not create a login", Error: err.Error()})
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

	exist, err := db.checkUserInDB(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error in DB with finding user", Error: err.Error()})
	} else if !exist {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No user in DB with required Id"})
	}
	var userLog string
	err = db.DB.QueryRowContext(c.Request().Context(), `SELECT login FROM users WHERE id = $1`, id).Scan(&userLog)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No password in DB, error no rows"})
	} else if err != nil {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No password in DB", Error: err.Error()})
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

	var delUser UserRequest
	if err := c.Bind(&delUser); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad JSON"})
	}

	exist, err := db.checkUserInDB(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error in DB with finding user", Error: err.Error()})
	} else if !exist {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No user in DB with required Id"})
	}

	var userLog string
	err = db.DB.QueryRowContext(c.Request().Context(), `SELECT login FROM users WHERE id = $1`, id).Scan(&userLog)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No password in DB, error no rows"})
	} else if err != nil {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No password in DB", Error: err.Error()})
	}

	if ok := checkLogin(userLog, delUser.Password); !ok {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Incorrect password"})
	}

	query := `DELETE FROM users WHERE id = $1`
	if _, err := db.DB.Exec(query, id); err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with deleting user", Error: err.Error()})
	}
	return c.JSON(http.StatusOK, AnswerToUsers{Message: "User was successfully deleted"})
}

//Main Messages func

func (db *DataBase) CheckMyMessageHistory(c *echo.Context) error {
	if c.Request().Method != "GET" {
		return c.JSON(http.StatusMethodNotAllowed, AnswerForErrors{Message: "Can't be method exept GET"})
	}

	var messageReq MessageRequest
	if err := c.Bind(&messageReq); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad JSON"})
	}

	existFrom, err := db.checkUserInDB(messageReq.IdUserFrom)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error in DB with finding user", Error: err.Error()})
	} else if !existFrom {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No userFrom in DB with required Id"})
	}
	existTo, err := db.checkUserInDB(messageReq.IdUserTo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error in DB with finding user", Error: err.Error()})
	} else if !existTo {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No userTo in DB with required Id"})
	}

	rows, err := db.DB.Query(`
	SELECT text FROM messages WHERE idUserFrom = $1 AND idUserTo = $2
	ORDER BY when_send DESC
	LIMIT $3`, messageReq.IdUserFrom, messageReq.IdUserTo, messageReq.Limit)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with selecting messages, error no rows"})
		}
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with selecting messages", Error: err.Error()})
	}
	defer rows.Close()
	var messages []string
	for rows.Next() {
		var text string
		err := rows.Scan(&text)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error witn scanning messages", Error: err.Error()})
		}
		messages = append(messages, text)
	}
	if err = rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, messages)
}

func (db *DataBase) SENDMessage(c *echo.Context) error {
	if c.Request().Method != "POST" {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Only can be method POST"})
	}

	var newMessage SendMessage
	if err := c.Bind(&newMessage); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad JSON", Error: err.Error()})
	}

	existFrom, err := db.checkUserInDB(newMessage.IdUserFrom)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error in DB with finding user", Error: err.Error()})
	} else if !existFrom {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No userFrom in DB with required Id"})
	}
	existTo, err := db.checkUserInDB(newMessage.IdUserTo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error in DB with finding user", Error: err.Error()})
	} else if !existTo {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No userTo in DB with required Id"})
	}

	var loginFrom string
	err = db.DB.QueryRowContext(c.Request().Context(), `SELECT login FROM users WHERE id = $1`, newMessage.IdUserFrom).Scan(&loginFrom)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No password in DB, error no rows"})
		} else {
			return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No password in DB", Error: err.Error()})
		}
	}

	if ok := checkLogin(loginFrom, newMessage.PasswordUserFrom); !ok {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Incorect password!"})
	}

	err = db.DB.QueryRowContext(c.Request().Context(), `
	INSERT INTO messages (idUserFrom, idUserTo, text)
	VALUES ($1, $2, $3)
	RETURNING messageId`, newMessage.IdUserFrom, newMessage.IdUserTo, newMessage.Text).Scan(&newMessage.MessageId)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with returning messageId", Error: err.Error()})
	}

	return c.JSON(http.StatusOK, ResponseForMessages{Message: "Message was successfully sended!", YourMessage: newMessage})
}

func (db *DataBase) UPGRADEMessage(c *echo.Context) error {
	if c.Request().Method != "PATCH" {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Only can be method POST"})
	}

	idParam := c.Param("messageId")
	messageId, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad ID"})
	}

	var newMessage SendMessage
	if err := c.Bind(&newMessage); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad JSON", Error: err.Error()})
	}

	existFrom, err := db.checkUserInDB(newMessage.IdUserFrom)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error in DB with finding user", Error: err.Error()})
	} else if !existFrom {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No userFrom in DB with required Id"})
	}
	existTo, err := db.checkUserInDB(newMessage.IdUserTo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error in DB with finding user", Error: err.Error()})
	} else if !existTo {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No userTo in DB with required Id"})
	}

	var loginFrom string
	err = db.DB.QueryRowContext(c.Request().Context(), `SELECT login FROM users WHERE id = $1`, newMessage.IdUserFrom).Scan(&loginFrom)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No password in DB, error no rows"})
		} else {
			return c.JSON(http.StatusNotFound, AnswerForErrors{Message: "No password in DB", Error: err.Error()})
		}
	}

	if ok := checkLogin(loginFrom, newMessage.PasswordUserFrom); !ok {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Incorect password!"})
	}

	_, err = db.DB.Exec(`UPDATE messages SET text = $1, when_send = NOW() WHERE messageId = $2`, newMessage.Text, messageId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with updating message!", Error: err.Error()})
	}

	return c.JSON(http.StatusOK, ResponseForMessages{Message: "Message was successfully updated!", YourMessage: newMessage})
}

func (db *DataBase) DELETEMessage(c *echo.Context) error {
	if c.Request().Method != "DELETE" {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Only can be method DELETE"})
	}

	idParam := c.Param("messageId")
	messageId, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Message: "Bad ID"})
	}

	_, err = db.DB.Exec(`DELETE FROM messages WHERE messageId = $1`, messageId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, AnswerForErrors{Message: "Error with deleting user", Error: err.Error()})
	}

	return c.JSON(http.StatusOK, AnswerToUsers{Message: "Message was successfully deleted"})

}