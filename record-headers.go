package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type RequestHandler struct{}

func (rh *RequestHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	if req.URL.Path != "/" {
		res.WriteHeader(http.StatusNotFound)
		res.Header().Set("Cache-control", "public, max-age=0")
		res.Header().Set("Content-Type", "text/html")
		res.Header().Set("Server", "GoBench Header Recorder")
		fmt.Fprintf(res, "Invalid path")
		return
	}

	res.Header().Set("Cache-control", "public, max-age=0")
	res.Header().Set("Content-Type", "text/html")
	res.Header().Set("Server", "GoBench Header Recorder")

	// Store all the headers from the current request in header map
	headers := make(map[string]interface{})
	for k, v := range req.Header {
		headers[k] = v[0]
	}

	output, _ := json.Marshal(response)
	fmt.Fprintf(res, string(output))

}

func main() {

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		err := http.ListenAndServe("0.0.0.0:8088", &RequestHandler{})
		if err != nil {
			fmt.Println("Record headers Error:", err)
			os.Exit(0)
		}
		wg.Done()
	}()

	fmt.Println("[" + dateStampAsString() + "] Logging server started on 0.0.0.0:8088")

	wg.Wait()

}
