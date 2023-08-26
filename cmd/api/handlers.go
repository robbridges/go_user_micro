package main

import "net/http"

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
