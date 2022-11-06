package cli

func (c IndexCommand) Help() string {
	return ""
}

func (c IndexCommand) Synopsis() string {
	return "Build an index of image duplicates."
}

func (c IndexCommand) Run(args []string) int {
	return 0
}
