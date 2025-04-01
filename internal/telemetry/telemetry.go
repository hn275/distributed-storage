package telemetry

import (
	"encoding/csv"
	"log"
	"log/slog"
	"os"
	"sync"
)

type EventType string

type Record interface {
	Row() []string
}

type Telemetry struct {
	fd     *os.File
	writer *csv.Writer
	mtx    *sync.Mutex
}

func init() {
	for _, v := range []string{"cluster", "lb", "user"} {
		if err := os.MkdirAll("tmp/output/"+v, 0755); err != nil {
			panic(err)
		}
	}
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

	tel := &Telemetry{fd, writer, new(sync.Mutex)}

	return tel, nil
}

func (t *Telemetry) Collect(record Record) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	if err := t.writer.Write(record.Row()); err != nil {
		panic(err)
	}
}

func (t *Telemetry) Done() {
	t.writer.Flush()
	if err := t.writer.Error(); err != nil {
		log.Println(err)
	}
	if err := t.fd.Close(); err != nil {
		log.Println(err)
	}
}
