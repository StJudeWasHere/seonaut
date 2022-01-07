package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	port = 9000
	host = "127.0.0.1"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "home.html")
}

func main() {
	/*
		http.HandleFunc("/", serveHome)

		fmt.Printf("Starting at %s on port %d...\n", host, port)

		err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
		if err != nil {
			fmt.Println(err)
		}
	*/

	db, err := sql.Open("mysql", "root:root@tcp(0.0.0.0:6306)/seo")
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("unable to reach database: %v", err)
	}
	fmt.Println("database is reachable")

	start := time.Now()

	startCrawler(db)

	fmt.Println(time.Since(start))
}

func startCrawler(db *sql.DB) {
	var crawled int
	a := os.Args[1]

	pageReport := make(chan PageReport)
	c := &Crawler{}

	u, err := url.Parse(a)
	if err != nil {
		fmt.Println(err)
		return
	}

	go c.Crawl(u, pageReport)

	for r := range pageReport {
		crawled++
		handlePageReport(r)

		res, err := db.Exec(
			"INSERT INTO pagereports (url, redirect_url, refresh, status_code, content_type, lang, title, description, robots, canonical, h1, h2, words, size) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			r.URL.String(),
			r.RedirectURL,
			r.Refresh,
			r.StatusCode,
			r.ContentType,
			r.Lang,
			r.Title,
			r.Description,
			r.Robots,
			r.Canonical,
			r.H1,
			r.H2,
			r.Words,
			len(r.Body),
		)

		if err != nil {
			log.Fatalf("could not insert row: %v", err)
			continue
		}

		lid, err := res.LastInsertId()
		if err != nil {
			log.Fatalf("Last id error %v", err)
			continue
		}

		for _, l := range r.Links {
			_, err := db.Exec("INSERT INTO links (pagereport_id, url, rel, text, external) values (?, ?, ?, ?, ?)", lid, l.URL, l.Rel, l.Text, l.External)
			if err != nil {
				log.Fatalf("could not insert row: %v", err)
			}
		}

		for _, h := range r.Hreflangs {
			_, err := db.Exec("INSERT INTO hreflangs (pagereport_id, url, lang ) values (?, ?, ?)", lid, h.URL, h.Lang)
			if err != nil {
				log.Fatalf("could not insert row: %v", err)
			}
		}

		for _, i := range r.Images {
			_, err := db.Exec("INSERT INTO images (pagereport_id, url, alt) values (?, ?, ?)", lid, i.URL, i.Alt)
			if err != nil {
				log.Fatalf("could not insert row: %v", err)
			}
		}

		for _, s := range r.Scripts {
			_, err := db.Exec("INSERT INTO scripts (pagereport_id, url) values (?, ?)", lid, s)
			if err != nil {
				log.Fatalf("could not insert row: %v", err)
			}
		}

		for _, s := range r.Styles {
			_, err := db.Exec("INSERT INTO styles (pagereport_id, url) values (?, ?)", lid, s)
			if err != nil {
				log.Fatalf("could not insert row: %v", err)
			}
		}
	}

	writer.Flush()

	fmt.Printf("%d pages crawled.\n", crawled)
}
