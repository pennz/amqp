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

// Session copied from https://pkg.go.dev/github.com/streadway/amqp#example-package
func myTLSConfig() *tls.Config {
	var tlsConfig tls.Config
	tlsConfig.RootCAs = x509.NewCertPool()

	// If you use self generated CA root (chain), you can append it. Otherwise,
	// the default ones trusted can work? -> no, for real CA root, this step is still needed

	if ca, err := ioutil.ReadFile("/usr/local/etc/letsencrypt/live/dotovertls.duckdns.org/chain.pem"); err == nil {
		if ok := tlsConfig.RootCAs.AppendCertsFromPEM(ca); ok == false {
			log.Fatalf("%s", "Failed to AppendCertsFromPEM")
		}
	} else {
		failOnError(err, "Failed to read PEM encoded certificates")
	}

	if cert, err := tls.LoadX509KeyPair("/usr/local/etc/letsencrypt/live/dotovertls.duckdns.org/fullchain.pem", "/usr/local/etc/letsencrypt/live/dotovertls.duckdns.org/privkey.pem"); err == nil {
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	} else {
		failOnError(err, "Failed to LoadX509KeyPair")
	}
	return &tlsConfig
}
