package telemetry

import (
	"encoding/csv"
	"log"
	"log/slog"
	"os"
)

type EventType string

type Record interface {
	Row() []string
}

type Telemetry struct {
	fd      *os.File
	writer  *csv.Writer
	telChan chan Record
	endChan chan struct{}
}

func New(outFileName string, headers []string) (*Telemetry, error) {
	fd, err := os.OpenFile(outFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(fd)
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	slog.Info("output telemetry file opened.", "file", outFileName)

	telChan := make(chan Record, 10)
	endChan := make(chan struct{}, 10)

	tel := &Telemetry{fd, writer, telChan, endChan}

	go tel.collect()

	return tel, nil
}

func (t *Telemetry) Collect(record Record) {
	t.telChan <- record
}

func (t *Telemetry) Done() {
	t.endChan <- struct{}{}

	t.writer.Flush()
	if err := t.writer.Error(); err != nil {
		panic(err)
	}
	t.fd.Close()

}

func (t *Telemetry) collect() {

	for {
		select {
		case record := <-t.telChan:
			log.Println("collected", record.Row())
			if err := t.writer.Write(record.Row()); err != nil {
				slog.Error("failed telemetry write", "err", err)
			}

		case <-t.endChan:
			return
		}
	}
}
