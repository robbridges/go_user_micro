package main

import (
	"fmt"
	"net/http"
	"the_lonely_road/data"
	"the_lonely_road/models"
)

type App struct {
	userModel models.UserModel
}

const (
	port = "8080"
)

func main() {
	app := App{}
	svr := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: app.SetRoutes(),
	}

	err := svr.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

	cfg := data.DefaultPostgresConfig()
	db, err := data.Open(cfg)
	if err != nil {
		panic(err)
	}

	defer db.Close()
	app.userModel.DB = db
	fmt.Println("Connected to DB")

}
