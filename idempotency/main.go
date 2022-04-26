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
	mongo *qmgo.QmgoClient
	db2   db2
}

func setup(shouldSleep ...bool) *fiber.App {
	//mongo := mongoClient()

	var sleeper bool
	if len(shouldSleep) > 0 {
		sleeper = true
	}

	s := Service{
		//mongo: mongo,

		db2: db2{
			db:          map[string][]byte{},
			shouldSleep: sleeper,
		},
	}

	app := fiber.New()
	app.Post("/order", s.HandlerOrder)
	return app
}

func main() {
	app := setup()
	log.Fatal(app.Listen(":8080"))
}
