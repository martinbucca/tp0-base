package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"io"
)

const BET_MESSAGE_ID = "BET"
const SEPARATOR = "|"


type BetSocket struct {
	conn net.Conn
}

// Constructor
func NewBetSocket(conn net.Conn) *BetSocket {
	return &BetSocket{conn: conn}
}


func serializeBet(config *ClientConfig) string {
	fields := []string{
		BET_MESSAGE_ID,
		config.ID,
		config.Name,
		config.Surname,
		config.DocumentId,
		config.BirthDate.Format("2006-01-02"),
		config.Number,
	}
	return fmt.Sprintf("%s", joinWithSeparator(fields, SEPARATOR))
}

func joinWithSeparator(fields []string, sep string) string {
	return fmt.Sprint(fields[0], sep, fields[1], sep, fields[2], sep, fields[3], sep, fields[4], sep, fields[5], sep, fields[6])
}

func (b *BetSocket) sendBet(config *ClientConfig) error {
	data := serializeBet(config)
	payload := []byte(data)
	length := uint32(len(payload))

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, length)
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

// Lee un mensaje con prefijo de longitud
func (b *BetSocket) readMessage() (string, error) {
	// Primero leemos los 4 bytes de longitud
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(b.conn, lenBuf); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint32(lenBuf)

	// Ahora leemos exactamente "length" bytes
	payload := make([]byte, length)
	if _, err := io.ReadFull(b.conn, payload); err != nil {
		return "", err
	}

	return string(payload), nil
}

func (b *BetSocket) Close() error {
	return b.conn.Close()
}
