package genetics

import (
	"log/slog"
)

// EvolutionEngine drives the genetic optimisation of scan strategies
// using real-world feedback from executed scans.
type EvolutionEngine struct {
	ga *GeneticAlgorithm
}

// NewEvolutionEngine creates a new EvolutionEngine with the given configuration.
func NewEvolutionEngine(cfg GAConfig, fitnessFunc FitnessFunc, logger *slog.Logger) *EvolutionEngine {
	return &EvolutionEngine{
		ga: NewGA(cfg, fitnessFunc, logger),
	}
}

// Run executes the evolution loop for the configured number of generations.
func (e *EvolutionEngine) Run(evaluator func(*ScanStrategy) float64) {
	e.ga.Evolve(evaluator)
}

// Best returns the current best strategy.
func (e *EvolutionEngine) Best() *ScanStrategy {
	return e.ga.BestStrategy()
}

// Population returns the current population snapshot.
func (e *EvolutionEngine) Population() *Population {
	return e.ga.Population()
}
