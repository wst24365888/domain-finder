package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func worker(jobs chan string, results chan string) {
	for job := range jobs {
		client := &http.Client{}

		req, err := http.NewRequest("GET", "https://www.domainnamesoup.com/cell3.php?domain="+job+"&pt=86", nil)
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
		results <- job + ": " + strconv.FormatBool(result)
	}
}

func main() {
	const WORKER_AMOUNT int = 10
	const TLD string = ".com"

	// Go doesn't support const arrays (or slices)
	ALPHABETS := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	JOB_AMOUNT := len(ALPHABETS) * 1 * len(ALPHABETS)

	jobs := make(chan string, JOB_AMOUNT)
	results := make(chan string, JOB_AMOUNT)

	for workerIndex := 0; workerIndex < WORKER_AMOUNT; workerIndex++ {
		go worker(jobs, results)
	}

	for _, firstLetter := range ALPHABETS {
		for _, secondLetter := range ALPHABETS {
			jobs <- firstLetter + "-" + secondLetter + TLD
		}
	}

	close(jobs)

	for r := 1; r <= JOB_AMOUNT; r++ {
		fmt.Println(<-results)
	}

	close(results)
}
