package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/abassGarane/url_shortener/api"
	"github.com/abassGarane/url_shortener/repository/mongo"
	rr "github.com/abassGarane/url_shortener/repository/redis"
	"github.com/abassGarane/url_shortener/shortener"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func httpPort() string {
	port := "8000"

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	return fmt.Sprintf(":%s", port)
}
func chooseRepository() shortener.RedirectRepository {
	switch os.Getenv("URL_DB") {
	case "redis":
		redisURL := os.Getenv("REDIS_URL")
		repo, err := rr.NewRedisRepository(redisURL)
		if err != nil {
			log.Fatal(err)
		}
		return repo
	case "mongo":
		mongoURL := os.Getenv("MONGO_URL")
		mongodb := os.Getenv("MONGO_DB")
		mongoTimeout, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
		repo, err := mongo.NewMongoRepository(mongoURL, mongodb, mongoTimeout)
		if err != nil {
			log.Fatal(err)
		}
		return repo
	}
	return nil
}
func main() {
	repo := chooseRepository()
	service := shortener.NewRedirectService(repo)
	handler := api.NewHandler(&service)
	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Get("/:code", handler.Get)

}
