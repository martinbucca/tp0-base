package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)


const FILEPATH = "/agency.csv"

type BetsChunk struct {
	Bets []*Bet
	Id   string
}

type Bet struct {
	Name      string
	Surname   string
	DocumentId  string
	Birthdate string
	Number    int
}

func NewBetFromCSVLine(record []string) (*Bet, error) {
	if len(record) != 5 {
		return nil, fmt.Errorf("Invalid CSV record: %v", record)
	}


	name := record[0]
	surname := record[1]
	documentId := record[2]
	birthDate := record[3]

	number, err := strconv.Atoi(record[4])
	if err != nil {
		return nil, err
	}

	return &Bet{
		Name:     name,
		Surname:  surname,
		DocumentId: documentId,
		Birthdate: birthDate,
		Number:   number,
	}, nil
}

type CSVReader struct {
	file   *os.File
	reader *csv.Reader
}

func NewCSVReader() (*CSVReader, error) {
	file, err := os.Open(FILEPATH)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(file)
	return &CSVReader{file: file, reader: reader}, nil
}

func (r *CSVReader) ReadChunk(chunkId string, maxAmount int) (*BetsChunk, error) {
	var bets []*Bet
	for i := 0; i < maxAmount; i++ {
		record, err := r.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		bet, err := NewBetFromCSVLine(record)
		if err != nil {
			return nil, err
		}
		bets = append(bets, bet)
	}

	return &BetsChunk{Bets: bets, Id: chunkId}, nil
}

func (r *CSVReader) Close() {
	if r.file != nil {
		r.file.Close()
	}
}
