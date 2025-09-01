package common

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	is_currently_running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		is_currently_running: true,
	}
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM)
	go func() {
		<-sigChannel
		log.Infof("action: shutdown | result: in_progress | reason: signal received")
		client.is_currently_running = false
		if client.conn != nil {
			client.conn.Close()
		}
	}()
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

func handleSigterm(c *Client, sigCh <-chan os.Signal) {
    <-sigCh
    log.Infof("action: shutdown | result: in_progress | reason: signal received")
    c.is_currently_running = false
    if c.conn != nil {
        c.conn.Close()
    }
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; c.is_currently_running && msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		if err := c.createClientSocket(); err != nil { return }
		// TODO: Modify the send to avoid short-write
		fmt.Fprintf(
			c.conn,
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	if c.is_currently_running {
		log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
	}
}
