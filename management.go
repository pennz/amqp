package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
)

var tlsConfig = myTLSConfig()

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func myTLSConfig() *tls.Config {
	var tlsConfig tls.Config
	tlsConfig.RootCAs = x509.NewCertPool()

	// If you use self generated CA root (chain), you can append it. Otherwise,
	// the default ones trusted can work?
	if ca, err := ioutil.ReadFile("/etc/letsencrypt/archive/vtool.duckdns.org/chain1_ff.pem"); err == nil {
		if ok := tlsConfig.RootCAs.AppendCertsFromPEM(ca); ok == false {
			log.Fatalf("%s", "Failed to AppendCertsFromPEM")
		}
	} else {
		failOnError(err, "Failed to read PEM encoded certificates")
	}

	if cert, err := tls.LoadX509KeyPair("/etc/letsencrypt/archive/vtool.duckdns.org/fullchain1.pem", "/etc/letsencrypt/archive/vtool.duckdns.org/privkey1.pem"); err == nil {
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	} else {
		failOnError(err, "Failed to LoadX509KeyPair")
	}
	return &tlsConfig
}

// Session copied from https://pkg.go.dev/github.com/streadway/amqp#example-package
