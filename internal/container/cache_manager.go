package container

import (
	"github.com/stjudewashere/seonaut/internal/models"
)

type CrawlCacheHandler interface {
	BuildCrawlCache(crawl *models.Crawl)
	RemoveCrawlCache(crawl *models.Crawl)
}

type CacheManager struct {
	handlers []CrawlCacheHandler
}

func NewCacheManager() *CacheManager {
	return &CacheManager{}
}

// AddCrawlCacheHandler adds a new crawl cache handler to the handlers slice.
func (cm *CacheManager) AddCrawlCacheHandler(handler CrawlCacheHandler) {
	cm.handlers = append(cm.handlers, handler)
}

// BuildCrawlCache calls the BuildCrawlCache method on all the handlers.
func (cm *CacheManager) BuildCrawlCache(crawl *models.Crawl) {
	for _, c := range cm.handlers {
		c.BuildCrawlCache(crawl)
	}
}

// BuildCrawlCache calls the RemoveCrawlCache method on all the handlers.
func (cm *CacheManager) RemoveCrawlCache(crawl *models.Crawl) {
	for _, c := range cm.handlers {
		c.RemoveCrawlCache(crawl)
	}
}
