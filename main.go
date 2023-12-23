package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
)

func main() {
	app := fiber.New()
	app.Use(
		recover.New(),
		logger.New(logger.Config{
			TimeFormat: ServerConfig.LogTimeFormat,
			TimeZone:   ServerConfig.LogTimezone,
		}),
		compress.New(compress.Config{
			Level: ServerConfig.CompressLevel,
		}),
		NewS3Handler(),
	)
	log.Fatal(app.Listen(ServerConfig.Listen))
}
