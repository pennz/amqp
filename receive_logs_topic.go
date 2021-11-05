package main

import (
	"bytes"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/pennz/amqp/config"
	"github.com/pennz/amqp/session"
	"github.com/pennz/amqp/utils"
)

func receiveLogTopic(queueName string, routingKeys []string) {

	if len(routingKeys) <= 0 {
		log.Printf("Usage: COMMAND r queueName routingKey...")
		os.Exit(0)
	}

	s := session.NewSession("test-session", config.AMQPURL, nil)
	defer s.Close()

	err := s.ExchangeDeclare("test-logs_topic", "topic")
	utils.FailOnError(err, "[S] Failed to declare an exchange")

	err = s.QueueDeclare(queueName)
	utils.FailOnError(err, "[R] Failed to declare a queue")

	for _, k := range routingKeys {
		log.Printf("[R] Binding queue %s to exchange %s with routing key %s", queueName, "test-logs_topic", k)

		err = s.QueueBind(
			queueName,         // queue name
			k,                 // routing key
			"test-logs_topic", // exchange
		)
		utils.FailOnError(err, "[R] Failed to bind a queue")
	}

	msgs, err := s.Consume(queueName, false)
	utils.FailOnError(err, "[R] Failed to register a consumer")
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

	messageLoopHandler := func() {
		for d := range msgs {
			log.Printf("[R] Received tag:%d body:%s", d.DeliveryTag, d.Body)
			dotCount := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dotCount)
			log.Printf("[W] Waiting for handling the message.Time estimated: %v\n", t*time.Second)
			time.Sleep(t * time.Second)
			log.Printf("[W] Working done for %v\n", t*time.Second)

			d.Ack(false)
		}
	}
	go messageLoopHandler()

	forever := make(chan bool)
	log.Printf("[R] Waiting for logs. To exit press CTRL+C")

	/*
		2021/05/08 08:35:10 Binding queue test-queue to exchange test-logs_topic with routing key a.info
		2021/05/08 08:35:10  [R] Waiting for logs. To exit press CTRL+C
		2021/05/08 08:35:40  [R] {0xc0000c25a0 map[]   1 0     0001-01-01 00:00:00 +0000 UTC    ctag-./amqp-1 0 1 false  test-queue [110 111 32 100 97 116 97]},  no data
	*/
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Printf("[R] %v received.\n", sig)
			forever <- true // main routine can go now, is will also ask other goroutines to exit
			break           // just one ^C is enough
		}
	}()
	<-forever
}
