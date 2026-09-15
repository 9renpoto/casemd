package domain

// ValidationRule identifies a v1 Markdown validation rule.
type ValidationRule string

const (
	RuleEmptyHeading       ValidationRule = "empty-heading"
	RuleDocumentTitle      ValidationRule = "document-title"
	RuleHeadingHierarchy   ValidationRule = "heading-hierarchy"
	RuleUnsupportedHeading ValidationRule = "unsupported-heading"
	RuleMissingSteps       ValidationRule = "missing-steps"
	RuleMissingCheckpoints ValidationRule = "missing-checkpoints"
)

// Diagnostic describes one v1 Markdown validation violation.
type Diagnostic struct {
	Source     string
	Line       int
	Rule       ValidationRule
	Message    string
	Suggestion string
}
