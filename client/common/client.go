package common

import (
	"os"
	"os/signal"
	"syscall"
	"github.com/op/go-logging"
	"fmt"
	"time"
)

var log = logging.MustGetLogger("log")
const AGENCY_SUCCESS_MESSAGE = "OK"
const MAX_AMOUNT_ALLOWED = 150

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


func (c *Client) createBetSocket() error {
	betSocket, err := NewBetSocket(c.config.ServerAddress, c.config.ID)
	if err != nil {
		return err
	}
	c.betSocket = betSocket
	return nil
}


func (c *Client) getWinners() ([]string, error) {
	var winners []string
	var err error
	retryInterval := 5 * time.Second

	for {
		if err := c.createBetSocket(); err != nil {
			log.Criticalf(
				"action: connect | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		if err = c.betSocket.sendGetWinners(); err != nil {
			return nil, err
		}

		winners, err = c.betSocket.waitForWinners()
		if err == nil {
			return winners, nil
		}

		if err := c.betSocket.Close(); err != nil {
			log.Errorf("action: close_socket | result: fail | error: %v", err)
		}

		time.Sleep(retryInterval)
		if c.betSocket != nil {
			c.betSocket.Close()
		}

	}
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	if !c.is_currently_running {
		return
	}
	

	csvReader, err := NewCSVReader()
	if err != nil {
		log.Criticalf("action: open_csv | result: fail | error: %v", err)
		return
	}
	log.Infof("action: open_csv | result: success | client_id: %v", c.config.ID)
	defer csvReader.Close()

	if err := c.createBetSocket(); err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}
	chunkID := 0
	maxBatchAmount := c.config.MaxBatchAmount
	if maxBatchAmount > MAX_AMOUNT_ALLOWED {
		maxBatchAmount = MAX_AMOUNT_ALLOWED
	}
	for c.is_currently_running {
		chunk, err := csvReader.ReadChunk(fmt.Sprintf("%d", chunkID), maxBatchAmount)
		if err != nil {
			log.Errorf("action: read_chunk | result: fail | error: %v", err)
			return
		}
		if len(chunk.Bets) == 0 {
			if err := c.betSocket.sendFinish(); err != nil {
				log.Errorf("action: send_finish | result: fail | error: %v", err)
			}
			log.Infof("action: send_finish | result: success | client id: %v", c.config.ID)
			break
		}
		if err := c.betSocket.sendBets(chunk); err != nil {
			log.Errorf("action: send_message | result: fail | error: %v", err)
			return
		}
		log.Infof("action: send_bet | result: success | cantidad: %d | client_id: %v | chunk_id: %d", len(chunk.Bets), c.config.ID, chunkID)
		if err := c.betSocket.waitForAck(chunkID); err != nil {
			log.Errorf("action: wait_for_ack | result: fail | error: %v", err)
		}

		log.Infof("action: apuesta_enviada | result: success | cantidad: %d | chunk_id: %d", len(chunk.Bets), chunkID)
		chunkID++
	}

	if err := c.betSocket.waitForFinish(); err != nil {
		log.Errorf("action: wait_for_finish | result: fail | error: %v", err)
		return
	}

	if c.betSocket != nil {
		if err := c.betSocket.Close(); err != nil {
			log.Errorf("action: close_socket | result: fail | error: %v", err)
			return
		}
	}

	log.Infof("action: ack_for_finish | result: success")

	winners, err := c.getWinners()
	if err != nil {
		log.Errorf("action: get_winners | result: fail | error: %v", err)
		return
	}

	log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d | ganadores: %v", len(winners), winners)

	log.Infof("action: exit | result: success")
	
}
