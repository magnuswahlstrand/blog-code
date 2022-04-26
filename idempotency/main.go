package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/qiniu/qmgo"
	"log"
)

type User struct {
	ID    string `json:"id" bson:"_id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

var mongoConfig = &qmgo.Config{
	Uri:      "mongodb://localhost:27017",
	Database: "user-db",
	Coll:     "users",
}

func mongoClient() *qmgo.QmgoClient {
	mongo, err := qmgo.Open(context.Background(), mongoConfig)
	if err != nil {
		log.Fatalln(mongo)
	}
	return mongo
}

type Service struct {
	mongo *qmgo.QmgoClient
}

func setup() *fiber.App {
	mongo := mongoClient()

	s := Service{
		mongo: mongo,
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{AllowOrigins: "http://localhost:3000"}))
	app.Get("/users", s.ListUsers)
	app.Post("/users", s.AddUser)
	return app
}

func main() {
	app := setup()
	log.Fatal(app.Listen(":8080"))
}
