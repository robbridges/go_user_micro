package main

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
	"the_lonely_road/mailer"
	"the_lonely_road/models"
)

// I did my best to get away with no env variables in a mock service, but I can't expose my SMTP credentials. You win this round, env variables.
func setViper() {
	viper.SetConfigFile("email.env")
	viper.AddConfigPath("./")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("init: %w", err))
	}
}

type Config struct {
	cors struct {
		trustedOrigins []string
	}
}

type App struct {
	userModel models.IUserModel
	emailer   mailer.EmailService
	Config    Config
	wg        sync.WaitGroup
}

const (
	port = "8080"
)

func main() {
	// we use viper to read in our env variables
	setViper()
	app := App{}
	err := app.Serve()
	if err != nil {
		fmt.Println(err)
	}
	// unless more granular control is needed, we can set the cors trusted origins to all
	app.Config.cors.trustedOrigins = []string{"*"}

}
