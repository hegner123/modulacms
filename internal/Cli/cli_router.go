package cli

func TableControlsRouter(page CliPage) CliInterface {
	switch page.Index {
	case 11:
		return createInterface
	default:
		return pageInterface
	}

}
