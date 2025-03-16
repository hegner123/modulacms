package cli

import "github.com/charmbracelet/huh"

func (m model) DBForm() model {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("class").
				Options(huh.NewOptions("Warrior", "Mage", "Rogue")...).
				Title("Choose your class"),

			huh.NewSelect[int]().
				Key("level").
				Options(huh.NewOptions(1, 20, 9999)...).
				Title("Choose your level"),
		),
	)
	m.form = form
	return m
}
