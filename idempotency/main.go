package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/qiniu/qmgo"
	"log"
)

type Order struct {
	ProductType string `json:"product_type"`
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
	mongo   *qmgo.QmgoClient
	db2     db2
	Process func()
}

func setup() (*fiber.App, *Service) {
	//mongo := mongoClient()

	s := &Service{
		db2:     db2{db: map[string][]byte{}},
		Process: func() {}, // Can be overridden in tests
	}

	app := fiber.New()
	app.Post("/order", s.HandlerOrder)
	return app, s
}

func main() {
	app, _ := setup()
	log.Fatal(app.Listen(":8080"))
}
