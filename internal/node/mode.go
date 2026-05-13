package node

// Mode represents the operational role of a node in the swarm.
type Mode string

const (
	// ModeStrategic indicates a node that coordinates the swarm, runs the genetic algorithm,
	// schedules tasks, and serves the operator dashboard.
	ModeStrategic Mode = "strategic"

	// ModeScanner indicates a node dedicated to executing scanning tasks.
	ModeScanner Mode = "scanner"

	// ModeHybrid indicates a node that performs both coordination and scanning.
	ModeHybrid Mode = "hybrid"
)

// ValidModes is the set of allowed node modes.
var ValidModes = map[Mode]bool{
	ModeStrategic: true,
	ModeScanner:    true,
	ModeHybrid:     true,
}

// IsCoordinator returns true if the mode includes strategic capabilities.
func (m Mode) IsCoordinator() bool {
	return m == ModeStrategic || m == ModeHybrid
}

// IsScanner returns true if the mode includes scanning capabilities.
func (m Mode) IsScanner() bool {
	return m == ModeScanner || m == ModeHybrid
}

// String returns the string representation of the mode.
func (m Mode) String() string {
	return string(m)
}
