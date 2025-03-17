package cli

type CLIRoute interface {
	NextPage(PageIndex) *CliPage
	NextController(PageIndex) *CliPage
}

func (c CliPage) NextPage(index PageIndex) *CliPage {
	switch index {
	case Read:
		return tablePage
	default:
		return homePage
	}

}

func (c CliPage) NextController(index PageIndex) *CliInterface {
	switch index {
	case Read:
		return &tableInterface
	default:
		return &pageInterface
	}

}

func NextFromTable(index int) (*CliInterface, *CliPage) {
    return nil, nil
}
