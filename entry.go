package cttail

import (
	"fmt"

	"software.sslmate.com/src/certspotter"
	"software.sslmate.com/src/certspotter/ct"
)

type Entry struct {
	Identifiers *Identifiers `json:"identifiers"`
}

func parseEntry(entry *ct.LogEntry) (*Entry, error) {
	certInfo, err := certspotter.MakeCertInfoFromLogEntry(entry)
	if err != nil {
		return nil, fmt.Errorf("parseCert: %v", err)
	}
	ret := &Entry{}
	identifiers, err := parseIdentifiers(certInfo)
	if err != nil {
		return nil, fmt.Errorf("parse identifiers: %v", err)
	}
	ret.Identifiers = identifiers
	return ret, nil
}
