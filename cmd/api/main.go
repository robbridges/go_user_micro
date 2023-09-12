package main

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
	"the_lonely_road/models"
)

// I did my best to get away with no env variables in a mock service, but I can't expose my SMTP credentials. You win this round, env variables.
func init() {
	viper.SetConfigFile("email.env")
	viper.AddConfigPath("./")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("init: %w", err))
	}
}

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
