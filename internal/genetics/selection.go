package genetics

import (
	"math/rand"
	"sort"
)

// Selector implements parent selection methods.
type Selector struct {
	tournamentSize int
	eliteCount     int
}

// NewSelector creates a new Selector with the given parameters.
func NewSelector(tournamentSize, eliteCount int) *Selector {
	if tournamentSize < 2 {
		tournamentSize = 2
	}
	return &Selector{
		tournamentSize: tournamentSize,
		eliteCount:     eliteCount,
	}
}

// SelectParents selects two parents from the population using tournament selection.
func (s *Selector) SelectParents(pop *Population) (*Individual, *Individual) {
	parent1 := s.tournamentSelect(pop)
	parent2 := s.tournamentSelect(pop)
	return parent1, parent2
}

// tournamentSelect picks an individual via tournament selection.
func (s *Selector) tournamentSelect(pop *Population) *Individual {
	individuals := pop.All()
	best := individuals[rand.Intn(len(individuals))]
	for i := 1; i < s.tournamentSize; i++ {
		contender := individuals[rand.Intn(len(individuals))]
		if contender.Fitness > best.Fitness {
			best = contender
		}
	}
	return best
}

// SelectElite returns the top eliteCount individuals.
func (s *Selector) SelectElite(pop *Population) []*Individual {
	sorted := pop.Sorted()
	if s.eliteCount > len(sorted) {
		s.eliteCount = len(sorted)
	}
	return sorted[:s.eliteCount]
}

// SelectSurvivors performs elitism: keep the best individuals, replace the rest.
// Returns a new population of the same size.
func (s *Selector) SelectSurvivors(pop *Population, offspring []*Individual) *Population {
	all := append(pop.All(), offspring...)
	sort.Slice(all, func(i, j int) bool {
		return all[i].Fitness > all[j].Fitness
	})

	newPop := &Population{size: pop.Size()}
	newPop.individuals = make([]*Individual, pop.Size())
	for i := 0; i < pop.Size(); i++ {
		newPop.individuals[i] = all[i]
	}
	return newPop
}
