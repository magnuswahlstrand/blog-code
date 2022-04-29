package main

import (
	"bytes"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/qiniu/qmgo"
	"log"
	"net/http"
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
	Process func()
}

type IdempotenceChecker struct {
	mongo  *qmgo.QmgoClient
	mongo2 inMemoryDB
}

func (i *IdempotenceChecker) IdempotencyMiddleware(c *fiber.Ctx) error {
	// 1. Parse the idempotency key
	idempotencyKey := c.Get(HeaderIdempotencyKey)
	if idempotencyKey == "" {
		return responseError(c, http.StatusBadRequest, "idempotency key missing")
	}

	if _, err := uuid.Parse(idempotencyKey); err != nil {
		return responseError(c, http.StatusBadRequest, "idempotency must be a UUID V4")
	}

	// 2. Check if request exists
	recorded, exists, err := i.mongo2.get(idempotencyKey)
	if err != nil {
		return responseError(c, http.StatusInternalServerError, err.Error())
	}

	if exists {
		// 3a. If it exists, is it the same request body as before? --> 422
		if !bytes.Equal(c.Body(), recorded.RequestBody) {
			return responseError(c, http.StatusUnprocessableEntity, "idempotency key previously used with other payload")
		}

		// 4a. Is the initial request still in flight? --> 409
		if recorded.Completed == false {
			return responseError(c, http.StatusConflict, "request already in process")
		}

		// 5a. Everything is OK. Re-send previous response
		c.Response().Header.SetContentType("application/json")
		c.Response().SetBodyRaw(recorded.Response.Data)
		c.Response().SetStatusCode(recorded.Response.Status)
		return nil
	}

	// 3b. The idempotency key hasn't been seen before.
	// Store the key and the request body in the database
	record := RecordedRequestResponse{
		RequestBody: c.Body(),
		Completed:   false,
	}
	if err = i.mongo2.update(idempotencyKey, record); err != nil {
		return responseError(c, http.StatusInternalServerError, err.Error())
	}

	// 4. Go to the regular handler
	if err = c.Next(); err != nil {
		return responseError(c, http.StatusInternalServerError, err.Error())
	}

	// Record results
	record.Completed = true
	record.Response = Response{
		Status: c.Response().StatusCode(),
		Data:   c.Response().Body(),
	}
	if err = i.mongo2.update(idempotencyKey, record); err != nil {
		return responseError(c, http.StatusInternalServerError, err.Error())
	}

	return nil
}

func setup() (*fiber.App, *Service) {
	mongo := mongoClient()
	checker := IdempotenceChecker{
		mongo:  mongo,
		mongo2: inMemoryDB{db: map[string][]byte{}},
	}
	s := &Service{
		Process: func() {}, // Can be overridden in tests
	}

	app := fiber.New()

	app.Use(checker.IdempotencyMiddleware)
	app.Post("/order", s.HandlerOrder)
	return app, s
}

func main() {
	app, _ := setup()
	log.Fatal(app.Listen(":8080"))
}
