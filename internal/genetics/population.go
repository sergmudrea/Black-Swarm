package genetics

import (
	"math/rand"
	"sort"
	"sync"
)

// Population manages a collection of chromosomes (scan strategies).
type Population struct {
	mu          sync.RWMutex
	individuals []*Individual
	size        int
}

// Individual pairs a strategy with its fitness score.
type Individual struct {
	Strategy *ScanStrategy `json:"strategy"`
	Fitness  float64       `json:"fitness"`
	ID       int           `json:"id"`
}

// NewPopulation creates a population of the given size with random strategies.
func NewPopulation(size int) *Population {
	pop := &Population{
		individuals: make([]*Individual, size),
		size:        size,
	}
	for i := 0; i < size; i++ {
		pop.individuals[i] = &Individual{
			Strategy: RandomStrategy(),
			Fitness:  0.0,
			ID:       i,
		}
	}
	return pop
}

// Size returns the population size.
func (p *Population) Size() int {
	return p.size
}

// Get returns the individual at the given index.
func (p *Population) Get(index int) *Individual {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if index < 0 || index >= len(p.individuals) {
		return nil
	}
	return p.individuals[index]
}

// Set updates an individual at the given index.
func (p *Population) Set(index int, ind *Individual) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if index >= 0 && index < len(p.individuals) {
		p.individuals[index] = ind
	}
}

// Best returns the individual with the highest fitness.
func (p *Population) Best() *Individual {
	p.mu.RLock()
	defer p.mu.RUnlock()

	best := p.individuals[0]
	for _, ind := range p.individuals[1:] {
		if ind.Fitness > best.Fitness {
			best = ind
		}
	}
	return best
}

// Worst returns the individual with the lowest fitness.
func (p *Population) Worst() *Individual {
	p.mu.RLock()
	defer p.mu.RUnlock()

	worst := p.individuals[0]
	for _, ind := range p.individuals[1:] {
		if ind.Fitness < worst.Fitness {
			worst = ind
		}
	}
	return worst
}

// AverageFitness returns the mean fitness of the population.
func (p *Population) AverageFitness() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sum := 0.0
	for _, ind := range p.individuals {
		sum += ind.Fitness
	}
	return sum / float64(len(p.individuals))
}

// Sorted returns individuals sorted by fitness descending.
func (p *Population) Sorted() []*Individual {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sorted := make([]*Individual, len(p.individuals))
	copy(sorted, p.individuals)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness > sorted[j].Fitness
	})
	return sorted
}

// Replace replaces the entire population with new individuals.
func (p *Population) Replace(individuals []*Individual) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.individuals = individuals
	p.size = len(individuals)
}

// RandomIndividual returns a random individual from the population.
func (p *Population) RandomIndividual() *Individual {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.individuals[rand.Intn(len(p.individuals))]
}

// All returns a copy of all individuals.
func (p *Population) All() []*Individual {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]*Individual, len(p.individuals))
	copy(result, p.individuals)
	return result
}
