package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/pennz/amqp/config"
	"github.com/streadway/amqp"

	"crypto/tls"
	"crypto/x509"
)

func receiveLogTopic() {

	var tlsConfig tls.Config
	tlsConfig.RootCAs = x509.NewCertPool()

	if ca, err := ioutil.ReadFile("/etc/letsencrypt/archive/vtool.duckdns.org/chain1_ff.pem"); err == nil {
		if ok := tlsConfig.RootCAs.AppendCertsFromPEM(ca); ok == false {
			log.Fatalf("%s", "Failed to AppendCertsFromPEM")
		}
	} else {
		failOnError(err, "Failed to read cacert")
	}

	if cert, err := tls.LoadX509KeyPair("/etc/letsencrypt/archive/vtool.duckdns.org/fullchain1.pem", "/etc/letsencrypt/archive/vtool.duckdns.org/privkey1.pem"); err == nil {
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	} else {
		failOnError(err, "Failed to LoadX509KeyPair")
	}

	// see a note about Common Name (CN) at the top
	conn, err := amqp.DialTLS(config.AMQPURL, &tlsConfig)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"test-logs_topic", // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"test-queue", // name
		false,        // durable
		false,        // delete when unused
		true,         // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	if len(os.Args) < 2 {
		log.Printf("Usage: %s [binding_key]...", os.Args[0])
		os.Exit(0)
	}
	for _, s := range os.Args[1:] {
		log.Printf("Binding queue %s to exchange %s with routing key %s", q.Name, "test-logs_topic", s)
		err = ch.QueueBind(
			q.Name,            // queue name
			s,                 // routing key
			"test-logs_topic", // exchange
			false,
			nil)
		failOnError(err, "Failed to bind a queue")
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}
