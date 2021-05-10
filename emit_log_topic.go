package main

import (
	"log"
	"os"

	"github.com/streadway/amqp"
)

func emitLogTopic() {

	connCloseCh := make(chan bool)
	defer func() {
		<-connCloseCh
		closeMyConnection()
	}()

	chCloseCh := make(chan bool)
	ch, err := myConn.Channel()
	failOnError(err, "[S] Failed to open a channel")
	defer func() {
		<-chCloseCh
		log.Printf("[S] conn.Channel Closing.\n")
		ch.Close()
		log.Printf("[S] conn.Channel Closed.\n")
	}()

	msgOpCh := make(chan int, 10)

	err = ch.Confirm(false)
	failOnError(err, "[S] Failed to set channel to confirm mode")

	/*serverErrorReturnCh := make(chan amqp.Return, 10)
	ch.NotifyReturn(serverErrorReturnCh)

	go func() {
		for errorReturn := <-serverErrorReturnCh; ; {
			log.Printf("[S] Failed for a publish %s\n", errorReturn.ReplyText)
		}
	}()*/

	err = ch.ExchangeDeclare(
		"test-logs_topic", // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
	failOnError(err, "[S] Failed to declare an exchange")

	args := make([]string, len(os.Args)-1)
	copy(args[1:], os.Args[2:])
	routingKey := severityFrom(args)
	body := bodyFrom(args)

	msgCount := 0
	for _, msg := range body {
		err = ch.Publish(
			"test-logs_topic", // exchange
			routingKey,        // routing key
			true,              // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(msg),
			})
		failOnError(err, "[S] Failed to publish a message")
		log.Printf("[S] Client published with RK: %s, msgCount %d \n",
			routingKey,
			msgCount)

		msgOpCh <- msgCount
		msgCount++
	}
	close(msgOpCh)

	confirmationCh := make(chan amqp.Confirmation, 10) // 10 Some magic number
	ch.NotifyPublish(confirmationCh)

	go func() {
		for confirm := <-confirmationCh; ; {
			msgCount, ok := <-msgOpCh
			if !ok {
				log.Printf("[S] Publish confirmed for all.\n")
				chCloseCh <- true
				connCloseCh <- true
			} else {
				log.Printf("[S] RK: %s, Publish confirmed for %v, msgCount %d \n",
					routingKey,
					confirm,
					msgCount)
			}
		}
	}()

	log.Printf("[S] Sent %s with routing key %s Done.", body, routingKey)
}

func bodyFrom(args []string) []string {
	var s []string

	if (len(args) < 3) || args[2] == "" {
		s = []string{"hello"}
	} else {
		s = args[2:]
	}
	return s
}

func severityFrom(args []string) string {
	var s string
	if (len(args) < 2) || args[1] == "" {
		s = "anonymous.info"
	} else {
		s = args[1]
	}
	return s
}
