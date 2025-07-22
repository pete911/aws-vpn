package prompt

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"strings"
)

var bold = promptui.Styler(promptui.FGBold)

func Prompt(label string) bool {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}

	if _, err := prompt.Run(); err != nil {
		// error is returned if the result is not "y"
		fmt.Println("operation canceled")
		return false
	}
	return true
}

func Select(label string, items []string, defaultItem string) (int, string) {
	if len(items) == 0 {
		return -1, ""
	}

	// no need for prompt if there is only one item to chose from
	if len(items) == 1 {
		// replicate prompt ui selected item
		fmt.Printf("%s %s\n", bold(promptui.IconGood), bold(items[0]))
		return 0, items[0]
	}

	var cursorPos int
	for k, v := range items {
		if v == defaultItem {
			cursorPos = k
			break
		}
	}

	p := promptui.Select{
		Label:     label,
		Items:     items,
		CursorPos: cursorPos,
		Size:      10,
		Searcher: func(input string, index int) bool {
			item := items[index]
			name := strings.Replace(strings.ToLower(item), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)
			return strings.Contains(name, input)
		},
	}
	i, result, err := p.Run()
	if err != nil {
		fmt.Printf("%s: %v\n", label, err)
		os.Exit(1)
	}
	return i, result
}
