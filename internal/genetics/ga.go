package genetics

import (
	"log/slog"
	"math/rand"
	"sync"
)

// GeneticAlgorithm encapsulates the core evolutionary loop.
type GeneticAlgorithm struct {
	pop          *Population
	selector     *Selector
	crossover    *CrossoverOperator
	mutator      *Mutator
	fitnessFunc  FitnessFunc
	generations  int
	eliteCount   int
	logger       *slog.Logger
}

// GAConfig holds the parameters needed to initialise a GeneticAlgorithm.
type GAConfig struct {
	PopulationSize int
	Generations    int
	MutationRate   float64
	CrossoverRate  float64
	EliteCount     int
	TournamentSize int
}

// NewGA creates a new GeneticAlgorithm with the given configuration.
func NewGA(cfg GAConfig, fitnessFunc FitnessFunc, logger *slog.Logger) *GeneticAlgorithm {
	if fitnessFunc == nil {
		fitnessFunc = DefaultFitness
	}
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.PopulationSize <= 0 {
		cfg.PopulationSize = 50
	}
	if cfg.Generations <= 0 {
		cfg.Generations = 20
	}
	if cfg.MutationRate <= 0 {
		cfg.MutationRate = 0.05
	}
	if cfg.CrossoverRate <= 0 {
		cfg.CrossoverRate = 0.7
	}
	if cfg.EliteCount <= 0 {
		cfg.EliteCount = 2
	}
	if cfg.TournamentSize <= 1 {
		cfg.TournamentSize = 3
	}

	return &GeneticAlgorithm{
		pop:         NewPopulation(cfg.PopulationSize),
		selector:    NewSelector(cfg.TournamentSize, cfg.EliteCount),
		crossover:   NewCrossoverOperator(cfg.CrossoverRate),
		mutator:     NewMutator(cfg.MutationRate),
		fitnessFunc: fitnessFunc,
		generations: cfg.Generations,
		eliteCount:  cfg.EliteCount,
		logger:      logger,
	}
}

// Population returns the current population (read-only snapshot).
func (ga *GeneticAlgorithm) Population() *Population {
	return ga.pop
}

// BestStrategy returns the strategy with the highest fitness.
func (ga *GeneticAlgorithm) BestStrategy() *ScanStrategy {
	best := ga.pop.Best()
	if best == nil {
		return nil
	}
	return best.Strategy
}

// EvaluateAll scores every individual in the population using the supplied evaluator.
// The evaluator receives a strategy and returns a fitness value.
func (ga *GeneticAlgorithm) EvaluateAll(evaluator func(*ScanStrategy) float64) {
	var wg sync.WaitGroup
	individuals := ga.pop.All()
	for _, ind := range individuals {
		wg.Add(1)
		go func(ind *Individual) {
			defer wg.Done()
			ind.Fitness = evaluator(ind.Strategy)
		}(ind)
	}
	wg.Wait()
}

// Evolve runs the genetic algorithm for the configured number of generations.
// The evaluator is called once per strategy per generation to assign fitness.
func (ga *GeneticAlgorithm) Evolve(evaluator func(*ScanStrategy) float64) {
	for gen := 0; gen < ga.generations; gen++ {
		ga.logger.Info("starting generation",
			"generation", gen,
			"pop_size", ga.pop.Size(),
		)

		// Evaluate current population
		ga.EvaluateAll(evaluator)

		// Preserve elite
		elite := ga.selector.SelectElite(ga.pop)

		// Create offspring
		offspringSize := ga.pop.Size() - len(elite)
		offspring := make([]*Individual, offspringSize)
		for i := 0; i < offspringSize; i++ {
			parent1, parent2 := ga.selector.SelectParents(ga.pop)
			childStrategy := ga.crossover.Crossover(parent1.Strategy, parent2.Strategy)
			ga.mutator.Mutate(childStrategy)

			offspring[i] = &Individual{
				Strategy: childStrategy,
				Fitness:  0.0,
				ID:       ga.pop.Size() + i,
			}
		}

		// Combine elite and offspring, evaluate offspring
		newIndividuals := append(elite, offspring...)
		for _, ind := range offspring {
			ind.Fitness = evaluator(ind.Strategy)
		}

		// Form new population (elitism already preserved)
		ga.pop.Replace(newIndividuals)
	}

	// Final evaluation
	ga.EvaluateAll(evaluator)
	ga.logger.Info("evolution complete",
		"best_fitness", ga.pop.Best().Fitness,
		"avg_fitness", ga.pop.AverageFitness(),
	)
}
