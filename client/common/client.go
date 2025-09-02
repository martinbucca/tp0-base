package common

import (
	"net"
	"time"
	"os"
	"os/signal"
	"syscall"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")
const AGENCY_SUCCESS_MESSAGE = "OK"

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	Name		  string
	Surname      string
	DocumentId   string
	BirthDate    time.Time
	Number       string
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	betSocket   *BetSocket
	is_currently_running bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		is_currently_running: true,
	}
	_ = setupSigtermHandler(client)
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createBetSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.betSocket = NewBetSocket(conn)
	return nil
}

func setupSigtermHandler(c *Client) <-chan os.Signal {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM)
	go handleSigterm(c, sigChannel)
	return sigChannel
}

func handleSigterm(c *Client, sigCh <-chan os.Signal) {
    <-sigCh
    log.Infof("action: shutdown | result: in_progress | reason: signal received")
    c.is_currently_running = false
    if c.betSocket != nil {
        c.betSocket.Close()
    }
}


// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	if err := c.createBetSocket(); err != nil {
		return
	}
	err := c.betSocket.sendBet(&c.config)
	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v | error: %v",
			c.config.DocumentId,
			c.config.Number,
			err,
		)
		return
	}

	msg, err := c.betSocket.readMessage()
	if err != nil {
		log.Errorf("action: respuesta_recibida | result: fail | dni: %v | numero: %v | error: %v",
			c.config.DocumentId,
			c.config.Number,
			err,
		)
		return
	}

	if msg == AGENCY_SUCCESS_MESSAGE {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			c.config.DocumentId,
			c.config.Number,
		)
	}
	c.betSocket.Close()
}
