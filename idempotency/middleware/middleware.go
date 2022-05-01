package middleware

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	errorhandler "github.com/magnuswahlstrand/blog-code/idempotency/error_handler"
	"net/http"
)

const HeaderIdempotencyKey = "Idempotency-Key"

type RequestRecording struct {
	RequestBody []byte `bson:"request_body"`

	ResponseStatus int    `bson:"response_status"`
	ResponseBody   []byte `bson:"response_body"`

	Completed bool `bson:"completed"`
}

type base struct {
	mongo2 db
}

func (i *base) IdempotencyMiddleware(c *fiber.Ctx) error {
	// 1. Parse the idempotency key
	idempotencyKey := c.Get(HeaderIdempotencyKey)
	if idempotencyKey == "" {
		return errorhandler.Handle(c, http.StatusBadRequest, "idempotency key missing")
	}

	if _, err := uuid.Parse(idempotencyKey); err != nil {
		return errorhandler.Handle(c, http.StatusBadRequest, "idempotency must be a UUID V4")
	}

	// 2. Check if key has been seen before
	record, exists, err := i.mongo2.get(c.Context(), idempotencyKey)
	if err != nil {
		return errorhandler.Handle(c, http.StatusInternalServerError, err.Error())
	}

	if exists {
		// 3a. If it exists, is it the same request body as before? --> 422
		if !bytes.Equal(c.Body(), record.RequestBody) {
			return errorhandler.Handle(c, http.StatusUnprocessableEntity, "idempotency key previously used with other payload")
		}

		// 4a. Is the initial request still in flight? --> 409
		if record.Completed == false {
			return errorhandler.Handle(c, http.StatusConflict, "request already in process")
		}

		// 5a. Everything is OK. Re-send previous response
		c.Response().SetStatusCode(record.ResponseStatus)
		c.Response().SetBodyRaw(record.ResponseBody)
		c.Response().Header.SetContentType("application/json")
		return nil
	}

	// 3b. The idempotency key hasn't been seen before.
	// Store the key and the request body in the database
	newRecord := RequestRecording{
		RequestBody: c.Body(),
		Completed:   false,
	}
	if err = i.mongo2.update(c.Context(), idempotencyKey, newRecord); err != nil {
		return errorhandler.Handle(c, http.StatusInternalServerError, err.Error())
	}

	// 4a. Go to the regular handler
	if err = c.Next(); err != nil {
		return errorhandler.Handle(c, http.StatusInternalServerError, err.Error())
	}

	// 5a. Record results
	newRecord.Completed = true
	newRecord.ResponseStatus = c.Response().StatusCode()
	newRecord.ResponseBody = c.Response().Body()
	if err = i.mongo2.update(c.Context(), idempotencyKey, newRecord); err != nil {
		return errorhandler.Handle(c, http.StatusInternalServerError, err.Error())
	}

	return nil
}

func New(db db) *base {
	return &base{
		mongo2: db,
	}
}
