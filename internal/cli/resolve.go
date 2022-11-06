package cli

func (c ResolveCommand) Help() string {
	return ""
}

func (c ResolveCommand) Synopsis() string {
	return "Resolve the image duplicates, by removing them."
}

func (c ResolveCommand) Run(args []string) int {
	return 0
}
