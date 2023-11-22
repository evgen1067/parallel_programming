package main

import (
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

var (
	left       = -4.5
	right      = 4.5
	iterations = 1_000_000
	n          = 3
	wg         sync.WaitGroup
	minValue   = 4.5
)

func eval() float64 {
	x := rand.Float64()*(right-left) + left
	y := rand.Float64()*(right-left) + left

	return math.Pow(1.0-x, 2.0) + 100.0*math.Pow(y-math.Pow(x, 2.0), 2.0)
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	for n != 9 {
		log.Printf("Число горутин = %v\n", n-2)
		wg.Add(n)
		done := make(chan struct{}, n-2)

		start := time.Now()
		f := make(chan float64)
		for i := 0; i < n-2; i++ {
			go func() {
				defer func() {
					wg.Done()
					done <- struct{}{}
				}()
				for i := 0; i < iterations/(n-2); i++ {
					if v := eval(); v < minValue {
						f <- v
					}

				}
			}()
		}

		go func() {
			defer func() {
				wg.Done()
				close(done)
				close(f)
			}()
			for i := 0; i < n-2; i++ {
				<-done
			}
		}()

		go func() {
			defer wg.Done()
			for v := range f {
				if minValue > v {
					minValue = v
				}
			}
		}()

		wg.Wait()
		log.Printf("Время выполнения: %v, Значение минимизируемой ф-ции: %.6f\n", time.Since(start).Seconds(), minValue)
		n++
	}
}
