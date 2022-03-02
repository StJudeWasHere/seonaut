package helper

import (
	"strconv"

	"github.com/stjudewashere/seonaut/internal/issue"
)

const (
	chartLimit = 4
)

type ChartItem struct {
	Key     string
	Value   int
	Percent int
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
			Key:     i.Key,
			Value:   i.Value,
			Percent: int(float64(i.Value) / float64(total) * 100),
		}

		if ci.Percent == 0 {
			ci.Percent = 1
		}

		if ci.Percent > 97 {
			ci.Percent = 100 - (len(c) - 1)
		}

		chart = append(chart, ci)
	}

	if len(chart) > chartLimit {
		chart[chartLimit-1].Key = "Other"
		for _, v := range chart[chartLimit:] {
			chart[chartLimit-1].Value += v.Value
			chart[chartLimit-1].Percent += v.Percent
		}

		chart = chart[:chartLimit]
	}

	return chart
}

func (c Chart) GetChart(c1, c2, c3, c4 string) string {
	colors := []string{c1, c2, c3, c4}
	var s string
	last := 0
	for i, v := range c {
		s = s + colors[i] + " " + strconv.Itoa(last) + "% " + strconv.Itoa(last+v.Percent) + "%,"
		last = last + v.Percent
	}

	if len(s) > 0 {
		return s[0 : len(s)-1]
	}

	return s
}
