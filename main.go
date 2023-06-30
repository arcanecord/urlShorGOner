package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func createTable(db *sql.DB) {
	urls_table := `CREATE TABLE urls(
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"short_url" TEXT,
        "original_url" TEXT);`

	query, err := db.Prepare(urls_table)
	if err != nil {
		log.Fatal(err)
	}
	query.Exec()
	fmt.Println("Table created successfully!")
}
func addUrl(db *sql.DB, short_url string, original_url string) {
	insert_url := `INSERT INTO urls(short_url, original_url) VALUES (?, ?)`
	query, err := db.Prepare(insert_url)
	if err != nil {
		log.Fatal(err)
	}
	query.Exec(short_url, original_url)
	fmt.Println("Url added successfully! ", short_url, " ", original_url)
}
func getUrl(db *sql.DB, short_url string) string {
	var original_url string
	query := `SELECT original_url FROM urls WHERE short_url = ?`
	row := db.QueryRow(query, short_url)
	row.Scan(&original_url)
	return original_url
}

func createShortUrl(original_url string) string {
	h := sha256.New()
	h.Write([]byte(original_url))
	hashed_url := fmt.Sprintf("%x", h.Sum(nil))
	return hashed_url[:8]
}
func main() {

	if _, err := os.Stat("database.db"); os.IsNotExist(err) {
		file, err := os.Create("database.db")
		if err != nil {
			log.Fatal(err)
		}
		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Fatal(err)
		}
		createTable(db)
		file.Close()
	}

	r := gin.Default()
	r.LoadHTMLGlob("index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.POST("/shorten", func(c *gin.Context) {
		original_url := c.PostForm("url")
		if original_url == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Url is empty",
			})
			return
		}

		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		short_url := createShortUrl(original_url)
		addUrl(db, short_url, original_url)
		// c.JSON(http.StatusOK, gin.H{
		// 	"short_url": short_url,
		// })
		// redirect to the index page with the short url in the url
		c.Redirect(http.StatusMovedPermanently, "/?shortenedUrl="+short_url)
	})
	r.GET("/:short_url", func(c *gin.Context) {
		short_url := c.Param("short_url")
		db, err := sql.Open("sqlite3", "./database.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		original_url := getUrl(db, short_url)
		c.Redirect(http.StatusMovedPermanently, original_url)
	})

	r.Run() // listen and serve on
}
