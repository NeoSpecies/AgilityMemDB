package agilitymemdb

type Item struct {
	Key   string
	Value string
}

type Transaction struct {
	ID         string
	Operations map[string]*Item
}
