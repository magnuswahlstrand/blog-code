package main

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

// Many tests taken from the IETF-draft at
// https://datatracker.ietf.org/doc/html/draft-ietf-httpapi-idempotency-key-header-00

//
//func addUser(t *testing.T, app *fiber.App, u Order) Order {
//	req := mustPostRequest(t, u)
//
//	res, err := app.Test(req, -1)
//	require.NoError(t, err)
//
//	require.Equalf(t, 200, res.StatusCode, res.Status)
//
//	var resp Order
//	err = json.NewDecoder(res.Body).Decode(&resp)
//	require.NoError(t, err)
//	return resp
//}

//
//func dropCollection(t *testing.T) {
//	ctx := context.Background()
//	db, err := qmgo.Open(ctx, mongoConfig)
//	require.NoError(t, err)

//	err = db.DropCollection(ctx)
//	require.NoError(t, err)
//
//	// Create index. Haven't found a way of doing this using qmgo. Slightly annoying ~_~
//	index := mongo.IndexModel{Keys: bsonx.Doc{{Key: "$**", Value: bsonx.String("text")}}}
//	mongoCollection, err := db.Collection.CloneCollection()
//	require.NoError(t, err)
//	_, err = mongoCollection.Indexes().CreateOne(ctx, index)
//	require.NoError(t, err)
//}

const HeaderIdempotencyKey = "Idempotency-Key"

func createOrder(t *testing.T, app *fiber.App, order Order, idempotencyKey string) *http.Response {
	req := mustPostRequest(t, "/order", order)
	req.Header.Set(HeaderIdempotencyKey, idempotencyKey)
	res, err := app.Test(req, -1)
	require.NoError(t, err)
	return res
}

func mustPostRequest(t *testing.T, url string, v interface{}) *http.Request {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(v)
	require.NoError(t, err)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		b,
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func Test_FirstRequest_Returns201(t *testing.T) {
	app, _ := setup()
	idempotencyKey := uuid.NewString()

	// Act
	res := createOrder(t, app, Order{ProductType: "bike"}, idempotencyKey)

	// Assert
	require.Equal(t, 201, res.StatusCode)
}

func Test_Retry_Returns201AndSamePayload(t *testing.T) {
	app, _ := setup()
	idempotencyKey := uuid.NewString()

	// Act
	res := createOrder(t, app, Order{ProductType: "bike"}, idempotencyKey)
	resRetry := createOrder(t, app, Order{ProductType: "bike"}, idempotencyKey)

	assert.Equal(t, http.StatusCreated, resRetry.StatusCode)
	assert.Equal(t, http.StatusCreated, resRetry.StatusCode)
	assert.Equal(t, res.Header, resRetry.Header)
	assert.Equal(t, res.Body, resRetry.Body)
}

/*
   If the "Idempotency-Key" request header is missing for a documented
   idempotent operation requiring this header, the resource server MUST
   reply with an HTTP "400" status code with body containing a link
   pointing to relevant documentation.  Alternately, using the HTTP
   header "Link", the client can be informed about the error as shown
   below.

   HTTP/1.1 400 Bad Request
   Link: <https://developer.example.com/idempotency>;
     rel="describedby"; type="text/html"

*/
func Test_IdempotencyKeyMissing_Returns400(t *testing.T) {
	app, _ := setup()

	// Act
	req := mustPostRequest(t, "/order", struct{}{})
	res, err := app.Test(req, -1)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 400, res.StatusCode)
	var respError fiber.Error
	require.NoError(t, json.NewDecoder(res.Body).Decode(&respError))
	require.Equal(t, "idempotency key missing", respError.Message)
}

func Test_IdempotencyKeyNotUUIDv4_Returns400(t *testing.T) {
	app, _ := setup()

	// Act
	req := mustPostRequest(t, "/order", struct{}{})
	req.Header.Set(HeaderIdempotencyKey, "not-a-uuid-v4")
	res, err := app.Test(req, -1)
	require.NoError(t, err)

	// Assert
	require.Equal(t, 400, res.StatusCode)
	var respError fiber.Error
	require.NoError(t, json.NewDecoder(res.Body).Decode(&respError))
	require.Equal(t, "idempotency must be a UUID V4", respError.Message)
}

/*
 If there is an attempt to reuse an idempotency key with a different
  request payload, the resource server MUST reply with a HTTP "422"
  status code with body containing a link pointing to relevant
  documentation.  The status code "422" is defined in Section 11.2 of
  [RFC4918].  The server can also inform the client by using the HTTP
  header "Link" as shown below.

  HTTP/1.1 422 Unprocessable Entity
  Link: <https://developer.example.com/idempotency>;
  rel="describedby"; type="text/html"

*/
func Test_IdempotencyKeyReusedWithDifferentPayload_Returns422(t *testing.T) {
	app, _ := setup()

	// Arrange
	idempotencyKey := uuid.NewString()
	res := createOrder(t, app, Order{ProductType: "bike"}, idempotencyKey)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	// Act
	res = createOrder(t, app, Order{ProductType: "car"}, idempotencyKey)

	// Assert
	require.Equal(t, 422, res.StatusCode)
	var respError fiber.Error
	require.NoError(t, json.NewDecoder(res.Body).Decode(&respError))
	require.Equal(t, "idempotency key previously used with other payload", respError.Message)
}

/*
  If the request is retried, while the original request is still being
  processed, the resource server MUST reply with an HTTP "409" status
  code with body containing a link or the HTTP header "Link" pointing
  to the relevant documentation.

  HTTP/1.1 409 Conflict
  Link: <https://developer.example.com/idempotency>;
  rel="describedby"; type="text/html"
*/

func Test_InitialRequestNotCompleted_Returns409(t *testing.T) {
	waitCh := make(chan bool)
	app, service := setup()

	// Override service process
	service.Process = func() {
		waitCh <- true // Switch to main
		<-waitCh       // Wait for main
	}

	// Arrange
	idempotencyKey := uuid.NewString()
	go func() {
		res := createOrder(t, app, Order{ProductType: "car"}, idempotencyKey)
		require.Equal(t, http.StatusCreated, res.StatusCode)
		waitCh <- true
	}()

	// Act
	<-waitCh // Wait for goroutine
	res := createOrder(t, app, Order{ProductType: "car"}, idempotencyKey)
	waitCh <- true // Restart goroutine to main

	// Assert
	require.Equal(t, 409, res.StatusCode)
	var respError fiber.Error
	require.NoError(t, json.NewDecoder(res.Body).Decode(&respError))
	require.Equal(t, "request already in process", respError.Message)

	<-waitCh // Wait for goroutine to finish
}
