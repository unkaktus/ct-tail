package cttail

import (
	"fmt"
	"strings"

	"software.sslmate.com/src/certspotter/ct/client"
)

type Tailer struct {
	client *client.LogClient
	tip    uint64
	bygone uint64
}

func NewTailer(logURL string, bygone uint64) *Tailer {
	c := client.New(strings.TrimRight(logURL, "/"))
	t := &Tailer{
		client: c,
		bygone: bygone,
	}
	return t
}

func (t *Tailer) FetchTip() ([]*Entry, error) {
	sth, err := t.client.GetSTH()
	if err != nil {
		return nil, fmt.Errorf("get STH: %w", err)
	}

	if t.tip == 0 {
		t.tip = sth.TreeSize - t.bygone
		t.bygone = 0
	}
	if t.tip == sth.TreeSize {
		return nil, nil
	}

	entries, err := t.client.GetEntries(int64(t.tip), int64(sth.TreeSize))
	if err != nil {
		return nil, fmt.Errorf("get entries: %w", err)
	}
	t.tip = sth.TreeSize
	ret := []*Entry{}

	for _, entry := range entries {
		e, err := parseEntry(&entry)
		if err != nil {
			return nil, fmt.Errorf("parse entry: %w", err)
		}
		ret = append(ret, e)
	}

	return ret, nil
}

func (t *Tailer) Reset(bygone uint64) {
	t.tip = 0
	t.bygone = bygone
}
