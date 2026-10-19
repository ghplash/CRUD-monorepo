package main

import (
	"database/sql"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"

	"github.com/go-playground/validator"
	_ "github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Adding PostgreSQL Database
	dsn := "postgres://postgres:647tryGoewers@localhost:5432/postgres"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error with configuration of databse", err)
	}
	defer db.Close()
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.GET("/getUsers", GETUser)
	e.POST("/postUsers", POSTUser)
	e.Start(":5050")
}

// Struct and slice with data
// User struct
type ReqisterData struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

type Users struct {
	ReqisterData
	ID int
}

type UserRequest struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

var userMap = make(map[int]Users)

// Message struct
type Message struct {
	IdUserFrom    int    `json:"idFrom"`
	LoginUserFrom string `json:"loginFrom"`
	IdUserTo      int    `json:"idTo"`
	Text          string `json:"text"`
}

var messagesMap = make(map[int]Message)

// Other struct
type CustomValidator struct {
	validator *validator.Validate
}

// Struct for answer

type AnswerToUsers struct {
	Users     map[int]Users
	You       Users
	YourLogin string
	YourID    int
	Message   string
}

type AnswerForErrors struct {
	Error string
}

// Main CRUD User func

func GETUser(c *echo.Context) error {
	var users Users
	if err := c.Bind(&users); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad JSON"})
	}
	if users.Role == "admin" {
		return c.JSON(http.StatusOK, AnswerToUsers{Users: userMap})
	} else {
		return c.JSON(http.StatusOK, AnswerToUsers{You: users})
	}
}

func POSTUser(c *echo.Context) error {
	var user Users
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad JSON"})
	}
	if err := c.Validate(&user); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: err.Error()})
	}
	id := rand.IntN(90000000)
	user.ID = id

	userLog, err := hashPassword(user.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Could not get a login"})
	}
	userMap[id] = user
	return c.JSON(http.StatusOK, AnswerToUsers{YourLogin: userLog, YourID: id, Message: "You were successfully register"})
}

func PATCHUserName(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad id"})
	}
	var upgUser UserRequest
	if err := c.Bind(&upgUser); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad JSON"})
	}

	user, exist := userMap[id]
	if !exist {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Error: "User not found, first, reqister"})
	}

	if ok := checkLogin(upgUser.Login, user.Password); !ok {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Incorrect login"})
	}

	user.Name = upgUser.Name
	return c.JSON(http.StatusOK, AnswerToUsers{You: user, Message: "Name was successfully updated"})
}

func DELETEUser(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad id"})
	}

	if _, exist := userMap[id]; !exist {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Error: "User not found, first, reqister"})
	}

	var delUser UserRequest
	if err := c.Bind(&delUser); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad JSON"})
	}
	delete(userMap, id)

	return c.JSON(http.StatusOK, AnswerToUsers{Message: "User was successfully deleted"})
}

//Main Messages func

func GETMessages(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad id"})
	}

	user, exist := userMap[id]
	if !exist {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Error: "No User with required ID"})
	}

	var userData UserRequest
	if err := c.Bind(&userData); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad JSON"})
	}

	if ok := checkLogin(userData.Login, user.Password); !ok {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad login"})
	}

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

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}
