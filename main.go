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
	setUpMyAMQPConnection()

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

package main

// This exports a Session object that wraps this library. It
// automatically reconnects when the connection fails, and
// blocks all pushes until the connection succeeds. It also
// confirms every outgoing message, so none are lost.
// It doesn't automatically ack each message, but leaves that
// to the parent process, since it is usage-dependent.
//
// Try running this in one terminal, and `rabbitmq-server` in another.
// Stop & restart RabbitMQ to see how the queue reacts.
func main() {
	name := "job_queue"
	addr := "amqp://guest:guest@localhost:5672/"
	queue := New(name, addr)
	message := []byte("message")
	// Attempt to push a message every 2 seconds
	for {
		time.Sleep(time.Second * 3)
		if err := queue.Push(message); err != nil {
			fmt.Printf("Push failed: %s\n", err)
		} else {
			fmt.Println("Push succeeded!")
		}
	}
}
