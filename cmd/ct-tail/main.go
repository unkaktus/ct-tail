package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	cttail "github.com/nogoegst/ct-tail"
)

func currentOak() string {
	year := time.Now().Year()
	return fmt.Sprintf("https://oak.ct.letsencrypt.org/%d/", year)
}

func run() error {
	var u = flag.String("u", "le-oak-current", "URL of the CT log")
	var bygone = flag.Uint64("n", 0, "Number of entries before the current STH to fetch (max. 254)")
	var interval = flag.Duration("i", 1*time.Second, "Interval between the fetches")
	flag.Parse()
	if *u == "le-oak-current" {
		*u = currentOak()
	}
	tailer := cttail.NewTailer(*u, *bygone)
	for {
		entries, err := tailer.FetchTip()
		if err != nil {
			return err
		}
		for _, entry := range entries {
			jsonEntry, _ := json.Marshal(entry)
			fmt.Printf("%s\n", jsonEntry)
		}
		time.Sleep(*interval)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
