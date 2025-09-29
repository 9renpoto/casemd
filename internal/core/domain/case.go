package domain

// Case represents a single test case from the Markdown input.
type Case struct {
	MajorItem       string
	MediumItem      string
	MinorItem       string
	ValidationSteps []string
	Checkpoints     []string
}
