package genetics

import (
	"github.com/blackswarm/siege/internal/protocol"
)

// FitnessFunc evaluates a scan strategy and returns a fitness score.
// Higher scores indicate better strategies.
type FitnessFunc func(strategy *ScanStrategy, results []protocol.Finding) float64

// DefaultFitness combines multiple fitness components into a single score.
func DefaultFitness(strategy *ScanStrategy, results []protocol.Finding) float64 {
	score := 0.0

	// Component 1: Coverage — number of findings produced
	coverageScore := float64(len(results)) / 100.0
	if coverageScore > 1.0 {
		coverageScore = 1.0
	}
	score += coverageScore * 0.3

	// Component 2: Severity — weighted by finding severity
	severityScore := severityWeightedScore(results) / 100.0
	if severityScore > 1.0 {
		severityScore = 1.0
	}
	score += severityScore * 0.3

	// Component 3: Efficiency — findings per second
	if strategy.Timeout > 0 && strategy.Concurrency > 0 {
		efficiency := coverageScore / (strategy.Timeout * float64(strategy.Concurrency) * 0.1)
		if efficiency > 1.0 {
			efficiency = 1.0
		}
		score += efficiency * 0.2
	}

	// Component 4: Diversity — reward for enabling more modules
	diversityScore := moduleDiversityScore(strategy)
	score += diversityScore * 0.1

	// Component 5: Stealth penalty — penalise high rate limits
	stealthScore := 1.0 - (float64(strategy.RateLimit) / 2000.0)
	if stealthScore < 0 {
		stealthScore = 0
	}
	score += stealthScore * 0.1

	return score
}

// severityWeightedScore returns a weighted score based on finding severities.
func severityWeightedScore(findings []protocol.Finding) float64 {
	score := 0.0
	for _, f := range findings {
		switch f.Severity {
		case "critical":
			score += 100
		case "high":
			score += 50
		case "medium":
			score += 20
		case "low":
			score += 5
		case "info":
			score += 1
		}
	}
	return score
}

// moduleDiversityScore rewards strategies that enable more scanning modules.
func moduleDiversityScore(strategy *ScanStrategy) float64 {
	count := 0
	if strategy.EnableTCP {
		count++
	}
	if strategy.EnableUDP {
		count++
	}
	if strategy.EnableService {
		count++
	}
	if strategy.EnableWeb {
		count++
	}
	if strategy.EnableVuln {
		count++
	}
	if strategy.EnableDNS {
		count++
	}
	if strategy.EnableSubdomain {
		count++
	}
	return float64(count) / 7.0
}
