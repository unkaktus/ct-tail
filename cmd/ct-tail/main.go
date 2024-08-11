package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	cttail "github.com/nogoegst/ct-tail"
)

// Catch the ones being issued by LE, thus expiring in 3 months
func currentOak() string {
	year := time.Now().AddDate(0, 3, 0).Year()
	yearHalf := 1
	if time.Now().After(time.Date(year, 6, 20, 0, 0, 0, 0, time.UTC)) {
		yearHalf = 2
	}
	name := fmt.Sprintf("%dh%d", year, yearHalf)
	return fmt.Sprintf("https://oak.ct.letsencrypt.org/%s/", name)
}

type LockableEntries struct {
	Entries []*cttail.Entry
	sync.RWMutex
}

func run() error {
	var u = flag.String("u", "le-oak-current", "URL of the CT log")
	var bygone = flag.Uint64("n", 0, "Number of entries before the current STH to fetch (max. 254)")
	var interval = flag.Duration("i", 1*time.Second, "Interval between the fetches")
	var webPort = flag.String("web-port", "", "Run CT Wall website on specified port")
	flag.Parse()

	if *u == "le-oak-current" {
		*u = currentOak()
	}
	tailer := cttail.NewTailer(*u, *bygone)

	var entries = LockableEntries{}
	activeListeners := new(uint64)

	go func() {
		for {
			switch atomic.LoadUint64(activeListeners) {
			case 0:
				tailer = cttail.NewTailer(*u, *bygone)
			default:
				log.Printf("%v active listeners, fetching", atomic.LoadUint64(activeListeners))
				newEntries, err := tailer.FetchTip()
				if err != nil {
					log.Printf("fetching new entries: %v", err)
					return
				}
				if len(newEntries) != 0 {
					log.Printf("fetched %v entries", len(newEntries))
				}
				entries.Lock()
				entries.Entries = newEntries
				entries.Unlock()
			}
			time.Sleep(*interval)
		}
	}()

	switch {
	case *webPort != "":
		var upgrader = websocket.Upgrader{}
		h := http.NewServeMux()
		h.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(staticPage))
		}))
		h.Handle("/log", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddUint64(activeListeners, 1)
			defer atomic.AddUint64(activeListeners, ^uint64(0))

			flusher := w.(http.Flusher)
			w.Header().Set("X-Content-Type-Options", "nosniff")

			for {
				entries.RLock()
				es := entries.Entries
				entries.RUnlock()
				for _, entry := range es {
					jsonEntry, _ := json.Marshal(entry)
					_, err := fmt.Fprintf(w, "%s\n", jsonEntry)
					if err != nil {
						log.Printf("cannot write to client: %v", err)
						return
					}
				}
				if len(entries.Entries) != 0 {
					flusher.Flush()
				}
				time.Sleep(*interval)
			}
		}))
		h.Handle("/log/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Print("upgrade:", err)
				return
			}
			defer c.Close()

			atomic.AddUint64(activeListeners, 1)
			defer atomic.AddUint64(activeListeners, ^uint64(0))

			for {
				entries.RLock()
				es := entries.Entries
				entries.RUnlock()
				for _, entry := range es {
					jsonEntry, _ := json.Marshal(entry)
					err = c.WriteMessage(websocket.TextMessage, []byte(jsonEntry))
					if err != nil {
						log.Println("write:", err)
						break
					}
				}
				time.Sleep(*interval)
			}
		}))
		if err := http.ListenAndServe(":"+*webPort, h); err != nil {
			return fmt.Errorf("serve HTTP: %v", err)
		}
	default:
		atomic.AddUint64(activeListeners, 1)
		for {
			entries.RLock()
			es := entries.Entries
			entries.RUnlock()

			for _, entry := range es {
				jsonEntry, _ := json.Marshal(entry)
				fmt.Printf("%s\n", jsonEntry)
			}
			time.Sleep(*interval)
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
