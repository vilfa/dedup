package cli

func (c DryRunCommand) Help() string {
	return ""
}

func (c DryRunCommand) Synopsis() string {
	return "Perform a dryrun resolve operation, log everything."
}

func (c DryRunCommand) Run(args []string) int {
	return 0
}
