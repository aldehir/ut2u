package ini

type Config struct {
	Sections []*Section
}

type Section struct {
	Name  string
	Items []*Item
}

type Item struct {
	Key    string
	Values []string
}
