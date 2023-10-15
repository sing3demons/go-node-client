package main

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"strconv"
	"syscall"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sirupsen/logrus"
)

type MyData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			err = ctx.Status(code).SendFile(fmt.Sprintf("./%d.html", code))
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}
			return nil
		},
	})
	app.Use(cors.New())
	app.Use(logger.New(
		logger.Config{
			Format:     "${pid} ${status} - ${method} ${path}\n",
			TimeFormat: "02-Jan-2006",
			TimeZone:   "Asia/Bangkok"},
	))
	app.Use(recover.New())
	app.Get("/api/v1/get_something", func(c *fiber.Ctx) error {
		id := c.Query("id")
		dataId, err := strconv.Atoi(id)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid ID")
		}
		response := MyData{ID: dataId, Name: fmt.Sprintf("Name_%d", dataId)}
		log.WithFields(logrus.Fields{
			"ID":   response.ID,
			"Name": response.Name,
		}).Info("Get Data")
		return c.Status(200).JSON(response)
	})

	//Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Info("shutting down the server")
		}
	}()

	<-ctx.Done()
	stop()

	fmt.Println("shutting down gracefully, press Ctrl+C again to force")

	if err := app.Shutdown(); err != nil {
		fmt.Println(err)
	}
}
