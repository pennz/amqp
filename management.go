package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

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
type Session struct {
	name               string
	logger             *log.Logger
	connection         *amqp.Connection
	channel            *amqp.Channel
	done               chan bool
	notifyConnClose    chan *amqp.Error
	notifyChanClose    chan *amqp.Error
	notifyConfirm      chan amqp.Confirmation
	clientPublishState chan int
	isReady            bool
	publishCount       int
}

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second

	// When setting up the channel after a channel exception
	reInitDelay = 2 * time.Second

	// When resending messages the server didn't confirm
	resendDelay = 5 * time.Second

	closeDelay = 1 * time.Second

	NotifyMax = 10
)

var (
	errNotConnected  = errors.New("not connected to a server")
	errAlreadyClosed = errors.New("already closed: not connected to the server")
	errShutdown      = errors.New("session has been shut down")
	errConfirm       = errors.New("confirmation is not right")
	tlsConfig        = myTLSConfig()
)

// NewSession creates a new consumer state instance, and automatically
// attempts to connect to the server.
func NewSession(name string, addr string, tlsConfig *tls.Config) *Session {
	session := Session{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		name:   name,
		done:   make(chan bool),
	}
	go session.handleReconnect(addr, tlsConfig)
	return &session
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (session *Session) handleReconnect(addr string, tlsConfig *tls.Config) {
	for {
		session.isReady = false
		log.Println("Attempting to connect")

		var conn *amqp.Connection
		var err error
		if tlsConfig == nil {
			conn, err = session.connect(addr)
		} else {
			conn, err = session.connectTLS(addr, tlsConfig)
		}

		if err != nil {
			log.Println("Failed to connect. Retrying...")

			select {
			case <-session.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}
		// Connection connected

		if done := session.handleReInit(conn); done {
			break // session.done and we can exit
		} // else we will reconnect, in the handleReInit, notifyConnClose is
		// monitored and if received, it will started over from the connection
		// establishing
	}
}

// handleReconnect will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (session *Session) handleReInit(conn *amqp.Connection) bool {
	for {
		session.isReady = false

		err := session.init(conn)

		if err != nil {
			log.Println("Failed to initialize channel. Retrying...")

			select {
			case <-session.done:
				return true
			case <-time.After(reInitDelay):
			}
			continue
		}

		select {
		case <-session.done:
			log.Println("Session closed. Exit ReInit.")
			return true
		case <-session.notifyConnClose:
			log.Println("Connection closed. Reconnecting...")
			return false
		case <-session.notifyChanClose:
			log.Println("Channel closed. Re-running init...")
		}
	}
}

// connect will create a new AMQP connection with given TLS config
func (session *Session) connectTLS(addr string, tlsConfig *tls.Config) (*amqp.Connection, error) {
	conn, err := amqp.DialTLS(addr, tlsConfig)

	if err != nil {
		return nil, err
	}

	session.changeConnection(conn)
	log.Println("Connected!")
	return conn, nil
}

// connect will create a new AMQP connection
func (session *Session) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)

	if err != nil {
		return nil, err
	}

	session.changeConnection(conn)
	log.Println("Connected!")
	return conn, nil
}

// init will initialize channel & declare queue
func (session *Session) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()

	if err != nil {
		return err
	}

	err = ch.Confirm(false)

	if err != nil {
		return err
	}
	_, err = ch.QueueDeclare(
		session.name,
		false, // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)

	if err != nil {
		return err
	}

	session.changeChannel(ch)
	session.isReady = true
	log.Printf("Setup! isReady: %v\n", session.isReady)

	return nil
}

// changeConnection takes a new connection to the queue,
// and updates the close listener to reflect this.
func (session *Session) changeConnection(connection *amqp.Connection) {
	session.connection = connection
	session.notifyConnClose = make(chan *amqp.Error)
	session.connection.NotifyClose(session.notifyConnClose)
}

// changeChannel takes a new channel to the queue,
// and updates the channel listeners to reflect this.
func (session *Session) changeChannel(channel *amqp.Channel) {
	session.channel = channel
	session.notifyChanClose = make(chan *amqp.Error)
	session.notifyConfirm = make(chan amqp.Confirmation, NotifyMax)
	session.clientPublishState = make(chan int, NotifyMax)
	session.channel.NotifyClose(session.notifyChanClose)
	session.channel.NotifyPublish(session.notifyConfirm)

	err := session.channel.Confirm(false)
	failOnError(err, "[S] Failed to set channel to confirm mode")
}

func (session *Session) WaitPublishConfirm() {
	for {
		confirm, confirmOK := <-session.notifyConfirm
		if confirmOK {
			_, ok := <-session.clientPublishState // if got one confirmed, client side should have already sent it.
			log.Printf("Publish Stat: len %d\n", len(session.clientPublishState))
			if ok {
				log.Printf("[S] Publish confirmed for %v", confirm)
				if len(session.clientPublishState) == 0 {
					close(session.clientPublishState)
					break
				}
			} else {
				log.Printf("[S] Error, publish state changed to finish too early\n")
				failOnError(errConfirm, "[S] Failed to declare an exchange")
			}
		}
	}
}
func (session *Session) Consume(queue string, autoACK bool) (<-chan amqp.Delivery, error) {
	msgs, err := session.channel.Consume(
		queue,   // queue
		"",      // consumer
		autoACK, // auto ack
		false,   // exclusive
		false,   // no local
		false,   // no wait
		nil,     // args
	)
	return msgs, err
}

func (session *Session) Publish(exchange string, key string, data []byte) error {
	err := session.channel.Publish(
		exchange, // exchange
		key,      // routing key
		true,     // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		})
	if err == nil {
		session.clientPublishState <- session.publishCount
		session.publishCount++

		log.Printf("[S] Client published with RK: %s, msgCount %d, len %d \n",
			key, session.publishCount, len(session.clientPublishState))
		return nil
	}
	return err
}

// Push will push data onto the queue, and wait for a confirm.
// If no confirms are received until within the resendTimeout,
// it continuously re-sends messages until a confirm is received.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (session *Session) Push(data []byte) error {
	if !session.isReady {
		return errors.New("failed to push push: not connected")
	}
	for {
		err := session.UnsafePush(data)
		if err != nil {
			session.logger.Println("Push failed. Retrying...")
			select {
			case <-session.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
			continue
		}
		// Push until confirmed
		select {

		case confirm := <-session.notifyConfirm:
			if confirm.Ack {
				session.logger.Println("Push confirmed!")
				return nil
			}
		case <-time.After(resendDelay):
		}
		session.logger.Println("Push didn't confirm. Retrying...")
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// recieve the message.
func (session *Session) UnsafePush(data []byte) error {
	if !session.isReady {
		return errNotConnected
	}
	return session.channel.Publish(
		"",           // Exchange
		session.name, // Routing key
		false,        // Mandatory
		false,        // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}

// Stream will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
func (session *Session) Stream() (<-chan amqp.Delivery, error) {
	if !session.isReady {
		return nil, errNotConnected
	}
	return session.channel.Consume(
		session.name,
		"",    // Consumer
		false, // Auto-Ack
		false, // Exclusive
		false, // No-local
		false, // No-Wait
		nil,   // Args
	)
}

// Close will cleanly shutdown the channel and connection.
func (session *Session) Close() error {
	if !session.isReady {
		return errAlreadyClosed
	}

	for len(session.clientPublishState) != 0 {
		time.Sleep(closeDelay)
		log.Printf("Wait %v for closing the session\n", closeDelay)
	}
	close(session.done) // Notify in client side, session thing is doen

	err := session.channel.Close()
	if err != nil {
		return err
	}
	err = session.connection.Close()
	if err != nil {
		return err
	}
	session.isReady = false
	log.Printf("Session Closed.")
	return nil
}

func (session *Session) confirm(noWait bool) error {
	err := session.channel.Confirm(noWait)
	return err
}

func (session *Session) QueueBind(queue string, key string, exchange string) error {
	err := session.channel.QueueBind(
		queue,    // queue name
		key,      // routing key
		exchange, // exchange
		false,    // noWait
		nil)      // amqp.Table
	return err
}

func (session *Session) QueueDeclare(name string) error {
	if !session.isReady {
		log.Println("Session is not ready and wait.")
	waitReady:
		for {
			select {
			case <-session.done:
				return errShutdown
			case <-time.After(reInitDelay):
				log.Printf("session.isReady = %v\n", session.isReady)

				if session.isReady {
					log.Printf("Waited %v and found session is ready\n", reInitDelay)
					break waitReady
				}
				log.Printf("Waited %v and found session is still not ready\n", reInitDelay)
			}
		}
	}

	_, err := session.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	return err
}

func (session *Session) ExchangeDeclare(name string, typeName string) error {
	if !session.isReady {
		log.Println("Session is not ready and wait.")
	waitReady:
		for {
			select {
			case <-session.done:
				return errShutdown
			case <-time.After(reInitDelay):
				log.Printf("session.isReady = %v\n", session.isReady)

				if session.isReady {
					log.Printf("Waited %v and found session is ready\n", reInitDelay)
					break waitReady
				}
				log.Printf("Waited %v and found session is still not ready\n", reInitDelay)
			}
		}
	}

	err := session.channel.ExchangeDeclare(
		name,     // name
		typeName, // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	return err
}

/*serverErrorReturnCh := make(chan amqp.Return, 10)
ch.NotifyReturn(serverErrorReturnCh)

go func() {
	for errorReturn := <-serverErrorReturnCh; ; {
		log.Printf("[S] Failed for a publish %s\n", errorReturn.ReplyText)
	}
}()*/
