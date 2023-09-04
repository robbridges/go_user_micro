package main

import (
	"net/http"
	"the_lonely_road/JWT"
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
		v.AddError("message", "User password must be 4 characters long and email must be 5 characters long")
		http.Error(w, v.Errors["message"], http.StatusBadRequest)
		return
	}

	err = app.userModel.Insert(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := JWT.GenerateJWT(int(user.ID))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the token in a cookie
	JWT.SetAuthCookie(w, token)

	app.writeJSON(w, http.StatusOK, &user)
}

func (app *App) getUserByEmail(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email string
	}
	v := validator.New()
	err := app.readJSON(w, r, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if models.ValidateEmail(v, payload.Email); !v.Valid() {
		v.AddError("message", "User password must be 4 characters long and email must be 5 characters long ")
		http.Error(w, v.Errors["message"], http.StatusBadRequest)
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
	v := validator.New()

	err := app.readJSON(w, r, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if models.ValidatePasswordPlaintext(v, payload.Password); !v.Valid() {
		v.AddError("message", "user password must be greater than 4 characters and less than 72")
		http.Error(w, v.Errors["message"], http.StatusBadRequest)
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
