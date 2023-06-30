package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func createTable(db *sql.DB) {
	urls_table := `CREATE TABLE IF NOT EXISTS urls(
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"short_url" TEXT,
        "original_url" TEXT);`

	_, err := db.Exec(urls_table)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Table created successfully!")
}
func addUrl(db *sql.DB, short_url string, original_url string) {
	insertURL := `INSERT INTO urls (short_url, original_url) VALUES (?, ?)`
	query, err := db.Prepare(insertURL)
	if err != nil {
		log.Fatal(err)
	}
	defer query.Close()

	_, err = query.Exec(short_url, original_url)
	if err != nil {
		log.Fatal(err)
	}
}
func getUrl(db *sql.DB, short_url string) string {
	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_url = ?`
	row := db.QueryRow(query, short_url)
	err := row.Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle case when no rows are found
		} else {
			log.Fatal(err)
		}
	}
	return originalURL
}

func createShortUrl(original_url string) string {
	h := sha256.New()
	h.Write([]byte(original_url))
	hashed_url := fmt.Sprintf("%x", h.Sum(nil))
	return hashed_url[:8]
}

// Returns PORT from environment if found, defaults to
// value in `port` parameter otherwise. The returned port
// is prefixed with a `:`, e.g. `":3000"`.
func envPortOr(port string) string {
	// If `PORT` variable in environment exists, return it
	if envPort := os.Getenv("PORT"); envPort != "" {
		return ":" + envPort
	}
	// Otherwise, return the value of `port` variable from function argument
	return ":" + port
}
func main() {

	db, err := sql.Open("mysql", "mysql://root:ja5HlxXH9WDlANfO4FV0@containers-us-west-202.railway.app:7900/railway")
	if err != nil {
		log.Fatal(err)
	}
	createTable(db)

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

		db, err := sql.Open("mysql", "mysql://root:ja5HlxXH9WDlANfO4FV0@containers-us-west-202.railway.app:7900/railway")
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
		db, err := sql.Open("mysql", "mysql://root:ja5HlxXH9WDlANfO4FV0@containers-us-west-202.railway.app:7900/railway")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		original_url := getUrl(db, short_url)
		c.Redirect(http.StatusMovedPermanently, original_url)
	})
	r.Run(envPortOr("3030")) // listen and serve on
}
