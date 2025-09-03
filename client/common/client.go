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

type ClientConfig struct {
	ID            string
	ServerAddress string
	MaxBatchAmount int
}

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
	betSocket, err := NewBetSocket(c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.betSocket = betSocket
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
	if !c.is_currently_running {
		return
	}
	

	csvReader, err := NewCSVReader()
	if err != nil {
		log.Criticalf("Could not open CSV file: %v", err)
		return
	}
	defer csvReader.Close()

	if err := c.createBetSocket(); err != nil {
		return
	}
	chunkID := 0
	for c.is_currently_running {
		chunk, err := csvReader.ReadChunk(fmt.Sprintf("%d", chunkID), c.config.MaxBatchAmount)
		if err != nil {
			log.Errorf("action: read_chunk | result: fail | error: %v", err)
			return
		}
		if err := c.betSocket.sendBet(chunk); err != nil {
			log.Errorf("action: send_message | result: fail | error: %v", err)
			return
		}
		// Wait to receive an ack with the chunk Id
		chunkID++
	}
	if c.betSocket != nil  {
		c.betSocket.Close()
	}
}
