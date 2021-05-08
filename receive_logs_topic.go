package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/pennz/amqp/config"
	"github.com/streadway/amqp"

	"crypto/tls"
	"crypto/x509"
)

func receiveLogTopic() {

	if len(os.Args) < 2 {
		log.Printf("Usage: %s [binding_key]...", os.Args[0])
		os.Exit(0)
	}

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
	conn, err := amqp.DialTLS(config.AMQPURL, &tlsConfig)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer func() { log.Printf("Connection Closing.\n"); conn.Close(); log.Printf("Connection Closed.\n") }()

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
	/*

		// Delivery captures the fields for a previously delivered message resident in
		// a queue to be delivered by the server to a consumer from Channel.Consume or
		// Channel.Get.
		type Delivery struct {
			Acknowledger Acknowledger // the channel from which this delivery arrived

			Headers Table // Application or header exchange table

			// Properties
			ContentType     string    // MIME content type
			ContentEncoding string    // MIME content encoding
			DeliveryMode    uint8     // queue implementation use - non-persistent (1) or persistent (2)
			Priority        uint8     // queue implementation use - 0 to 9
			CorrelationId   string    // application use - correlation identifier
			ReplyTo         string    // application use - address to reply to (ex: RPC)
			Expiration      string    // implementation use - message expiration spec
			MessageId       string    // application use - message identifier
			Timestamp       time.Time // application use - message timestamp
			Type            string    // application use - message type name
			UserId          string    // application use - creating user - should be authenticated user
			AppId           string    // application use - creating application id

			// Valid only with Channel.Consume
			ConsumerTag string

			// Valid only with Channel.Get
			MessageCount uint32

			DeliveryTag uint64
			Redelivered bool
			Exchange    string // basic.publish exchange
			RoutingKey  string // basic.publish routing key

			Body []byte
		}
	*/
	failOnError(err, "Failed to register a consumer")

	messageLoopHandler := func() {
		for d := range msgs {
			log.Printf(" [x] %v,  %s", d, d.Body)
		}
	}
	go messageLoopHandler()

	forever := make(chan bool)
	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")

	/*
		2021/05/08 08:35:10 Binding queue test-queue to exchange test-logs_topic with routing key a.info
		2021/05/08 08:35:10  [*] Waiting for logs. To exit press CTRL+C
		2021/05/08 08:35:40  [x] {0xc0000c25a0 map[]   1 0     0001-01-01 00:00:00 +0000 UTC    ctag-./amqp-1 0 1 false  test-queue [110 111 32 100 97 116 97]},  no data
	*/
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Printf("%v", sig)
		}
	}()
	<-forever
}
