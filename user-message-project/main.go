package main

import (
	"math/rand/v2"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	e := echo.New()
	e.GET("/getUsers", GETUser)
	e.POST("/postUsers", POSTUser)
}

// Struct and slice with data

type ReqisterData struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required,onof=admin user"`
}

type Users struct {
	UserInput ReqisterData
	ID        int
	Login     string
}

var userMap = make(map[int]Users)

// Struct for answer

type AnswerToUsers struct {
	Users map[int]Users
	You   Users
}

type AnswerForErrors struct {
	Error string
}

// Main CRUD func

func GETUser(c *echo.Context) error {
	var users Users
	if err := c.Bind(&users); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad JSON"})
	}
	if users.UserInput.Role == "admin" {
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
	if err := c.Validate(&user.UserInput); err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Role can only be 'user' or 'admin'"})
	}
	id := rand.IntN(90000000) + 100000000
	user.ID = id

	userLog, err := hashPassword(user.UserInput.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Could not get a login"})
	}
	user.Login = userLog
	userMap[id] = user
	return c.JSON(http.StatusOK, AnswerToUsers{You: user})
}

func PATCHUserName(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: "Bad id"})
	}
	user, exist := userMap[id]
	if !exist {
		return c.JSON(http.StatusNotFound, AnswerForErrors{Error: "User not found, first, reqister"})
	}
	login := c.Param("login")
	if !checkLogin(login, user.UserInput.Password) {
		return c.JSON(http.StatusBadRequest, AnswerForErrors{Error: ""})
	}
	name := c.Param("name")
	user.UserInput.Name = name
	return c.JSON(http.StatusOK, AnswerToUsers{You: user})
}

// Other func

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 16)
	return string(bytes), err
}

func checkLogin(login string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(login), []byte(password))
	return err == nil
}
