package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const BET_SEPARATOR = "|"
const CHUNK_SEPARATOR = "&"
const CHUNK_BET_MESSAGE_ID = uint16(12)
const ACK_CHUNK_BET_MESSAGE_ID = uint16(13)
const FINISH_MESSAGE_ID = uint16(14)
const GET_WINNERS_MESSAGE_ID = uint16(15)
const NO_WINNERS_MESSAGE_ID = uint16(16)
const WINNERS_RESULT_MESSAGE_ID = uint16(17)

const BYTES_MESSAGE_ID = 2
const BYTES_PAYLOAD_LENGTH = 2
const BYTES_CHUNK_ID_OK_MESSAGE = 4
const BYTES_CLIENT_ID_FINISH_MESSAGE = 4
const BYTES_CLIENT_ID_WINNERS_MESSAGE = 4



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


func (b *BetSocket) writeFull(buf []byte) error {
	totalWritten := 0
	for totalWritten < len(buf) {
		n, err := b.conn.Write(buf[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += n
	}
	return nil
}

func (b *BetSocket) readFull(buf []byte) error {
	totalRead := 0
	for totalRead < len(buf) {
		n, err := b.conn.Read(buf[totalRead:])
		if err != nil {
			return err
		}
		totalRead += n
	}
	return nil
}

func makeMessageIDBuf(messageID uint16) []byte {
	buf := make([]byte, BYTES_MESSAGE_ID)
	binary.BigEndian.PutUint16(buf, messageID)
	return buf
}

func (b *BetSocket) sendBets(betsChunk *BetsChunk) error {
	data := b.serializeBetsChunk(betsChunk)
	payload := []byte(data)
	length := uint16(len(payload))

	if err := b.writeFull(makeMessageIDBuf(CHUNK_BET_MESSAGE_ID)); err != nil {
		return err
	}

	lenBuf := make([]byte, BYTES_PAYLOAD_LENGTH)
	binary.BigEndian.PutUint16(lenBuf, length)
	if err := b.writeFull(lenBuf); err != nil {
		return err
	}

	return b.writeFull(payload)
}

func (b *BetSocket) sendFinish() error {
	if err := b.writeFull(makeMessageIDBuf(FINISH_MESSAGE_ID)); err != nil {
		return err
	}

	clientIdInt, err := strconv.Atoi(b.clientId)
	if err != nil {
		return err
	}
	clientIdBuf := make([]byte, BYTES_CLIENT_ID_FINISH_MESSAGE)
	binary.BigEndian.PutUint32(clientIdBuf, uint32(clientIdInt))
	return b.writeFull(clientIdBuf)
}

func (b *BetSocket) sendGetWinners() error {
	if err := b.writeFull(makeMessageIDBuf(GET_WINNERS_MESSAGE_ID)); err != nil {
		return err
	}

	clientIdInt, err := strconv.Atoi(b.clientId)
	if err != nil {
		return err
	}
	clientIdBuf := make([]byte, BYTES_CLIENT_ID_WINNERS_MESSAGE)
	binary.BigEndian.PutUint32(clientIdBuf, uint32(clientIdInt))
	return b.writeFull(clientIdBuf)
}

func (b *BetSocket) waitForAck(expectedChunkId int) error {
	messageIdBuf := make([]byte, BYTES_MESSAGE_ID)
	if err := b.readFull(messageIdBuf); err != nil {
		return err
	}
	messageId := binary.BigEndian.Uint16(messageIdBuf)
	if messageId != ACK_CHUNK_BET_MESSAGE_ID {
		return fmt.Errorf("unexpected message ID: %d", messageId)
	}

	// chunk id
	chunkIdBuf := make([]byte, BYTES_CHUNK_ID_OK_MESSAGE)
	if err := b.readFull(chunkIdBuf); err != nil {
		return err
	}
	chunkId := binary.BigEndian.Uint32(chunkIdBuf)
	if chunkId != uint32(expectedChunkId) {
		return fmt.Errorf("unexpected chunk ID: %d", chunkId)
	}

	return nil
}

func (b *BetSocket) waitForFinish() error {
	// message id
	messageIdBuf := make([]byte, BYTES_MESSAGE_ID)
	if err := b.readFull(messageIdBuf); err != nil {
		return err
	}
	messageId := binary.BigEndian.Uint16(messageIdBuf)
	if messageId != FINISH_MESSAGE_ID {
		return fmt.Errorf("unexpected message ID: %d", messageId)
	}

	// client id
	clientIdBuf := make([]byte, BYTES_CLIENT_ID_FINISH_MESSAGE)
	if err := b.readFull(clientIdBuf); err != nil {
		return err
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

func (b *BetSocket) deserializeWinners(payload []byte) []string {
	return strings.Split(string(payload), ",")
}

func (b *BetSocket) waitForWinners() ([]string, error) {
	messageIdBuf := make([]byte, BYTES_MESSAGE_ID)
	if err := b.readFull(messageIdBuf); err != nil {
		return nil, err
	}
	messageId := binary.BigEndian.Uint16(messageIdBuf)

	if messageId == NO_WINNERS_MESSAGE_ID {
		return nil, fmt.Errorf("no winners found")
	}

	if messageId != WINNERS_RESULT_MESSAGE_ID {
		return nil, fmt.Errorf("unexpected message ID: %d", messageId)
	}

	lengthBuf := make([]byte, BYTES_PAYLOAD_LENGTH)
	if err := b.readFull(lengthBuf); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, fmt.Errorf("unexpected payload length: %d", length)
	}

	payloadBuf := make([]byte, length)
	if err := b.readFull(payloadBuf); err != nil {
		return nil, err
	}
	return b.deserializeWinners(payloadBuf), nil
}

func (b *BetSocket) Close() error {
	return b.conn.Close()
}
