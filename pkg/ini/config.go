package ini

import "strings"

type Config struct {
	Sections []*Section
}

func (c *Config) Values(section string, key string) ([]string, bool) {
	sec, found := c.Section(section)
	if !found {
		return nil, false
	}

	values, found := sec.Values(key)
	if !found {
		return nil, false
	}

	return values, true
}

func (c *Config) Section(name string) (*Section, bool) {
	for _, s := range c.Sections {
		if s.Name == name {
			return s, true
		}
	}

	return nil, false
}

type Section struct {
	Name  string
	Items []*Item
}

func (s *Section) Values(key string) ([]string, bool) {
	for _, i := range s.Items {
		if strings.EqualFold(i.Key, key) {
			return i.Values, true
		}
	}

	return nil, false
}

type Item struct {
	Key    string
	Values []string
}
