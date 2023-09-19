package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"the_lonely_road/data"
	"the_lonely_road/mailer"
	"the_lonely_road/models"
	"time"
)

func (app *App) Serve() error {
	svr := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: app.SetRoutes(),
	}

	shutDownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		// listen for sigint or sigterm calls, relay them to the channel
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		fmt.Println("Shutting down server", map[string]string{
			"signal": s.String(),
		})

		// in flight requests have a 30-second grace period
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		defer cancel()

		err := svr.Shutdown(ctx)
		if err != nil {
			shutDownError <- err
		}

		fmt.Println("completing background tasks", map[string]string{
			"addr": svr.Addr,
		})

		app.wg.Wait()
		shutDownError <- nil
	}()

	db, err := data.OpenDSN("DSN_DB")
	if err != nil {
		return err
	}
	fmt.Println("Connected to database")

	defer func() {
		if err := db.Close(); err != nil {
			fmt.Println("Error closing server:", err)
		}
	}()

	mailCfg := mailer.DefaultSMTPConfig()
	appMailer := mailer.NewEmailService(mailCfg)
	app.userModel = &models.UserModel{
		DB: db,
	}
	app.emailer = appMailer
	fmt.Println("Server running on port 8080")
	err = svr.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
