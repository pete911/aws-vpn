package prompt

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"strings"
)

var bold = promptui.Styler(promptui.FGBold)

func Select(label string, items []string) (int, string) {
	if len(items) == 0 {
		return -1, ""
	}

	// no need for prompt if there is only one item to chose from
	if len(items) == 1 {
		// replicate prompt ui selected item
		fmt.Printf("%s %s\n", bold(promptui.IconGood), bold(items[0]))
		return 0, items[0]
	}

	p := promptui.Select{
		Label: label,
		Items: items,
		Size:  10,
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
