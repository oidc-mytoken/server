package server

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

var server *fiber.App

func init() {
	server = fiber.New(fiber.Config{
		ReadTimeout: 30*time.Second,
		WriteTimeout:90*time.Second,
		IdleTimeout: 150*time.Second,
		ReadBufferSize: 8192,
		//WriteBufferSize: 4096,
	})
	addMiddlewares(server)
	addRoutes(server)
}

func addRoutes(s fiber.Router) {
	s.Get("/", handleTest)
	s.Get("/hello", handleTest)
	s.Get("/test", handleTest)
}

func start(s *fiber.App) {
	log.Fatal(s.Listen(":3000"))
}

func Start() {
	start(server)
}