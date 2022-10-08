package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/long2ice/s3web/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	app := fiber.New()
	app.Use(
		recover.New(),
		logger.New(logger.Config{
			TimeFormat: config.ServerConfig.LogTimeFormat,
			TimeZone:   config.ServerConfig.LogTimezone,
		}),
		compress.New(compress.Config{
			Level: config.ServerConfig.CompressLevel,
		}),
		NewS3Handler(),
	)
	log.Fatal(app.Listen(config.ServerConfig.Listen))
}
