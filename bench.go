package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	//"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DEFAULT_UA            = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/33.0.1750.152 Safari/537.36"
	VERSION_MAJOR  int    = 1
	VERSION_MINOR  int    = 1
	VERSION_PATCH  int    = 0
	VERSION_SUFFIX string = ""
)

type BenchOptions struct {
	Method      string
	UserAgent   string
	Header      map[string]string
	HeaderList  []map[string]string
	Timeout     int
	PostData    map[string]string
	TimeBetween int64
	TotalTests  int
	Concurency  int
	UrlList     []string
	UrlListLen  int
	Url         string
	Cookies     []http.Cookie
	UserAgents  []string
	ApiKey      string
}

type BenchTest struct {
	TimeTaken       int64
	StatusCode      int
	Url             string
	BytesDownloaded int64
}

type BenchStats struct {
	TestCount        int32
	TotalTime        int64
	TestsExecuted    int32
	TestTime         []int
	AvgTime          int32
	MedianTime       int
	UrlListTestCount map[string]int
	StatusCode       map[string]int32
	RespCF           int32
	Resp2xx          int32
	Resp3xx          int32
	Resp4xx          int32
	Resp5xx          int32
	TestStart        int64
	TestEnd          int64
	NumFail          int32
	NumPass          int32
	BytesDownloaded  int
	ServerType       string
	Lock             *sync.Mutex
}

func getVersion() string {
	return strconv.Itoa(VERSION_MAJOR) + "." + strconv.Itoa(VERSION_MINOR) + "." + strconv.Itoa(VERSION_PATCH) + "-" + VERSION_SUFFIX
}

func SortResponseTimes(sl []int) []int {
	sort.Sort(sort.IntSlice(sl))
	return sl
}

func FromNanoToMilli(ts int64) int64 {
	return (ts / 1000000)
}

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

const debug bool = false

var options = BenchOptions{
	Method:      "GET",
	UserAgent:   "Mozilla/5.0 GoBench",
	Timeout:     2,
	TimeBetween: 1000,
	TotalTests:  30,
	UrlList:     []string{},
	UserAgents:  []string{},
	ApiKey:      "",
}

func dumpToReportFile(bs *BenchStats, bo *BenchOptions, fileNamePrefix string) []string {

	ts := int(time.Now().Unix())
	y, m, d := time.Now().Date()

	files := []string{fileNamePrefix + "_general_report_" + strconv.Itoa(y) + m.String() + strconv.Itoa(d) + "_" + strconv.Itoa(ts) + ".txt",
		fileNamePrefix + "_url_hit_report_" + strconv.Itoa(y) + m.String() + strconv.Itoa(d) + "_" + strconv.Itoa(ts) + ".txt",
		fileNamePrefix + "_time_report_" + strconv.Itoa(y) + m.String() + strconv.Itoa(d) + "_" + strconv.Itoa(ts) + ".txt"}

	fh1, err1 := os.Create(files[0])
	check(err1)
	fh2, err2 := os.Create(files[1])
	check(err2)
	fh3, err3 := os.Create(files[2])
	check(err3)
	defer fh1.Close()
	defer fh2.Close()
	defer fh3.Close()

	var out string
	out = "\nStress Testing Report\n"
	out += "Num CPU cores used: " + strconv.Itoa(bo.Concurency) + "\n"

	if bo.UrlListLen >= 1 {
		out += "Total distinct URLs: " + strconv.Itoa(bo.UrlListLen) + "\n"
	} else {
		out += "URL tested: " + bo.Url + "\n"
	}

	out += "Server Type: " + bs.ServerType + "\n\n"

	out += "Total tests: " + strconv.Itoa(int(bs.TestCount)) + "\n"
	out += "Total bytes downloaded: " + strconv.Itoa(bs.BytesDownloaded) + "\n"
	out += "Total passed: " + strconv.Itoa(int(bs.NumPass)) + "\n"
	out += "Total failed: " + strconv.Itoa(int(bs.NumFail)) + "\n"
	out += "\t2xx responses: " + strconv.Itoa(int(bs.Resp2xx)) + "\n"
	out += "\t3xx responses: " + strconv.Itoa(int(bs.Resp3xx)) + "\n"
	out += "\t4xx responses: " + strconv.Itoa(int(bs.Resp4xx)) + "\n"
	out += "\t5xx responses: " + strconv.Itoa(int(bs.Resp5xx)) + "\n\n"

	out += "Shortest time: " + strconv.Itoa(bs.TestTime[0]) + " ms\n"
	out += "Longest time: " + strconv.Itoa(bs.TestTime[len(bs.TestTime)-1]) + " ms\n"
	out += "Median time: " + strconv.Itoa(bs.MedianTime) + " ms\n"
	out += "Average time: " + strconv.Itoa(int(bs.AvgTime)) + " ms\n\n"

	totalBytes := 0

	// --------------- Write the 1st report file
	nb, _ := fh1.WriteString(out)
	totalBytes += nb
	fh1.Sync()

	// --------------- Write the 2nd report file with the number of hits to each url
	out = "URL,Hits\n"
	for k, v := range bs.UrlListTestCount {
		out += k + "," + strconv.Itoa(v) + "\n"
	}
	nb, _ = fh2.WriteString(out)
	totalBytes += nb

	fh2.Sync()

	// --------------- Write the 3rd report file with the number of hits to each url
	out = ""
	lenTestTimes := len(bs.TestTime)
	for i := int32(0); i < bs.TestCount; i++ {
		if i < int32(lenTestTimes) {
			out += strconv.Itoa(bs.TestTime[i])
			if i < (bs.TestCount - 1) {
				out += ","
			}
		} else {
			fmt.Println("\tWarning: Test time", i, "not set")
		}
	}
	nb, _ = fh3.WriteString(out)
	totalBytes += nb
	fh3.Sync()

	if totalBytes > 1 {
		return files
	}

	return nil
}

func loadCookieData(inFile string) []http.Cookie {

	var cookies []http.Cookie

	// Extract the valid properties of http.Cookie with Reflection
	validCookieAttrs := map[string]int{"name": 1, "value": 1, "path": 1, "domain": 1, "expires": 1, "rawexpires": 1, "maxage": 1, "secure": 1, "httponly": 1, "raw": 1, "unparsed": 1}

	f, err := os.Open(inFile)
	if err == nil {
		r := bufio.NewReader(f)
		for s, e := Readln(r); e == nil; s, e = Readln(r) {
			if s == "" || len(strings.Trim(s, " ")) < 2 {
				goto goreturn2
			}

			parts := strings.Split(s, "~")
			cookie := http.Cookie{}
			ps := reflect.ValueOf(&cookie)
			s := ps.Elem() // Extract the struct itself

			for i := 0; i < len(parts); i++ {
				cookieParts := strings.Split(parts[i], "=")
				_, ok := validCookieAttrs[strings.Trim(cookieParts[0], " ")]
				if ok {
					f := s.FieldByName(strings.ToUpper(cookieParts[0][0:1]) + cookieParts[0][1:])
					f.SetString(cookieParts[1])
				}
			}

			cookies = append(cookies, cookie)
		}
	goreturn2:
		return cookies
	}
	return nil

}

func loadPostData(inFile string) map[string]string {
	pd := make(map[string]string)
	f, err := os.Open(inFile)
	if err == nil {
		r := bufio.NewReader(f)
		for s, e := Readln(r); e == nil; s, e = Readln(r) {
			// Read a line from the file
			if s == "" || len(s) < 2 {
				goto goreturn
			}
			parts := strings.SplitN(s, "=", 2)
			if len(parts) == 2 {
				pd[strings.Trim(parts[0], " ")] = string(strings.Trim(parts[1], " "))
			}
		}
	goreturn:
		return pd
	}
	return nil
}

func loadUrlList(inFile string) []string {
	var ul []string
	f, err := os.Open(inFile)
	if err == nil {
		r := bufio.NewReader(f)
		for s, e := Readln(r); e == nil; s, e = Readln(r) {
			// Read a line from the file
			if s == "" {
				continue
			}
			ul = append(ul, strings.Trim(s, " "))
		}
		return ul
	}
	return nil
}

func loadFileToArray(inFile string) []string {
	var list []string
	f, err := os.Open(inFile)
	if err == nil {
		r := bufio.NewReader(f)
		for s, e := Readln(r); e == nil; s, e = Readln(r) {
			// Read a line from the file
			if s == "" {
				continue
			}
			list = append(list, strings.Trim(s, " "))
		}
		return list
	}
	return nil
}

func loadHeadersFile(inFile string) []map[string]string {

	hdrs := loadFileToArray(inFile)
	headersMap := []map[string]string{}
	for _, h := range hdrs {
		requestHeaders := strings.Split(h, "~")
		for _, hp := range requestHeaders {
			headerParts := strings.Split(hp, ":")
			headersMap = append(headersMap, map[string]string{headerParts[0]: headerParts[1]})
		}
	}

	return headersMap
}

func loadJsonHeadersFile(inFile string) []map[string]string {
	return []map[string]string{}
}

/*
func loadJsonHeadersFileFromAPI(inUrl string) []map[string]string {

	payload := map[string]string{
		"of": "JSON", // output format
		"k":  drc.apiKey,
	}

	jsonContent, _ := json.Marshal(payload)
	url := drc.urlEncode(drc.baseUrl+drc.actionTestHeaders, payload)

	if drc.debugMode == 1 {
		fmt.Println("SetTestHeaders URL:\n", url)
		fmt.Println("SetTestHeaders JSON Payload:\n", string(jsonContent))
	}

	drc.headers = drc.getProfile(url)
}
*/

func makeRequest(urlToCall string, opt *BenchOptions, stats *BenchStats, wg *sync.WaitGroup, benchTests chan BenchTest) { //w *sync.WaitGroup) {

	fmt.Println("Requesting: ", urlToCall)

	defer wg.Done()
	client := http.Client{}

	values := make(url.Values)
	if len(opt.PostData) >= 1 {
		opt.Method = "POST"
		for k, v := range opt.PostData {
			values.Add(k, v)
		}
	}

	var req *http.Request

	if opt.Method == "POST" {
		req, _ = http.NewRequest(opt.Method, urlToCall, strings.NewReader(values.Encode()))
	} else {
		req, _ = http.NewRequest(opt.Method, urlToCall, nil)
	}

	// Add all request cookies if any have been specified
	numCookies := len(opt.Cookies)
	if numCookies > 0 {
		for i := 0; i < numCookies; i++ {
			req.AddCookie(&opt.Cookies[i])
		}
	}

	if opt.UserAgents != nil {
		rand.Seed(time.Now().UnixNano())
		uaListLen := len(options.UserAgents)
		req.Header.Add("User-Agent", options.UserAgents[rand.Intn(uaListLen)])
	} else {
		req.Header.Add("User-Agent", opt.UserAgent)
	}

	if len(opt.HeaderList) >= 1 {
		if len(opt.PostData) >= 1 {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		}
		//req.Header.Add("Connection", "Keep-alive")
		index := rand.Intn(len(opt.HeaderList))
		for k, v := range opt.HeaderList[index] {
			req.Header.Add(k, strings.Trim(v, " "))
		}
	}

	tStart := time.Now().UnixNano()
	resp, err := client.Do(req)

	// Increment the total test count
	atomic.AddInt32(&stats.TestCount, 1)

	if err != nil {
		fmt.Println("\tWarning: Failed to connect, ", err)
		//stats.TestTime = append(stats.TestTime, 0)
		//stats.TestCount++
		//stats.NumFail++
		//stats.StatusCode["CF"]++
		benchTests <- BenchTest{TimeTaken: 0, StatusCode: 000, Url: urlToCall, BytesDownloaded: int64(0)}
		atomic.AddInt32(&stats.TestCount, 1)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	//stats.BytesDownloaded += len(body)
	bytesDownloaded := len(body)
	body = nil

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		//stats.StatusCode["2xx"]++
		atomic.AddInt32(&stats.Resp2xx, 1)
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		//stats.StatusCode["3xx"]++
		atomic.AddInt32(&stats.Resp3xx, 1)
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		//stats.StatusCode["4xx"]++
		atomic.AddInt32(&stats.Resp4xx, 1)
	case resp.StatusCode >= 500:
		//stats.StatusCode["5xx"]++
		atomic.AddInt32(&stats.Resp5xx, 1)
	}

	if resp.StatusCode != 200 {
		fmt.Println("\tWarning: Failed request, resp code ", resp.StatusCode)
		//stats.TestTime = append(stats.TestTime, 0)
		//stats.TestCount++
		//stats.NumFail++
		atomic.AddInt32(&stats.NumFail, 1)
		benchTests <- BenchTest{TimeTaken: 0, StatusCode: resp.StatusCode, Url: urlToCall, BytesDownloaded: int64(bytesDownloaded)}
		atomic.AddInt32(&stats.TestCount, 1)
		return
	}

	defer resp.Body.Close()
	tEnd := time.Now().UnixNano()

	//stats.TestTime = append(stats.TestTime, FromNanoToMilli(tEnd-tStart)) // convert to milliseconds
	//stats.TestCount++
	//stats.NumPass++

	benchTests <- BenchTest{TimeTaken: FromNanoToMilli(tEnd - tStart), StatusCode: resp.StatusCode, Url: urlToCall, BytesDownloaded: int64(bytesDownloaded)}
	atomic.AddInt32(&stats.TestCount, 1)

	//if resp.Header.Get("Server") != "" {
	if stats.ServerType == "" {
		// Lock the variable
		stats.Lock.Lock()
		// then set it
		stats.ServerType = resp.Header.Get("Server")
		// and release the lock
		stats.Lock.Unlock()
	}

}

func showCommandUsage(desc map[string]string) {

	fmt.Println("Command usage:\n")
	for k, v := range desc {
		fmt.Println("\t-" + k + " : " + v)
	}
	fmt.Println("\n")
	os.Exit(0)
}

func main() {

	var url, postDataFile, urlList, uaList, cookieFile, userAgent, headersList string
	var concurency, totalTests, maxCores, rampTime, timeWait int
	//var once [4]sync.Once
	var wg sync.WaitGroup
	var benchTests chan BenchTest

	description := map[string]string{
		"u":  "The full url to test (required)",
		"c":  "Number of requests to run concurrently (default = 1)",
		"p":  "Number of processor cores to use (default = 1)",
		"m":  "Total number of tests to run (default = 25)",
		"rt": "The number of milliseconds until ramped up to specified concurency",
		"tw": "The number of milliseconds to wait betweeen concurent test runs",
		"pd": "Enable POST request and use specifid data file",
		"l":  "File containing the list of urls to test",
		"cf": "File containing the cookies to send for a given request",
		"ul": "File containing the list of user-agents to use for each request at random.",
		"ua": "User agent to use for stress test (overridden by ul)",
		"hl": "Headers list file to use",
	}

	if len(os.Args) == 2 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			fmt.Println("GoLog - Version " + getVersion() + "\n")
			os.Exit(0)
		} else if os.Args[1] == "-h" || os.Args[1] == "--help" {
			showCommandUsage(description)
		}
	} else if len(os.Args) == 1 {
		showCommandUsage(description)
	}

	flag.StringVar(&url, "u", "", description["u"])
	flag.IntVar(&concurency, "c", 1, description["c"])
	flag.IntVar(&maxCores, "p", 1, description["p"])
	flag.IntVar(&totalTests, "m", 25, description["m"])
	//flag.IntVar(&rampThreads, "rf", 1, "The number of thread to gradually ramp up by")
	flag.IntVar(&rampTime, "rt", 1, description["rt"])
	flag.IntVar(&timeWait, "tw", 1000, description["tw"])
	flag.StringVar(&postDataFile, "pd", "", description["pd"])
	flag.StringVar(&urlList, "l", "", description["l"])
	flag.StringVar(&cookieFile, "cf", "", description["cf"])
	flag.StringVar(&uaList, "ul", "", description["ul"])
	flag.StringVar(&userAgent, "ua", "", description["ua"])
	flag.StringVar(&headersList, "hl", "", description["hl"])

	flag.Parse()

	if url == "" {
		showCommandUsage(description)
	}

	if maxCores > runtime.NumCPU() {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(maxCores)
	}

	if postDataFile == "" {
		options.Method = "GET"
		options.PostData = nil
		if urlList != "" {
			options.UrlList = loadFileToArray(urlList)
			options.UrlListLen = len(options.UrlList)
		}
	} else {
		options.Method = "POST"
		options.PostData = loadPostData(postDataFile)
		fmt.Println("Post data:", options.PostData)
		if options.PostData == nil {
			options.Method = "GET"
			fmt.Println("Warning: Post data file " + postDataFile + " does not exist or has no data!")
		}
	}

	if cookieFile != "" {
		options.Cookies = loadCookieData(cookieFile)
	}

	if options.UrlList != nil {
		fmt.Println("\nRunning tests on", options.UrlListLen, "different URLS randomly")
	} else {
		fmt.Println("\nRunning tests on " + url)
	}

	if uaList != "" {
		options.UserAgents = loadFileToArray(uaList)
	}

	if headersList != "" {
		options.HeaderList = loadHeadersFile("configs/playback_headers.txt")
	}

	options.TotalTests = totalTests
	options.Concurency = concurency

	statusCodes := map[string]int32{"2xx": 0, "3xx": 0, "4xx": 0, "5xx": 0}
	stats := BenchStats{TestStart: time.Now().UnixNano(), BytesDownloaded: 0, ServerType: "", StatusCode: statusCodes, UrlListTestCount: map[string]int{}, Lock: &sync.Mutex{}}
	wg.Add(totalTests)

	//i := totalTests
	//done := make(chan bool) // totalTests
	benchTests = make(chan BenchTest, totalTests)

	if rampTime > 0 && concurency >= 1 {

	}

	for j := 0; j < concurency; j++ {
		// GO thread to run the individual tests
		go func() {
			// only do test if there are
			if stats.TestCount < int32(options.TotalTests) {
				if options.UrlList != nil {
					rand.Seed(time.Now().UnixNano())
					url = options.UrlList[rand.Intn(options.UrlListLen)]
				}
				makeRequest(url, &options, &stats, &wg, benchTests)
				//stats.UrlListTestCount[url]++
				//done <- true
			}

		}()

	}
	wg.Done()
	fmt.Println("Issued all threads!")
	wg.Wait()

	fmt.Println("All threads done!")

	close(benchTests)
	stats.TestEnd = time.Now().UnixNano()
	stats.TestTime = SortResponseTimes(stats.TestTime)

	fmt.Println("\n-------------- Test Statistics ---------------")
	fmt.Println("Num CPU cores used:", maxCores)

	if options.UrlList != nil {
		fmt.Println("Total URL variations:", options.UrlListLen)
	} else {
		fmt.Println("URL Requested: " + url)
	}

	// Now process the stats from the BenchTest objects
	/*
		TimeTaken       int64
		StatusCode      int
		Url             string
		BytesDownloaded int64
	*/

	/*
		for i := range benchTests {
			fmt.Println(i)
		}
	*/

	fmt.Println("Server type: ", stats.ServerType)
	fmt.Println("Total tests run: ", stats.TestCount)
	if stats.BytesDownloaded > 0 {
		fmt.Println("Total bytes downloaded: ", stats.BytesDownloaded, "("+strconv.Itoa((stats.BytesDownloaded/1024))+" KB)")
	}

	fmt.Println("Total pass: ", stats.NumPass)
	fmt.Println("Total fail: ", stats.NumFail)
	fmt.Println("\tTotal failed connections:", stats.RespCF)
	fmt.Println("\tTotal responses in 2xx:", stats.Resp2xx)
	fmt.Println("\tTotal responses in 3xx:", stats.Resp3xx)
	fmt.Println("\tTotal responses in 4xx:", stats.Resp4xx)
	fmt.Println("\tTotal responses in 5xx:", stats.Resp5xx)

	/*
		fmt.Println("Shortest time: ", stats.TestTime[0], "ms")
		fmt.Println("Longest time: ", stats.TestTime[len(stats.TestTime)-1], "ms")

		if totalTests%2 == 1 {
			var index int = int(math.Ceil(float64(len(stats.TestTime) / 2)))
			stats.MedianTime = int(stats.TestTime[index])
		} else {
			var index int = totalTests / 2
			stats.MedianTime = (stats.TestTime[index] + stats.TestTime[index+1]) / 2
		}

		for i := 0; i < len(stats.TestTime); i++ {
			stats.AvgTime += int32(stats.TestTime[i])
		}
		stats.AvgTime = int32(stats.AvgTime) / stats.TestCount

		fmt.Println("Median time: ", stats.MedianTime, "ms")
		fmt.Println("Avg. time:", stats.AvgTime, "ms")

		fmt.Println("")
		files := dumpToReportFile(&stats, &options, "stress_test")
		fmt.Println("For more details, please view the following saved reports:")
		for i := 0; i < len(files); i++ {
			fmt.Println("\t" + files[i])
		}
		fmt.Println("")
	*/
	//<-done

}
