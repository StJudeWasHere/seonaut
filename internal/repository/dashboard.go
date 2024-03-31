package repository

import (
	"database/sql"
	"log"
	"sort"

	"github.com/stjudewashere/seonaut/internal/models"
)

type DashboardRepository struct {
	DB *sql.DB
}

// CountByCanonical returns the number of pagereports in a crawl that are of type "text/html"
// and have an empty canonical or a canonical pointing to its own url.
func (ds *DashboardRepository) CountByCanonical(cid int64) int {
	query := `
		SELECT
			count(id)
		FROM pagereports 
		WHERE crawl_id = ? AND media_type = "text/html" AND (canonical = "" OR canonical = url)
			AND status_code >= 200 AND status_code < 300
	`

	row := ds.DB.QueryRow(query, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountByCanonical: %v\n", err)
	}

	return c
}

// CountByNonCanonical returns the number of pagereports in a crawl that are of type "text/html"
// and have a non empty canonical or a canonical pointing to a different url.
func (ds *DashboardRepository) CountByNonCanonical(cid int64) int {
	query := `
		SELECT
			count(id)
		FROM pagereports 
		WHERE crawl_id = ? AND media_type = "text/html" AND canonical != "" AND canonical != url
			AND status_code >= 200 AND status_code < 300
	`

	row := ds.DB.QueryRow(query, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("CountByNonCanonical: %v\n", err)
	}

	return c
}

// CountImagesAlt returns an AltCount model with the total number of images that have an alt attribute
// and the total number of images that don't.
func (ds *DashboardRepository) CountImagesAlt(cid int64) *models.AltCount {
	query := `
		SELECT 
			if(alt = "", "no alt", "alt") as a,
			count(*)
		FROM images
		WHERE crawl_id = ?
		GROUP BY a
	`

	c := &models.AltCount{}

	rows, err := ds.DB.Query(query, cid)
	if err != nil {
		log.Println(err)
		return c
	}

	for rows.Next() {
		var v int
		var a string

		err := rows.Scan(&a, &v)
		if err != nil {
			continue
		}

		if a == "alt" {
			c.Alt = v
		} else {
			c.NonAlt = v
		}
	}

	return c
}

// CountScheme returns an SchemeCount model with the total number of pagereports that use the
// http and the total number of pagereports that use https.
func (ds *DashboardRepository) CountScheme(cid int64) *models.SchemeCount {
	query := `
		SELECT
			scheme,
			count(*)
		FROM pagereports
		WHERE crawl_id = ?
		GROUP BY scheme
	`

	c := &models.SchemeCount{}

	rows, err := ds.DB.Query(query, cid)
	if err != nil {
		log.Println(err)
		return c
	}

	for rows.Next() {
		var v int
		var a string

		err := rows.Scan(&a, &v)
		if err != nil {
			continue
		}

		if a == "https" {
			c.HTTPS = v
		} else {
			c.HTTP = v
		}
	}

	return c
}

// CountByMediaType returns a CountList model with the total number of pagereports by media type.
func (ds *DashboardRepository) CountByMediaType(cid int64) *models.CountList {
	query := `
		SELECT media_type, count(*)
		FROM pagereports
		WHERE crawl_id = ? AND crawled = 1
		GROUP BY media_type`

	return ds.countListQuery(query, cid)
}

// CountByStatusCode returns a CountList model with the total number of pagereports by status code.
func (ds *DashboardRepository) CountByStatusCode(cid int64) *models.CountList {
	query := `
		SELECT
			status_code,
			count(*)
		FROM pagereports
		WHERE crawl_id = ? AND crawled = 1
		GROUP BY status_code`

	return ds.countListQuery(query, cid)
}

// countListQuery is a helper function used to build the CountList model.
func (ds *DashboardRepository) countListQuery(query string, cid int64) *models.CountList {
	m := models.CountList{}
	rows, err := ds.DB.Query(query, cid)
	if err != nil {
		log.Println(err)
		return &m
	}

	for rows.Next() {
		c := models.CountItem{}
		err := rows.Scan(&c.Key, &c.Value)
		if err != nil {
			log.Println(err)
			continue
		}
		m = append(m, c)
	}

	sort.Sort(sort.Reverse(m))

	return &m
}

// GetStatusCodeByDepth returns a slice of StatusCodeByDepth models with the total number of
// pagereports by depth and status code.
func (ds *DashboardRepository) GetStatusCodeByDepth(cid int64) []models.StatusCodeByDepth {
	query := `
	SELECT
		d.depth,
		COALESCE(SUM(CASE WHEN pr.status_code BETWEEN 0 AND 199 THEN 1 ELSE 0 END), 0) AS status_0_to_199,
		COALESCE(SUM(CASE WHEN pr.status_code BETWEEN 200 AND 299 THEN 1 ELSE 0 END), 0) AS status_200_to_299,
		COALESCE(SUM(CASE WHEN pr.status_code BETWEEN 300 AND 399 THEN 1 ELSE 0 END), 0) AS status_300_to_399,
		COALESCE(SUM(CASE WHEN pr.status_code BETWEEN 400 AND 499 THEN 1 ELSE 0 END), 0) AS status_400_to_499,
		COALESCE(SUM(CASE WHEN pr.status_code >= 500 THEN 1 ELSE 0 END), 0) AS status_500_and_above
	FROM
		(SELECT 1 AS depth
		UNION SELECT 2
		UNION SELECT 3
		UNION SELECT 4
		UNION SELECT 5
		UNION SELECT 6
		UNION SELECT 7
		UNION SELECT 8) d
	LEFT JOIN pagereports pr ON pr.depth = d.depth AND pr.crawl_id = ?
	GROUP BY d.depth
	ORDER BY d.depth;
	`

	s := []models.StatusCodeByDepth{}

	rows, err := ds.DB.Query(query, cid)
	if err != nil {
		log.Println(err)
		return s
	}

	for rows.Next() {
		c := models.StatusCodeByDepth{}
		err := rows.Scan(&c.Depth, &c.StatusCode100, &c.StatusCode200, &c.StatusCode300, &c.StatusCode400, &c.StatusCode500)
		if err != nil {
			log.Println(err)
			continue
		}
		s = append(s, c)
	}

	return s
}
