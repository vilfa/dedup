package cli

func (c ExportCommand) Help() string {
	return ""
}

func (c ExportCommand) Synopsis() string {
	return "Export the image duplicates in CSV or JSON."
}

func (c ExportCommand) Run(args []string) int {
	return 0
}
