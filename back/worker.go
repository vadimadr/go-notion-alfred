package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Worker struct {
	Env        EnvSetting
	QueryCache map[string]QCacheEntry
	cQuery     chan string
	cDone      chan bool
}

type QCacheEntry struct {
	CreatedAt time.Time
	Response  []ResponseItem
	Query     string
}

func CreateWorker() *Worker {
	var w Worker
	w.Env = ReadEnv()
	w.QueryCache = make(map[string]QCacheEntry)
	w.cQuery = make(chan string)
	w.cDone = make(chan bool)

	go w.UpdateCacheWorker()

	return &w
}

func (w *Worker) Search(query string) []byte {
	query = NormalizeQueryString(query)
	entry, found := w.QueryCache[query]
	if found {
		log.Printf("Cache hit: %s", query)
		w.cQuery <- query
		expired := time.Now().Sub(entry.CreatedAt) > 30*time.Second
		if expired {
			w.cQuery <- query
		}
		return SerializeSearchResp(entry.Response, expired)
	}

	// not in qury cahce, perform search directly
	return w.RunQuery(query)
}

func NormalizeQueryString(query string) string {
	return query
}

func (w *Worker) UpdateCacheWorker() {
	for {
		select {
		case q := <-w.cQuery:
			log.Printf("Updating cache for: %s", q)
			go w.RunQuery(q)
		case <-w.cDone:
			return
		}
	}
}

func (w *Worker) RunQuery(q string) []byte {
	begin := time.Now()
	log.Print(fmt.Sprintf("Running qury: %s", q))
	resp := NotionSearch(w.Env, q)
	parsed := NotionParseSearchResponse(resp)

	// update cache
	w.QueryCache[q] = QCacheEntry{time.Now(), parsed, q}

	elapsed := time.Now().Sub(begin)
	log.Println("Done query ", q, elapsed)
	return SerializeSearchResp(parsed, false)
}

func SerializeSearchResp(items []ResponseItem, rerun bool) []byte {
	var aw AlfredFeedback

	for _, item := range items {
		newItem := AlfredItem{item.Title, item.Subtitle}
		aw.Items = append(aw.Items, newItem)
	}

	if rerun {
		aw.Rerun = 0.2
	}

	bytes, err := json.Marshal(&aw)
	HandleError(err)
	return bytes
}
