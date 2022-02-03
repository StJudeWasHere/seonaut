package main

import "strconv"

type ChartItem struct {
	Key     string
	Value   int
	Percent int
	Data    CountList
}

type Chart []ChartItem

func NewChart(c CountList) Chart {
	chart := Chart{}
	total := 0
	var ce CountList

	for _, i := range c {
		total = total + i.Value
	}

	if len(c) > 4 {
		ce = c[4:]
		c = c[:4]
	}

	for _, i := range c {
		ci := ChartItem{
			Key:     i.Key,
			Value:   i.Value,
			Percent: int(float64(i.Value) / float64(total) * 100),
		}
		chart = append(chart, ci)
	}

	chart[len(chart)-1].Data = ce

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

	return s[0 : len(s)-1]
}
