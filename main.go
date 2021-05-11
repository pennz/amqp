package main

import (
	"log"
	"os"
)

func main() {
	//emitLogTopic()

	if len(os.Args) < 2 {
		log.Printf("Usage: %s command [command_paramter...]\n", os.Args[0])
		os.Exit(0)
	}
	if os.Args[1] == "r" {
		if len(os.Args) < 4 {
			log.Printf("Usage: %s r queueName routingKey...", os.Args[0])
			os.Exit(0)
		}
		receiveLogTopic(os.Args[2], os.Args[3:])
	}
	if os.Args[1] == "s" {
		emitLogTopic()
	}
}
