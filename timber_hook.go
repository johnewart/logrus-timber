package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

type TimberHookLogRecord struct {
	DateTime string        `json:"dt"`
	Level    string        `json:"level"`
	Message  string        `json:"message"`
	Context  logrus.Fields `json:"context"`
}

type TimberHook struct {
	sync.Mutex
	ApiKey    string
	SourceId  string
	LogLevels []logrus.Level
	Buffer    []TimberHookLogRecord
	Interval  time.Duration
}

func (h *TimberHook) Levels() []logrus.Level {
	return h.LogLevels
}

func (h *TimberHook) Fire(e *logrus.Entry) error {
	record := TimberHookLogRecord{
		DateTime: e.Time.Format(time.RFC3339),
		Level:    e.Level.String(),
		Message:  e.Message,
		Context:  e.Data,
	}

	h.Buffer = append(h.Buffer, record)

	return nil
}

func (h *TimberHook) startBufferFlush() error {
	go func(hook *TimberHook) {
		tick := time.Tick(hook.Interval)
		for range tick {
			hook.flushToTimber()
		}
	}(h)

	return nil
}

func (h *TimberHook) flushToTimber() error {
	httpsUrl := "https://logs.timber.io"
	sourceId := h.SourceId
	apiKey := h.ApiKey

	client := &http.Client{}

	h.Lock()
	jsonData, err := json.Marshal(h.Buffer)
	h.Buffer = h.Buffer[:0]
	h.Unlock()

	if err != nil {
		// Who logs the logger .. ?
		return err
	}

	endpoint := fmt.Sprintf("%s/sources/%s/frames", httpsUrl, sourceId)
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if resp.StatusCode == 202 {
		return nil
	}

	return err
}

func NewTimberLogHook(apiKey string, sourceId string) *TimberHook {
	hook := &TimberHook{
		ApiKey:    apiKey,
		SourceId:  sourceId,
		LogLevels: logrus.AllLevels,
		Buffer:    make([]TimberHookLogRecord, 0),
		Interval:  5 * time.Second,
	}

	hook.startBufferFlush()

	return hook
}
