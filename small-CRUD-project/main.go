package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

func main() {
	e := echo.New()
	e.GET("/messages", GetHandler)
	e.POST("/messages", PostHandler)
	e.GET("getHi", GetHi)
	e.GET("/hello", GETHelloName)
	e.POST("/calculate", POSTCaclulate)
	e.GET("/tasks", GETTasks)
	e.POST("/tasks", POSTTasks)
	e.PATCH("/tasks/:id", PATCHTasks)
	e.DELETE("/tasks/:id", DELETETasks)
	e.GET("/books", GETBooks)
	e.POST("/books", POSTBooks)
	e.PATCH("/books/:id", PATCHBooks)
	e.DELETE("/books/:id", DELETEBooks)
	e.Start(":8080")
}

// Message
type Message struct {
	Text string `json:"text"`
	ID   int    `json:"id"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var messages = make(map[int]Message)
var nextID int = 1

func GetHi(c *echo.Context) error {
	return c.JSON(http.StatusOK, "Hi friend, have a nice day!")
}

func GetHandler(c *echo.Context) error {
	var msgSlice []Message
	for _, msg := range messages {
		msgSlice = append(msgSlice, msg)
	}
	return c.JSON(http.StatusOK, &msgSlice)
}

func PostHandler(c *echo.Context) error {
	var message Message
	if err := c.Bind(&message); err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Could not add the message"})
	}
	message.ID = nextID
	nextID++
	messages[message.ID] = message
	return c.JSON(http.StatusOK, Response{Status: "Success", Message: "Message was successfully added"})
}

// Practice
func GETHelloName(c *echo.Context) error {
	username := c.QueryParam("name")
	if username == "" {
		return c.JSON(http.StatusOK, Response{Status: "Success", Message: "Hello unknown"})
	} else {
		return c.JSON(http.StatusOK, Response{Status: "Success", Message: "Hello " + username})
	}
}

// Calculater

type Caclulater struct {
	Num1   int    `json:"num1"`
	Num2   int    `json:"num2"`
	Action string `json:"action"`
}

func POSTCaclulate(c *echo.Context) error {
	var result int
	var caclulate Caclulater
	if err := c.Bind(&caclulate); err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Bad request"})
	}
	num1 := caclulate.Num1
	num2 := caclulate.Num2
	action := caclulate.Action
	switch action {
	case "sum":
		result = num1 + num2
	case "sub":
		result = num1 - num2
	case "mul":
		result = num1 * num2
	case "div":
		if num2 == 0 {
			return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Could not div on zero"})
		} else {
			result = num1 / num2
		}
	}
	return c.JSON(http.StatusOK, Response{Status: "Success", Message: "result: " + strconv.Itoa(result)})
}

//To-Do-List

type Answer struct {
	Slise map[int]Tasks
}

var tasksMap = make(map[int]Tasks, 0)

type Tasks struct {
	Title       string `json:"title"`
	Is_complite bool   `json:"is_complite"`
}

func GETTasks(c *echo.Context) error {
	return c.JSON(http.StatusOK, Answer{Slise: tasksMap})
}

var taskID int

func POSTTasks(c *echo.Context) error {
	var task Tasks
	if err := c.Bind(&task); err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Could not add the tasks"})
	}
	tasksMap[taskID] = task
	taskID++
	return c.JSON(http.StatusOK, Response{Status: "Success", Message: "Task add in the map"})
}

func PATCHTasks(c *echo.Context) error {
	var tasks Tasks
	if err := c.Bind(&tasks); err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Bad JSON"})
	}

	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Bad ID"})
	}

	_, exist := tasksMap[id]
	if !exist {
		return c.JSON(http.StatusNotFound, Response{Status: "Not Found", Message: "No task with required ID"})
	}
	tasksMap[id] = tasks
	return c.JSON(http.StatusOK, Response{Status: "Success", Message: "Task was successfully updated"})
}

func DELETETasks(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Bad ID"})
	}

	if _, exist := tasksMap[id]; !exist {
		return c.JSON(http.StatusNotFound, Response{Status: "Not Found", Message: "No task with required ID"})
	}

	delete(tasksMap, id)
	return c.JSON(http.StatusOK, Response{Status: "Success", Message: "Task was successfully deleted"})
}

//Book catalog

type CatalogBooks struct {
	Title        string `json:"title"`
	Author       string `json:"author"`
	Is_available bool   `json:"is_available"`
}

var bookId int

var booksMap = make(map[int]CatalogBooks)

func GETBooks(c *echo.Context) error {
	return c.JSON(http.StatusOK, booksMap)
}

func POSTBooks(c *echo.Context) error {
	var catBook *CatalogBooks
	if err := c.Bind(&catBook); err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Bad request"})
	}

	booksMap[bookId] = *catBook
	bookId++

	return c.JSON(http.StatusOK, Response{Status: "Success", Message: "Book was successfully added"})
}

func PATCHBooks(c *echo.Context) error {
	var updBook *CatalogBooks
	if err := c.Bind(&updBook); err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Bad JSON"})
	}
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Bad ID"})
	}

	_, exist := booksMap[id]
	if !exist {
		return c.JSON(http.StatusNotFound, Response{Status: "Error", Message: "No book with the required ID in the catalog"})
	}
	booksMap[id] = *updBook

	return c.JSON(http.StatusOK, Response{Status: "Updated", Message: "Book was successfully updated"})
}

func DELETEBooks(c *echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Response{Status: "Error", Message: "Bad ID"})
	}

	_, exist := booksMap[id]
	if !exist {
		return c.JSON(http.StatusNotFound, Response{Status: "Error", Message: "No book with the required ID in the catalog"})
	}

	delete(booksMap, id)
	return c.JSON(http.StatusOK, Response{Status: "Delete", Message: "Book was successfully deleted"})
}
