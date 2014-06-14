package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	//"strings"
	"flag"
	"sync"
	"time"
)

const (
	DEBUG                 = true
	DEFAULT_UA            = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/33.0.1750.152 Safari/537.36"
	VERSION_MAJOR  int    = 1
	VERSION_MINOR  int    = 1
	VERSION_PATCH  int    = 0
	VERSION_SUFFIX string = ""
)

type RequestHandler struct{}

var sampleRate float64
var total int64
var count int64
var header_list []map[string]string

func getVersion() string {
	return strconv.Itoa(VERSION_MAJOR) + "." + strconv.Itoa(VERSION_MINOR) + "." + strconv.Itoa(VERSION_PATCH) + "-" + VERSION_SUFFIX
}

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
	headers := map[string]string{}
	var parts []string
	for k, v := range req.Header {
		headers[k] = v[0]
		parts = append(parts, k+": "+v[0])
	}

	if count < total {

		if rand.Float64() <= sampleRate {
			fmt.Println("Adding to headers!")
			header_list = append(header_list, headers)
			count++
		}
	} else {

		hList, _ := json.Marshal(header_list)
		_ = writeToFile("playback_headers.txt", string(hList))
		fmt.Println("Header collection complete.\n")
		fmt.Println("Headers written to playback_headers.txt\n")
		os.Exit(0)
	}

}

func showCommandUsage(desc map[string]string) {

	fmt.Println("Command usage:")
	for k, v := range desc {
		fmt.Println("\t-" + k + " : " + v)
	}
	fmt.Println("")
	os.Exit(0)
}

func main() {

	description := map[string]string{
		"t": "Total number of headers to record (default 50)",
		"s": "Sample rate x.y - (1.0 <= x > 0.0, default 0.2)",
	}

	if len(os.Args) == 2 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			fmt.Println("GoBench Header Recorder - Version " + getVersion() + "\n")
			os.Exit(0)
		} else if os.Args[1] == "-h" || os.Args[1] == "--help" {
			showCommandUsage(description)
		}
	} else if len(os.Args) == 1 {
		showCommandUsage(description)
	}

	flag.Int64Var(&total, "t", 50, description["t"])
	flag.Float64Var(&sampleRate, "s", 0.1, description["s"])

	flag.Parse()

	if sampleRate > 1.0 {
		sampleRate = 1.0
	}

	if sampleRate == 0.0 {
		sampleRate = 0.2
	}

	if total < 1 {
		total = 25
	}

	if total >= 100000 {
		total = 100000
	}

	rand.Seed(total)

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
