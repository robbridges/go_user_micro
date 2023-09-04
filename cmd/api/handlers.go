package main

import (
	"net/http"
	"the_lonely_road/models"
	"the_lonely_road/validator"
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	v := validator.New()

	user := models.User{
		Email:     payload.Email,
		Password:  payload.Password,
		CreatedAt: time.Now(),
	}

	if models.ValidateUser(v, &user); !v.Valid() {
		http.Error(w, "user failed validation", http.StatusBadRequest)
		return
	}

	err = app.userModel.Insert(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.writeJSON(w, 200, &user)
}

func (app *App) getUserByEmail(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string
	}
	err := app.readJSON(w, r, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := app.userModel.GetByEmail(payload.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	app.writeJSON(w, 200, &user)
}

func (app *App) updateUserPassword(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		Email    string
		Password string
	}
	err := app.readJSON(w, r, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := app.userModel.GetByEmail(payload.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = app.userModel.UpdatePassword(int(user.ID), user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, 200, &user)
}
