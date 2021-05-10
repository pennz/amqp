package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"github.com/pennz/amqp/config"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

var myConn *amqp.Connection

func setUpMyAMQPConnection() *amqp.Connection {
	if myConn == nil {
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

		// see a note about Common Name (CN) at the top
		var err error
		myConn, err = amqp.DialTLS(config.AMQPURL, &tlsConfig)
		failOnError(err, "Failed to connect to RabbitMQ")
	}
	return myConn
}

func closeMyConnection() {
	log.Printf("[M] Connection Closing.\n")
	myConn.Close()
	log.Printf("[M] Connection Closed.\n")
}
