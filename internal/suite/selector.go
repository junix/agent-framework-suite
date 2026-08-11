package suite

import (
	"fmt"
	"strings"
)

func Select(catalog []Case, specs []string, tag string) ([]Case, error) {
	selected := make(map[string]bool)
	if len(specs) == 0 {
		for _, item := range catalog {
			selected[item.ID] = true
		}
	} else {
		for _, raw := range specs {
			for _, spec := range strings.Split(raw, ",") {
				spec = strings.TrimSpace(spec)
				if spec == "" {
					continue
				}
				matched := false
				for _, item := range catalog {
					if spec == "all" || strings.EqualFold(spec, item.ID) || strings.EqualFold(spec, item.Name) || strings.EqualFold(spec, strings.Split(item.ID, "-")[0]) {
						selected[item.ID] = true
						matched = true
					}
				}
				if !matched {
					return nil, fmt.Errorf("selector %q matched no cases", spec)
				}
			}
		}
	}

	var result []Case
	for _, item := range catalog {
		if !selected[item.ID] || (tag != "" && !contains(item.Tags, tag)) {
			continue
		}
		result = append(result, item)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("selection is empty")
	}
	return result, nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
