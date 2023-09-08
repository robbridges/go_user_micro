package main

import (
	"fmt"
	"sync"
	"the_lonely_road/models"
)

type App struct {
	userModel models.IUserModel
	wg        sync.WaitGroup
}

const (
	port = "8080"
)

func main() {

	app := App{}
	err := app.Serve()
	if err != nil {
		fmt.Println(err)
	}

}
