package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"io"
)

const BET_SEPARATOR = "|"
const CHUNK_SEPARATOR = "&"
const CHUNK_BET_MESSAGE_ID = uint8(12)
const CHUNK_FINISH_MESSAGE_ID = uint8(13)
const BYTES_MESSAGE_ID = 1
const BYTES_PAYLOAD_LENGTH = 2


type BetSocket struct {
	clientId string
	conn net.Conn
}

// Constructor
func NewBetSocket(ServerAddress string, clientId string) (*BetSocket, error) {
	conn, err := net.Dial("tcp", ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | error: %v",
			err,
		)
		return nil, err
	}
	return &BetSocket{conn: conn, clientId: clientId}, nil
}


func joinWithSeparator(fields []string, sep string) string {
	amountFields := len(fields)
	if amountFields == 0 {
		return ""
	}

	result := fields[0]
	for i := 1; i < amountFields; i++ {
		result += sep + fields[i]
	}
	return result
}

func serializeBet(bet *Bet) string {
	fields := []string{
		bet.Name,
		bet.Surname,
		bet.DocumentId,
		bet.Birthdate,
		fmt.Sprintf("%d", bet.Number),
	}
	return fmt.Sprintf("%s", joinWithSeparator(fields, BET_SEPARATOR))
}

func (b *BetSocket) serializeBetsChunk(betsChunk *BetsChunk) string {
	fields := []string{
		b.clientId,
		betsChunk.Id,
	}

	bets := betsChunk.Bets
	for _, bet := range bets {
		serializedBet := serializeBet(bet)
		fields = append(fields, serializedBet)
	}

	return fmt.Sprintf("%s", joinWithSeparator(fields, CHUNK_SEPARATOR))
}



func (b *BetSocket) sendBet(betsChunk *BetsChunk) error {
	data := serializeBetsChunk(betsChunk)
	payload := []byte(data)
	length := uint16(len(payload))

	messageIdBuf := make([]byte, BYTES_MESSAGE_ID)
	binary.BigEndian.PutUint8(messageIdBuf, CHUNK_BET_MESSAGE_ID)

	lenBuf := make([]byte, BYTES_PAYLOAD_LENGTH)
	binary.BigEndian.PutUint16(lenBuf, length)

	totalWritten := 0
	for totalWritten < int(length) {
		if totalWritten < BYTES_MESSAGE_ID {
			// Write the message ID (1 byte, big endian)
			n, err := b.conn.Write(messageIdBuf)
			if err != nil {
				return err
			}
			totalWritten += n
		} else if totalWritten > BYTES_MESSAGE_ID && totalWritten < BYTES_PAYLOAD_LENGTH {
			// Write the length prefix (2 bytes, big endian)
			n, err := b.conn.Write(lenBuf[totalWritten:])
			if err != nil {
				return err
			}
			totalWritten += n
		} else {
			// Write the payload (the bets)
			n, err := b.conn.Write(payload[totalWritten:])
			if err != nil {
				return err
			}
			totalWritten += n
		}
	}

	return nil
}

func (b *BetSocket) sendFinish() error {
	messageIdBuf := make([]byte, BYTES_MESSAGE_ID)
	binary.BigEndian.PutUint8(messageIdBuf, CHUNK_FINISH_MESSAGE_ID)

	totalWritten := 0
	for totalWritten < BYTES_MESSAGE_ID {
		n, err := b.conn.Write(messageIdBuf[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += n
	}

	return nil
}

// Lee un mensaje con prefijo de longitud
func (b *BetSocket) readMessage() (string, error) {
	// Primero leemos los 4 bytes de longitud
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(b.conn, lenBuf); err != nil {
		log.Errorf("action: read_message_length | result: fail | error: %v", err)
		return "", err
	}
	length := binary.BigEndian.Uint32(lenBuf)

	// Ahora leemos exactamente "length" bytes
	payload := make([]byte, length)
	if _, err := io.ReadFull(b.conn, payload); err != nil {
		log.Errorf("action: read_message_payload | result: fail | error: %v", err)
		return "", err
	}

	return string(payload), nil
}

func (b *BetSocket) Close() error {
	return b.conn.Close()
}
