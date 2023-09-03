package ini

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

func Parse(r io.Reader) (*Config, error) {
	cfg := &Config{}

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		if err := cfg.handleLine(scanner.Text()); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (c *Config) handleLine(line string) error {
	line = stripComments(line)

	section, ok := sectionName(line)
	if ok {
		c.Sections = append(c.Sections, &Section{Name: section})
		return nil
	}

	key, value, ok := strings.Cut(line, "=")
	if ok {
		if len(c.Sections) == 0 {
			return errors.New("found key/value pair before section")
		}

		section := c.Sections[len(c.Sections)-1]

		var item *Item
		for _, it := range section.Items {
			if strings.EqualFold(it.Key, key) {
				item = it
				break
			}
		}

		if item == nil {
			item = &Item{key, nil}
			section.Items = append(section.Items, item)
		}

		item.Values = append(item.Values, value)

		return nil
	}

	return nil
}

func stripComments(line string) string {
	cutoff := len(line)

	for _, delim := range ";#" {
		if idx := strings.IndexRune(line, delim); idx != -1 {
			if idx < cutoff {
				cutoff = idx
			}
		}
	}

	return strings.TrimSpace(line[:cutoff])
}

func sectionName(line string) (string, bool) {
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		return line[1 : len(line)-1], true
	}
	return "", false
}
