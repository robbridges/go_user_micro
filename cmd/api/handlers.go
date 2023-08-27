package main

import (
	"net/http"
	"the_lonely_road/models"
	"time"
)

type jsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *App) HandleHome(w http.ResponseWriter, r *http.Request) {
	mockPayload := jsonPayload{
		Name: "User greet",
		Data: "Hello user",
	}
	app.writeJSON(w, http.StatusOK, &mockPayload)

}

func (app *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string
		Password string
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		w.Write([]byte("Error reading JSON"))
		return
	}
	passwordHash, err := models.EncryptPassword(payload.Password)
	user := models.User{
		Email:     payload.Email,
		Password:  passwordHash,
		CreatedAt: time.Now(),
	}
	err = app.userModel.Insert(&user)
	if err != nil {
		w.Write([]byte("Error inserting user " + err.Error()))
		return
	}
	app.writeJSON(w, 200, &user)
}
