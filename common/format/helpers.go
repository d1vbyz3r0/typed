package format

func splitByDive(rules []string) [][]string {
	out := make([][]string, 0)
	start := 0
	for i, r := range rules {
		if r == "dive" {
			out = append(out, rules[start:i])
			start = i + 1
		}
	}
	out = append(out, rules[start:])
	return out
}

func filterEmpty(rules []string) []string {
	out := make([]string, 0, len(rules))
	for _, r := range rules {
		if r != "" && r != "dive" {
			out = append(out, r)
		}
	}
	return out
}

func stripKeysBlock(rules []string) []string {
	out := make([]string, 0, len(rules))
	skipping := false

	for _, r := range rules {
		if r == "keys" {
			skipping = true
			continue
		}

		if r == "endkeys" {
			skipping = false
			continue
		}

		if skipping {
			continue
		}

		out = append(out, r)
	}

	return out
}
