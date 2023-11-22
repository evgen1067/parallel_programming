package main

import (
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

var (
	min         = -4.5
	max         = 4.5
	generations = 100
	size        = 1000
	n           = 3
	//	wg         sync.WaitGroup
	//	minValue   = 4.5
)

type point struct {
	x float64
	y float64
	f float64
}

func createRandomIndividual() *point {
	x := rand.Float64()*(max-min) + min
	y := rand.Float64()*(max-min) + min
	return &point{
		x: x,
		y: y,
		f: f(x, y),
	}
}

func f(x, y float64) float64 {
	return math.Pow(1.0-x, 2.0) + 100.0*math.Pow(y-math.Pow(x, 2.0), 2.0)
}

// Функция для получения лучшего индивида в популяции
func getBestIndividual(population []*point) float64 {
	best := 1_000_000.00 // Предполагаем, что первый индивид - лучший

	for _, individual := range population {
		// Сравниваем fitness текущего индивида с fitness лучшего индивида
		if individual.f < best {
			best = individual.f
		}
	}

	return best
}

// Шаг 4: Мутация потомков
func mutate(p *point) {

}

// Шаг 5: Замена старой популяции новой
func replacePopulation(oldPopulation []*point, newPopulation []*point) {
	// Заменяем старую популяцию новой
	copy(oldPopulation, newPopulation)
}

// Шаг 6: Повторение шагов 2-6 до достижения критерия останова (например, максимального числа поколений)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	for n != 9 {
		bestIndividual := 0.0
		log.Printf("Число горутин = %v\n", n-2)
		start := time.Now()
		for generation := 0; generation < generations; generation++ {
			population := make([]*point, 0)
			totalFitness := 0.0

			cumulativeFitness := 0.0

			parentWg := &sync.WaitGroup{}
			daughterWg := &sync.WaitGroup{}
			points := make(chan *point)
			for i := 0; i < size/(n-2); i++ {
				daughterWg.Add(1)
				go func() {
					defer daughterWg.Done()
					points <- createRandomIndividual()
				}()
			}
			parentWg.Add(2)
			go func() {
				defer parentWg.Done()
				for v := range points {
					population = append(population, v)
					totalFitness += v.f
				}
			}()
			go func() {
				defer parentWg.Done()
				daughterWg.Wait()
				close(points)
			}()
			parentWg.Wait()

			normalizedFitness := make([]float64, len(population))
			rouletteWheel := make([]float64, len(population))
			selectedPopulation := make([]*point, 0)
			for idx, v := range population {
				cumulativeFitness += v.f / totalFitness
				rouletteWheel[idx] = cumulativeFitness
				normalizedFitness[idx] = v.f / totalFitness
			}

			// Выбираем индивидов
			for _, v := range population {
				spin := rand.Float64()
				for _, w := range normalizedFitness {
					if spin <= w && len(selectedPopulation) < len(population) {
						selectedPopulation = append(selectedPopulation, v)
					}
				}
			}

			crossoverRate := 0.7 // Вероятность кроссинговера

			offspring := make([]*point, 0)

			for i := 0; i < len(selectedPopulation); i += 2 {
				// Проверяем, есть ли достаточное количество индивидов для кроссинговера
				if i+1 < len(selectedPopulation) {
					parent1 := selectedPopulation[i]
					parent2 := selectedPopulation[i+1]

					// Проверяем, произойдет ли кроссинговер
					if rand.Float64() < crossoverRate {
						// Одноточечный кроссинговер
						child1 := &point{x: parent1.x, y: parent2.y, f: f(parent1.x, parent2.y)}
						child2 := &point{x: parent2.x, y: parent1.y, f: f(parent2.x, parent1.y)}

						offspring = append(offspring, child1, child2)
					} else {
						// Если кроссинговер не произошел, просто добавляем родителей в потомство
						offspring = append(offspring, parent1, parent2)
					}
				} else {
					// Если не хватает индивидов для кроссинговера, просто добавляем последнего индивида
					offspring = append(offspring, selectedPopulation[i])
				}
			}

			offspringChan := make(chan *point, len(offspring))
			go func() {
				defer func() {
					close(offspringChan)
				}()
				for _, v := range offspring {
					select {
					case offspringChan <- v:
					default:
						return
					}
				}
			}()
			for i := 1; i <= n-1; i++ {
				daughterWg.Add(1)
				go func() {
					defer daughterWg.Done()
					for v := range offspringChan {
						mutationRate := 0.05 // Вероятность мутации

						// Для каждой координаты (x и y) индивида
						if rand.Float64() < mutationRate {
							v.x = rand.Float64()*(max-min) + min
						}
						if rand.Float64() < mutationRate {
							v.y = rand.Float64()*(max-min) + min
						}
						v.f = f(v.x, v.y)
					}
				}()
			}
			daughterWg.Wait()

			replacePopulation(population, offspring)

			bestIndividual = getBestIndividual(population)
		}
		n++
		log.Printf("Время выполнения: %v, Значение минимизируемой ф-ции: %.6f\n", time.Since(start).Seconds(), bestIndividual)
	}

}
