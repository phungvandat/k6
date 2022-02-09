package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

type handler struct {
	db       *mongo.Database
	jobChn   chan interface{}
	maxBatch int
	doneChan chan struct{}
}

func NewHandler(ctx context.Context, db *mongo.Database) (*handler, func()) {
	h := &handler{
		db:       db,
		jobChn:   make(chan interface{}, 10000),
		maxBatch: 1000,
		doneChan: make(chan struct{}),
	}

	go h.dispatcher(ctx)

	return h, h.close
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			log.Printf("path: %s method: %s err: %v", r.URL.Path, r.Method, err)
		}
	}()

	ctx := context.Background()
	res := ""
	isNotFound := false

	res = r.Method + r.URL.Path

	switch r.Method {
	case http.MethodGet:
		switch r.URL.Path {
		case "/":
			h.get()
		default:
			isNotFound = true
		}

	case http.MethodPost:
		data := make(map[string]interface{})
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		switch r.URL.Path {
		case "/":
			h.post()
		case "/sync":
			h.postSync(ctx, data)
		case "/async":
			h.postAsync(ctx, data)
		case "/batch":
			h.postBatch(data)
		default:
			isNotFound = true
		}

	default:
		isNotFound = true
	}

	statusCode := http.StatusOK
	if isNotFound {
		statusCode = http.StatusNotFound
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(res))
}

// test overhead of package http
func (h *handler) get() {}

// test network latency with body payload
func (h *handler) post() {}

// sync insert log
func (h *handler) postSync(ctx context.Context, data interface{}) {
	h.insertData(ctx, data, "postSync")
}

// async insert log
func (h *handler) postAsync(ctx context.Context, data interface{}) {
	go h.insertData(ctx, data, "postAsync")
}

// async batch insert log
func (h *handler) postBatch(data interface{}) {
	go h.addDataToJob(data)
}

//------------//

func (h *handler) insertData(ctx context.Context, data interface{}, label string) {
	_, err := h.db.Collection("logs").InsertOne(ctx, data)
	if err != nil {
		log.Printf("failed %s %v\n", label, err)
	}
}

func (h *handler) insertMany(ctx context.Context, data []interface{}) {
	_, err := h.db.Collection("logs").InsertMany(ctx, data)
	if err != nil {
		log.Printf("failed insert many %v\n", err)
	}
}

func (h *handler) addDataToJob(data interface{}) {
	h.jobChn <- data
}

func (h *handler) dispatcher(ctx context.Context) {
doneLabel:
	for {
		select {
		case data := <-h.jobChn:
			listData := []interface{}{data}
		loopDataLabel:
			for {
				if len(listData) == h.maxBatch {
					break loopDataLabel
				}
				select {
				case d := <-h.jobChn:
					listData = append(listData, d)
				default:
					break loopDataLabel
				}
			}
			h.insertMany(ctx, listData)
		case <-ctx.Done():
			break doneLabel
		}
	}

	h.doneChan <- struct{}{}
}

func (h *handler) close() {
	close(h.jobChn)

	<-h.doneChan

	log.Println("job dispatcher closed")
}

func PrintJSON(val interface{}) {
	b, err := json.MarshalIndent(val, "", "\t")
	fmt.Println(string(b), err)
}
