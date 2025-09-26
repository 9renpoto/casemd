package app

// CLI defines the interface for command-line frontends that drive the application.
type CLI interface {
	Run(args []string) error
}

// App wires the external interface with the core application use cases.
type App struct {
	cli CLI
}

// New creates a new application instance with the provided CLI adapter.
func New(cli CLI) *App {
	return &App{cli: cli}
}

// Run delegates execution to the CLI adapter. Additional orchestrations will be
// added here as the application layers grow.
func (a *App) Run(args []string) error {
	return a.cli.Run(args)
}
