package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type counter map[string]int

var n = 8

func main() {
	result := make(counter)
	start := time.Now()
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1500; i++ {
				resp, err := http.Get("http://localhost:3030")
				if err != nil {
					log.Fatal(err)
				}
				mu.Lock()
				_, ok := result[resp.Status]
				if !ok {
					result[resp.Status] = 1
				} else {
					result[resp.Status]++
				}
				mu.Unlock()

			}
		}()
	}
	wg.Wait()

	log.Println(result)
	log.Printf("Прошло времени: %v\n", time.Since(start).Seconds())
}
