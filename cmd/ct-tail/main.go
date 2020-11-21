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

	cttail "github.com/nogoegst/ct-tail"
)

func currentOak() string {
	year := time.Now().Year()
	return fmt.Sprintf("https://oak.ct.letsencrypt.org/%d/", year)
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
			if atomic.LoadUint64(activeListeners) != 0 {
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
