package cttail

import (
	"fmt"

	"golang.org/x/net/idna"
	"software.sslmate.com/src/certspotter"
)

type Identifiers struct {
	DNSNames []string `json:"dns_names,omitempty"`
	IPAddrs  []string `json:"ip_addrs,omitempty"`
}

func parseIdentifiers(certInfo *certspotter.CertInfo) (*Identifiers, error) {
	rawIdentifiers, err := certInfo.ParseIdentifiers()
	if err != nil {
		return nil, fmt.Errorf("parse identifiers: %v", err)
	}
	identifiers := mapIdentifiers(rawIdentifiers)
	return identifiers, nil
}

func mapIdentifiers(in *certspotter.Identifiers) *Identifiers {
	out := &Identifiers{}
	for _, dnsname := range in.DNSNames {
		if unicode, err := idna.Punycode.ToUnicode(dnsname); err != nil {
			out.DNSNames = append(out.DNSNames, unicode)
		} else {
			out.DNSNames = append(out.DNSNames, dnsname)
		}
	}
	for _, ipaddr := range in.IPAddrs {
		out.IPAddrs = append(out.IPAddrs, ipaddr.String())
	}
	return out
}
