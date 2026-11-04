package main

import "database/sql"

// Struct with data
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
	IdUserFrom       int    `json:"idUserFrom"`
	PasswordUserFrom string `json:"passwordUserFrom"`
	IdUserTo         int    `json:"idUserTo"`
	Text             string `json:"text"`
	MessageId        int
}

type MessageRequest struct {
	IdUserFrom int `json:"idUserFrom"`
	IdUserTo   int `json:"idUserTo"`
	Limit      int `json:"limit"`
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
