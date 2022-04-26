package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/qiniu/qmgo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"net/http"
	"testing"
)

func listUsers(t *testing.T, app *fiber.App) ListUsersResponse {

	req, _ := http.NewRequest(
		"GET",
		"/users",
		nil,
	)

	res, err := app.Test(req, -1)
	require.NoError(t, err)

	require.Equal(t, 200, res.StatusCode)

	var resp ListUsersResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	require.NoError(t, err)
	return resp
}

func addUser(t *testing.T, app *fiber.App, u User) User {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(u)
	require.NoError(t, err)

	req, _ := http.NewRequest(
		"POST",
		"/users",
		b,
	)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := app.Test(req, -1)
	require.NoError(t, err)

	require.Equalf(t, 200, res.StatusCode, res.Status)

	var resp User
	err = json.NewDecoder(res.Body).Decode(&resp)
	require.NoError(t, err)
	return resp
}

func dropCollection(t *testing.T) {
	ctx := context.Background()
	db, err := qmgo.Open(ctx, mongoConfig)
	require.NoError(t, err)
	err = db.DropCollection(ctx)
	require.NoError(t, err)

	// Create index. Haven't found a way of doing this using qmgo. Slightly annoying ~_~
	index := mongo.IndexModel{Keys: bsonx.Doc{{Key: "$**", Value: bsonx.String("text")}}}
	mongoCollection, err := db.Collection.CloneCollection()
	require.NoError(t, err)
	_, err = mongoCollection.Indexes().CreateOne(ctx, index)
	require.NoError(t, err)
}

func Test_ListUsers(t *testing.T) {
	dropCollection(t)

	app := setup()
	addUser(t, app, User{
		Email: "magnus@example.org",
		Name:  "Magnus",
		Phone: "+46791008821",
	})
	addUser(t, app, User{
		Email: "john@website.com",
		Name:  "John",
		Phone: "+46788723230",
	})

	resp := listUsers(t, app)
	require.Len(t, resp.Users, 2)
}

func Test_ListUsersWhenEmpty(t *testing.T) {
	dropCollection(t)

	app := setup()

	resp := listUsers(t, app)
	require.Len(t, resp.Users, 0)
}

func Test_GenerateFakeData(t *testing.T) {
	dropCollection(t)

	gofakeit.Seed(0)

	app := setup()
	for i := 0; i < 100; i++ {
		addUser(t, app, User{
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
			Phone: gofakeit.PhoneFormatted(),
		})
	}

	resp := listUsers(t, app)
	require.Len(t, resp.Users, 100)
}
