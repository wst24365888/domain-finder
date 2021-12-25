package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

type Domain struct {
	name      string
	available bool
}

func dispatcher(jobChan chan Domain, closed chan bool) {
	const TLD string = ".com"

	// Go doesn't support const arrays (or slices)
	ALPHABETS := []string{"-", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

	defer close(jobChan)

	var counter int = 0

	for _, firstLetter := range ALPHABETS {
		for _, secondLetter := range ALPHABETS {
			for _, thirdLetter := range ALPHABETS {
				for _, fourthLetter := range ALPHABETS {
					domain := firstLetter + secondLetter + thirdLetter + fourthLetter + TLD
					match, _ := regexp.MatchString("(([a-z0-9][a-z0-9][a-z0-9\\-][a-z0-9])|([a-z0-9][a-z0-9\\-][a-z0-9][a-z0-9]))\\.com", domain)

					if match {
						jobChan <- Domain{domain, false}

						counter++

						if counter%100 == 0 {
							fmt.Print("\033[H\033[2J")
							fmt.Println("dispatched: " + strconv.Itoa(counter) + " jobs")
						}
					}
				}
			}
		}
	}

	closed <- true
}

func workerPool(max int, jobChan chan Domain, respChan chan Domain) {
	t := &http.Transport{}
	for i := 0; i < max; i++ {
		go worker(t, jobChan, respChan)
	}
}

func worker(t *http.Transport, jobChan chan Domain, respChan chan Domain) {
	for d := range jobChan {
		req, err := http.NewRequest("GET", "https://www.domainnamesoup.com/cell3.php?domain="+d.name+"&pt=86", nil)
		if err != nil {
			log.Println(err)
			continue
		}

		req.Header.Set("Referer", "https://www.google.com")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36")

		resp, err := t.RoundTrip(req)
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()

		page, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}

		d.available = (string(page) == "-1")
		respChan <- d
	}
}

func consumer(respChan chan Domain, closed chan bool) {
	now := time.Now()
	f, err := os.Create("./results/domains-available-" + now.Format("2006-02-01T15-04-05") + ".txt")
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-closed:
			break
		case d := <-respChan:
			if d.available {
				f.WriteString(d.name + "\n")
			}
		}
	}
}

func main() {
	const WORKER_AMOUNT int = 100
	const BUFFER_SIZE int = WORKER_AMOUNT

	runtime.GOMAXPROCS(runtime.NumCPU())
	jobChan := make(chan Domain, BUFFER_SIZE)
	respChan := make(chan Domain, BUFFER_SIZE)
	closed := make(chan bool)

	go workerPool(WORKER_AMOUNT, jobChan, respChan)
	go dispatcher(jobChan, closed)
	consumer(respChan, closed)
}
