package main

import (
	"fmt"
	"net/http"
	"the_lonely_road/JWT"
	"the_lonely_road/errors"
	"the_lonely_road/models"
	"the_lonely_road/token"
	"the_lonely_road/validator"
	"time"
)

type jsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// more or less make sure the server can receieve requests when running
func (app *App) HandleHome(w http.ResponseWriter, r *http.Request) {
	mockPayload := jsonPayload{
		Name: "User greet",
		Data: "Hello user",
	}
	app.writeJSON(w, http.StatusOK, &mockPayload)

}

// create a user
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
		v.AddError("message", errors.InvalidUser)
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
		http.Error(w, errors.InternalServerError, http.StatusInternalServerError)
		return
	}

	// Set the token in a cookie
	JWT.SetAuthCookie(w, token)

	app.writeJSON(w, http.StatusOK, &user)
}

// get user by email
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
		v.AddError("message", errors.InvalidUser)
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

// update user password
func (app *App) updateUserPassword(w http.ResponseWriter, r *http.Request) {

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
		v.AddError("message", errors.InvalidUser)
		http.Error(w, v.Errors["message"], http.StatusBadRequest)
		return
	}
	user, err := app.userModel.GetByEmail(payload.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	passwordToken, salt, err := token.GenerateTokenAndSalt(32, 16)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	hashedToken := token.HashToken(passwordToken, salt)
	err = app.userModel.EnterPasswordHash(user.Email, hashedToken, salt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// send email with token
	err = app.emailer.ForgotPassword(user.Email, fmt.Sprintf("localhost:8080/users/password/reset?token=%s", passwordToken))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, 200, errors.PasswordResetEmail)
}

func (app *App) ProcessPasswordReset(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string
		Password string
	}
	passwordToken := r.URL.Query().Get("token")
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

	if user.PasswordResetExpiry.Before(time.Now()) {
		http.Error(w, errors.PasswordResetExpired, http.StatusBadRequest)
		return
	}

	ok := token.IsValidToken(passwordToken, user.PasswordResetHashToken, user.PasswordResetSalt)
	if !ok {
		http.Error(w, errors.InvalidToken, http.StatusBadRequest)
		return
	}

	err = app.userModel.UpdatePassword(int(user.ID), payload.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.userModel.ConsumePasswordReset(user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, 200, "Password updated successfully")

}

func (app *App) Authenticate(w http.ResponseWriter, r *http.Request) {
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
	if models.ValidateEmail(v, payload.Email); !v.Valid() {
		v.AddError("message", errors.InvalidUser)
		http.Error(w, v.Errors["message"], http.StatusBadRequest)
		return
	}

	user, err := app.userModel.Authenticate(payload.Email, payload.Password)
	if err != nil {
		http.Error(w, errors.InvalidCredentials, http.StatusBadRequest)
		return
	}

	token, err := JWT.GenerateJWT(int(user.ID))
	if err != nil {
		http.Error(w, errors.InternalServerError, http.StatusInternalServerError)
		return
	}

	// Set the token in a cookie
	JWT.SetAuthCookie(w, token)

	app.writeJSON(w, 200, &user)
}
