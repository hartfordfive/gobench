package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DEBUG = true
)

type RequestHandler struct{}

func dateStampAsString() string {
	t := time.Now()
	return ymdToString() + " " + fmt.Sprintf("%02d", t.Hour()) + ":" + fmt.Sprintf("%02d", t.Minute()) + ":" + fmt.Sprintf("%02d", t.Second())
}

func ymdToString() string {
	t := time.Now()
	y, m, d := t.Date()
	return strconv.Itoa(y) + fmt.Sprintf("%02d", m) + fmt.Sprintf("%02d", d)
}

func writeToFile(filePath string, dataToDump string) int {

	fh, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0640)
	if err != nil {
		//panic(err)
		fh, _ = os.Create(filePath)
		if DEBUG {
			fmt.Println("File doesn't exist.  Creating it.")
		}
	} else {
		if DEBUG {
			fmt.Println("Appending to log file.")
		}
	}
	defer fh.Close()
	nb, _ := fh.WriteString(string(dataToDump))
	fh.Sync()
	if DEBUG {
		fmt.Println("Wrote " + strconv.Itoa(nb) + " bytes to " + filePath)
	}
	return nb
}

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
	headers := make(map[string]string)
	var parts []string
	for k, v := range req.Header {
		headers[k] = v[0]
		parts = append(parts, k+": "+v[0])
	}

	_ = writeToFile("playback_headers.txt", strings.Join(parts, "~")+"\n")

	output, _ := json.Marshal(headers)
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
