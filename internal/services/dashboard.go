package services

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

const (
	chartLimit = 4
)

type (
	DashboardServiceStorage interface {
		CountByMediaType(int64) *models.CountList
		CountByStatusCode(int64) *models.CountList

		CountByCanonical(int64) int
		CountImagesAlt(int64) *models.AltCount
		CountScheme(int64) *models.SchemeCount
		CountByNonCanonical(int64) int
		GetStatusCodeByDepth(crawlId int64) []models.StatusCodeByDepth
	}

	DashboardService struct {
		store DashboardServiceStorage
	}
)

func NewDashboardService(store DashboardServiceStorage) *DashboardService {
	return &DashboardService{store: store}
}

// Returns a Chart with the PageReport's media type chart data.
func (s *DashboardService) GetMediaCount(crawlId int64) *models.Chart {
	v := s.store.CountByMediaType(crawlId)
	return newChart(v)
}

// Returns a Chart with the PageReport's status code chart data.
func (s *DashboardService) GetStatusCount(crawlId int64) *models.Chart {
	v := s.store.CountByStatusCode(crawlId)
	return newChart(v)
}

// Returns the count Images with and without the alt attribute.
func (s *DashboardService) GetImageAltCount(crawlId int64) *models.AltCount {
	return s.store.CountImagesAlt(crawlId)
}

// Returns the count of PageReports with and without https.
func (s *DashboardService) GetSchemeCount(crawlId int64) *models.SchemeCount {
	return s.store.CountScheme(crawlId)
}

// Returns a count of PageReports that are canonical or not.
func (s *DashboardService) GetCanonicalCount(crawlId int64) *models.CanonicalCount {
	return &models.CanonicalCount{
		Canonical:    s.store.CountByCanonical(crawlId),
		NonCanonical: s.store.CountByNonCanonical(crawlId),
	}
}

func (s *DashboardService) GetStatusCodeByDepth(crawlId int64) []models.StatusCodeByDepth {
	return s.store.GetStatusCodeByDepth(crawlId)
}

// Returns a Chart containing the keys and values from the CountList.
// It limits the slice to the chartLimit value.
func newChart(c *models.CountList) *models.Chart {
	chart := models.Chart{}
	total := 0

	for _, i := range *c {
		total = total + i.Value
	}

	for _, i := range *c {
		ci := models.ChartItem{
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

	return &chart
}
