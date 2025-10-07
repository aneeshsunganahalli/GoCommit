package types

// GenerationOptions controls how commit messages should be produced by LLM providers.
type GenerationOptions struct {
	// StyleInstruction contains optional tone/style guidance appended to the base prompt.
	StyleInstruction string
	// Attempt records the 1-indexed attempt number for this generation request.
	// Attempt > 1 signals that the LLM should provide an alternative output.
	Attempt int
}
