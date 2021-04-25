package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/lib/pq"
)

var config sync.Map

func main() {
	connStr := "postgresql://john:123456@127.0.0.1:5432/development?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	minReconn := 10 * time.Second
	maxReconn := time.Minute
	listener := pq.NewListener(connStr, minReconn, maxReconn, reportProblem)
	err = listener.Listen("config")
	if err != nil {
		panic(err)
	}

	done := make(chan bool)
	defer close(done)

	go func() {
		fmt.Println("entering main loop")
		for {
			select {
			case <-done:
				return
			default:
				// Process all available work before waiting for notifications.
				//getWork(db)
				waitForNotification(listener)
			}
		}
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/", configHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]interface{})
	config.Range(func(k, v interface{}) bool {
		m[k.(string)] = v
		return true
	})
	json.NewEncoder(w).Encode(m)
}

func waitForNotification(l *pq.Listener) {
	select {
	case msg := <-l.Notify:
		fmt.Println("received notification, new work available")
		m := make(map[string]interface{})
		if err := json.Unmarshal([]byte(msg.Extra), &m); err != nil {
			fmt.Printf("error unmarshal %s: %v\n", msg.Extra, err)
		}
		for k, v := range m {
			config.Store(k, v)
		}
		fmt.Println(msg.Channel, msg.Extra)
	case <-time.After(90 * time.Second):
		go l.Ping()
		// Check if there's more work available, just in case
		// it takes a while for the Listener to notice
		// connection lost and reconnect.
		fmt.Println("received no work for 90 seconds, chekcing for new work")
	}
}
