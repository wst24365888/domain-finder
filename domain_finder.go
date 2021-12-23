package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

type DomainStatus struct {
	domain    string
	available bool
	message   string
}

func worker(jobs chan string, results chan DomainStatus) {
	for job := range jobs {
		client := &http.Client{}

		req, err := http.NewRequest("GET", "https://www.domainnamesoup.com/cell3.php?domain="+job+"&pt=86", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36")
		req.Header.Set("Referer", "https://www.google.com")
		if err != nil {
			log.Println(err)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()

		page, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}

		result := (string(page) == "-1")
		d := DomainStatus{domain: job, available: result, message: job + ": " + strconv.FormatBool(result)}
		results <- d
	}
}

func main() {
	const WORKER_AMOUNT int = 10
	const TLD string = ".com"

	// Go doesn't support const arrays (or slices)
	ALPHABETS := []string{"-", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	JOB_AMOUNT := len(ALPHABETS) * len(ALPHABETS) * len(ALPHABETS) * len(ALPHABETS)
	var ACTUAL_JOB_AMOUNT int = 0

	now := time.Now()
	f, err := os.Create("./results/domains-available-" + now.Format("2006-02-01T15-04-05") + ".txt")
	if err != nil {
		log.Fatal(err)
	}

	jobs := make(chan string, JOB_AMOUNT)
	results := make(chan DomainStatus, JOB_AMOUNT)

	for workerIndex := 0; workerIndex < WORKER_AMOUNT; workerIndex++ {
		go worker(jobs, results)
	}

	for _, firstLetter := range ALPHABETS {
		for _, secondLetter := range ALPHABETS {
			for _, thirdLetter := range ALPHABETS {
				for _, fourthLetter := range ALPHABETS {
					domain := firstLetter + secondLetter + thirdLetter + fourthLetter + TLD
					match, _ := regexp.MatchString("(([a-z0-9][a-z0-9][a-z0-9\\-][a-z0-9])|([a-z0-9][a-z0-9\\-][a-z0-9][a-z0-9]))\\.com", domain)

					if match {
						jobs <- domain
						ACTUAL_JOB_AMOUNT++
					}
				}
			}
		}
	}

	close(jobs)

	for r := 1; r <= ACTUAL_JOB_AMOUNT; r++ {
		d := <-results

		if d.available {
			f.WriteString(d.domain + "\n")
		}
	}

	close(results)
}
