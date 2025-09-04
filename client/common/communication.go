package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

const BET_SEPARATOR = "|"
const CHUNK_SEPARATOR = "&"
const CHUNK_BET_MESSAGE_ID = uint16(12)
const CHUNK_FINISH_MESSAGE_ID = uint16(13)
const AGENCY_SUCCESS_MESSAGE_ID = uint16(14)
const FINISH_MESSAGE_ID = uint16(15)

const BYTES_MESSAGE_ID = 2
const BYTES_PAYLOAD_LENGTH = 2
const BYTES_CHUNK_ID_OK_MESSAGE = 4
const BYTES_CLIENT_ID_FINISH_MESSAGE = 4



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
	data := b.serializeBetsChunk(betsChunk)
	log.Infof("data: %s", data)
	payload := []byte(data)
	length := uint16(len(payload))
	log.Infof("length: %d", length)
	messageIdBuf := make([]byte, BYTES_MESSAGE_ID)
	binary.BigEndian.PutUint16(messageIdBuf, CHUNK_BET_MESSAGE_ID)

	lenBuf := make([]byte, BYTES_PAYLOAD_LENGTH)
	binary.BigEndian.PutUint16(lenBuf, length)

	if _, err := b.conn.Write(messageIdBuf); err != nil {
		return err
	}

	if _, err := b.conn.Write(lenBuf); err != nil {
		return err
	}

	totalWritten := 0
	for totalWritten < int(length) {
		n, err := b.conn.Write(payload[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += n
	}

	return nil
}

func (b *BetSocket) sendFinish() error {
	messageIdBuf := make([]byte, BYTES_MESSAGE_ID)
	binary.BigEndian.PutUint16(messageIdBuf, FINISH_MESSAGE_ID)

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



func (b *BetSocket) waitForAck(expectedChunkId int) error {
	messageIdBuf := make([]byte, BYTES_MESSAGE_ID)
	totalRead := 0
	for totalRead < BYTES_MESSAGE_ID {
		n, err := b.conn.Read(messageIdBuf[totalRead:])
		if err != nil {
			return err
		}
		totalRead += n
	}
	messageId := binary.BigEndian.Uint16(messageIdBuf)
	if messageId != AGENCY_SUCCESS_MESSAGE_ID {
		return fmt.Errorf("unexpected message ID: %d", messageId)
	}
	chunkIdBuf := make([]byte, BYTES_CHUNK_ID_OK_MESSAGE)
	totalRead = 0
	for totalRead < BYTES_CHUNK_ID_OK_MESSAGE {
		n, err := b.conn.Read(chunkIdBuf[totalRead:])
		if err != nil {
			return err
		}
		totalRead += n
	}
	chunkId := binary.BigEndian.Uint32(chunkIdBuf)
	if chunkId != uint32(expectedChunkId) {
		return fmt.Errorf("unexpected chunk ID: %d", chunkId)
	}

	return nil
}

func (b *BetSocket) waitForFinish() error {
	messageIdBuf := make([]byte, BYTES_MESSAGE_ID)
	totalRead := 0
	for totalRead < BYTES_MESSAGE_ID {
		n, err := b.conn.Read(messageIdBuf[totalRead:])
		if err != nil {
			return err
		}
		totalRead += n
	}
	messageId := binary.BigEndian.Uint16(messageIdBuf)
	if messageId != FINISH_MESSAGE_ID {
		return fmt.Errorf("unexpected message ID: %d", messageId)
	}
	clientIdBuf := make([]byte, BYTES_CLIENT_ID_FINISH_MESSAGE)
	totalRead = 0
	for totalRead < BYTES_CLIENT_ID_FINISH_MESSAGE {
		n, err := b.conn.Read(clientIdBuf[totalRead:])
		if err != nil {
			return err
		}
		totalRead += n
	}
	clientId := binary.BigEndian.Uint32(clientIdBuf)
	expectedClientId, err := strconv.Atoi(b.clientId)
	if err != nil {
		return err
	}
	if clientId != uint32(expectedClientId) {
		return fmt.Errorf("unexpected client ID: %d", clientId)
	}
	return nil
}

func (b *BetSocket) Close() error {
	return b.conn.Close()
}
