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

func main() {

	var connectionInfo = "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"

	db, err := sql.Open("postgres", connectionInfo)
	if err != nil {
		log.Fatal(err)
	}

	if err := runMigrations(connectionInfo); err != nil {
		log.Fatal("failed to run migrations", err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(connectionInfo, 10*time.Second, time.Minute, reportProblem)
	if err = listener.Listen("webhook_created"); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Start monitoring PostgreSQL...")

	s := Service{
		sessions: map[uuid.UUID]*WebhookSession{},
		queries:  webhooksdb.New(db),
	}
	go func() {
		for {
			s.waitForNotification(listener)
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	m := melody.New()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		sID := r.URL.Query().Get("sid")
		keys := map[string]interface{}{
			"subscription_id": uuid.MustParse(sID),
		}
		m.HandleRequestWithKeys(w, r, keys)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal("io.ReadAll", err)
		}

		id, err := s.queries.InsertWebhook(r.Context(), webhooksdb.InsertWebhookParams{
			SubscriptionID: uuid.MustParse("a2cce679-0b59-4245-a389-298a423945c0"),
			Payload:        b,
		})
		if err != nil {
			log.Fatal("insertWebhook", err)
		}

		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"id": id,
		}); err != nil {
			log.Fatal("encode JSON", err)
		}
		return
	})

	m.HandleConnect(func(sess *melody.Session) {
		fmt.Println("New connection")
		subID := sess.MustGet("subscription_id").(uuid.UUID)
		s.AddWebhookSession(sess, subID)
	})
	m.HandleDisconnect(func(sess *melody.Session) {
		fmt.Println("Disconnected")
		subID := sess.MustGet("subscription_id").(uuid.UUID)
		s.RemoveWebhookSession(subID)
	})
	m.HandleMessage(func(sess *melody.Session, received []byte) {
		fmt.Println("Receive message")
		subID := sess.MustGet("subscription_id").(uuid.UUID)

		// TODO: handle JSON events
		id, err := uuid.Parse(string(received))
		if err != nil {
			log.Println("error when parsing received message", err)
			return
		}

		s.ForwardAck(id, subID)
	})
	if err := http.ListenAndServe(":5001", r); err != nil {
		log.Fatal(err)
	}
}

//
//type Webhook struct {
//	ID             uuid.UUID              `json:"id"`
//	SubscriptionID uuid.UUID              `json:"subscription_id"`
//	Payload        map[string]interface{} `json:"payload"`
//	PublishedAt    bool                   `json:"published_at"`
//}
