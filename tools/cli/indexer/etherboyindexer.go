package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	loom "github.com/loomnetwork/go-loom"
)

var queueName = "loomevents"
var elasticURL = "http://localhost:9200"
var elasticIndex = "etherboy"

type emitData struct {
	Caller     loom.Address `json:"caller"`
	Address    loom.Address `json:"address"`
	PluginName string       `json:"plugin"`
	Data       []byte       `json:"encodedData"`
}

type emitMsg struct {
	Owner  string
	Method string
	Addr   []byte
	Value  int64
}

type indexEntry struct {
	Plugin      string
	BlockHeight int64
	CallerAddr  []byte
	ChainAddr   []byte
	Owner       string
	Method      string
	Value       int64
}

func main() {
	c, err := redis.DialURL("redis://localhost:6379")
	if err != nil {
		log.Fatal("Cannot connect to redis")
	}
	checkpoint := getCheckPoint(c)
	log.Printf("Checkpoint: %d\n", checkpoint)
	for {
		nextCP := checkpoint + 1
		events, eof, _ := getEvents(c, nextCP)
		if eof {
			log.Println("No more events, sleeping ...")
			time.Sleep(3 * time.Second)
			continue
		}
		log.Printf("Fetched %d events for height %d\n", len(events), checkpoint)
		for _, event := range events {
			indexEvent(checkpoint, event)
			checkPointEvent(c, nextCP)
		}
		checkpoint = nextCP
		//		purgeEvents(c, checkpoint)
	}
}

func getCheckPoint(c redis.Conn) int64 {
	cp, _ := redis.Int64(c.Do("GET", "checkpoint"))
	return cp
}

func getEvents(c redis.Conn, height int64) ([]*emitData, bool, error) {
	count, err := redis.Int(c.Do("ZCOUNT", queueName, height, "+inf"))
	if err != nil {
		log.Printf("Error fetching event count: %v", err)
		return nil, false, err
	}
	if count == 0 {
		return nil, true, nil
	}
	eventVals, err := redis.Values(c.Do("ZRANGEBYSCORE",
		queueName,
		height,
		height))
	if err != nil {
		log.Printf("Unable to fetch event data: %v\n", err)
		return nil, false, err
	}
	events := []*emitData{}
	for _, ev := range eventVals {
		var ed emitData
		evBytes, ok := ev.([]byte)
		if !ok {
			log.Printf("Error typecasting event data %v", ev)
			continue
		}
		if err := json.Unmarshal(evBytes, &ed); err != nil {
			log.Printf("Error unmarshaling event data %s: %v", ev, err)
			continue
		}
		events = append(events, &ed)
	}
	return events, false, nil
}

func indexEvent(height int64, event *emitData) {
	callerAddr := event.Caller.Local
	chainAddr := event.Address.Local
	plugin := event.PluginName
	var msg emitMsg
	json.Unmarshal(event.Data, &msg)
	owner := msg.Owner
	method := msg.Method
	value := msg.Value
	indexEntry := &indexEntry{
		Plugin:      plugin,
		BlockHeight: height,
		CallerAddr:  callerAddr,
		ChainAddr:   chainAddr,
		Owner:       owner,
		Method:      method,
		Value:       value,
	}
	indexJSON, err := json.Marshal(indexEntry)
	if err != nil {
		log.Printf("Error marshalling index json: %v", err)
		return
	}
	elasticResourceURL := fmt.Sprintf("%s/%s/%s/", elasticURL, elasticIndex, "app")
	log.Println(elasticResourceURL)
	resp, err := http.Post(elasticResourceURL, "application/json", bytes.NewReader(indexJSON))
	if err != nil {
		log.Printf("Error writing to index: %v\n", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("%+v", indexEntry)
	return
}

func checkPointEvent(c redis.Conn, cp int64) error {
	c.Do("SET", "checkpoint", cp)
	return nil
}

func purgeEvents(c redis.Conn, cp int64) error {
	return nil
}
