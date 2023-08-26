package main

import (
	"fmt"
	"net/http"
)

type App struct {
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

}
