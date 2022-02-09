package main

type Crawl struct {
	Id        int
	ProjectId int
	URL       string
	Start     time.Time
	End       sql.NullTime
}

func (c Crawl) TotalTime() time.Duration {
	if c.End.Valid {
		return c.End.Time.Sub(c.Start)
	}

	return 0
}
