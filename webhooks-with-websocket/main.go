package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/lib/pq"
	webhooksdb "github.com/magnuswahlstrand/blog/webhooks-with-websocket/db"
	"gopkg.in/olahol/melody.v1"
	"io"
	"log"
	"net/http"
	"time"
)

func waitForNotification(l *pq.Listener, callback func(string) error) {
	fmt.Println("Start monitoring PostgreSQL...")
	for {
		select {
		case n := <-l.Notify:

			fmt.Println("Received data from channel [", n.Channel, "] :")
			// Prepare notification payload for pretty print

			if err := callback(n.Extra); err != nil {
				return
			}

		case <-time.After(90 * time.Second):
			fmt.Println("Received no events for 90 seconds, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}

type handler struct {
	websockets *melody.Melody
	queries    *webhooksdb.Queries
}

func (h *handler) handleWSUpgrade(w http.ResponseWriter, r *http.Request) {
	// TODO: Improve authorization :-)
	sID := r.URL.Query().Get("sid")
	id, err := uuid.Parse(sID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	keys := map[string]interface{}{
		"subscription_id": id,
	}
	if err := h.websockets.HandleRequestWithKeys(w, r, keys); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("io.ReadAll", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := h.queries.InsertWebhook(r.Context(), webhooksdb.InsertWebhookParams{
		SubscriptionID: uuid.MustParse("a2cce679-0b59-4245-a389-298a423945c0"),
		Payload:        b,
	})
	if err != nil {
		log.Println("insertWebhook", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	}); err != nil {
		log.Println("Encode", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}

func main() {
	var connectionInfo = "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"

	db, err := sql.Open("postgres", connectionInfo)
	if err != nil {
		log.Fatal(err)
	}

	if err := runMigrations(connectionInfo); err != nil {
		log.Fatal("failed to run migrations", err)
	}

	listener, err := dbListener(connectionInfo)
	if err != nil {
		log.Fatal("failed to start listener", err)
	}

	m := melody.New()
	queries := webhooksdb.New(db)

	s := Service{
		pubsub:  NewPubsub(),
		queries: queries,
	}
	h := handler{
		websockets: m,
		queries:    queries,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	// HTTP handlers
	r.Get("/ws", h.handleWSUpgrade)
	r.Post("/", h.handleEvent)

	// Websocket handlers
	m.HandleConnect(s.handleWebsocketConnect)
	m.HandleDisconnect(s.handleWebsocketDisconnect)
	m.HandleMessage(s.handleWebsocketMessage)

	go func() {
		waitForNotification(listener, s.handleDBNotifications)
	}()

	if err := http.ListenAndServe(":5001", r); err != nil {
		log.Fatal(err)
	}
}

func dbListener(connectionInfo string) (*pq.Listener, error) {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(connectionInfo, 10*time.Second, time.Minute, reportProblem)
	if err := listener.Listen("webhook_created"); err != nil {
		return nil, err
	}
	return listener, nil
}
