package main

import (
	"log"
	"os"

	"github.com/pennz/amqp/config"
	"github.com/pennz/amqp/session"
	"github.com/pennz/amqp/utils"
)

func emitLogTopic() {

	s := session.NewSession("test-session", config.AMQPURL, nil)
	defer s.Close()

	//var err error
	err := s.ExchangeDeclare("test-logs_topic", "topic")
	utils.FailOnError(err, "[S] Failed to declare an exchange")

	args := make([]string, len(os.Args)-1)
	copy(args[1:], os.Args[2:])
	routingKey := severityFrom(args)
	body := bodyFrom(args)

	for _, msg := range body {
		err = s.Publish("test-logs_topic", routingKey, []byte(msg))
		utils.FailOnError(err, "[S] Failed to publish a message")
	}

	// s.WaitPublishConfirm() // signal the channel, and the sesion go routine can return. Publish is
	// safe now.

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
