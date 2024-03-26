package models

type CountItem struct {
	Key   string
	Value int
}

type CountList []CountItem

func (c CountList) Len() int {
	return len(c)
}

func (c CountList) Less(i, j int) bool {
	return c[i].Value < c[j].Value
}

func (c CountList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
