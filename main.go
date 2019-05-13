package main

import (
	"fmt"
	"github.com/johnewart/logrus-timber/timberlog"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {

	apiKey, exists := os.LookupEnv("TIMBER_API_KEY")
	if !exists {
		fmt.Println("Unable to run, no TIMBER_API_KEY set")
		os.Exit(1)
	}

	sourceId, exists := os.LookupEnv("TIMBER_SOURCE_ID")
	if !exists {
		fmt.Println("Unable to run, no TIMBER_SOURCE_ID set")
		os.Exit(1)
	}

	l := logrus.New()

	l.Out = os.Stdout

	logFields := logrus.Fields{
		"os":       "linux",
		"core":     "4.21.5-ac",
		"platform": "linux-amd64",
		"host":     "hex",
		"cpus":     4,
	}

	hook := timberlog.NewTimberLogHook(apiKey, sourceId)

	l.AddHook(hook)
	e := l.WithFields(logFields)

	go func() {
		i := 1
		tick := time.Tick(500 * time.Millisecond)
		for range tick {
			e.Infof("Message #: %d", i)
			i = i + 1
		}
	}()

	for {
		time.Sleep(10 * time.Second)
	}

}
