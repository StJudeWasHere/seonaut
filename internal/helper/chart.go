package helper

import (
	"github.com/stjudewashere/seonaut/internal/issue"
)

const (
	chartLimit = 5
)

type ChartItem struct {
	Key   string
	Value int
}

type Chart []ChartItem

func NewChart(c issue.CountList) Chart {
	chart := Chart{}
	total := 0

	for _, i := range c {
		total = total + i.Value
	}

	for _, i := range c {
		ci := ChartItem{
			Key:   i.Key,
			Value: i.Value,
		}

		chart = append(chart, ci)
	}

	if len(chart) > chartLimit {
		chart[chartLimit-1].Key = "Other"
		for _, v := range chart[chartLimit:] {
			chart[chartLimit-1].Value += v.Value
		}

		chart = chart[:chartLimit]
	}

	return chart
}
