package main

import (
	"log"
	"math/big"
	"net/http"

	"crypto/rand"
	"database/sql"

	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dsn   = "/mnt/volume_tor1_01/gutenberg/chunker.db?cache=shared&mode=r"
	maxID = 9739473
)

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type chunk struct {
	Chunk  string
	Name   string
	Author string
}

func main() {
	r := gin.Default()
	r.LoadHTMLFiles("templates/index.tmpl")

	randMax := big.NewInt(maxID)

	r.GET("/", func(c *gin.Context) {
		db, err := connectDB()
		if err != nil {
			log.Println(err.Error())
			c.String(http.StatusInternalServerError, "oh no.")
			return
		}

		id, err := rand.Int(rand.Reader, randMax)
		if err != nil {
			log.Println(err.Error())
			c.String(http.StatusInternalServerError, "oh no.")
			return
		}

		stmt, err := db.Prepare("select c.chunk, f.name, f.author from chunks c join files f on c.sourceid = f.id where c.id = ?")
		if err != nil {
			log.Println(err.Error())
			c.String(http.StatusInternalServerError, "oh no.")
			return
		}

		row := stmt.QueryRow(id.Int64())
		var dest chunk
		err = row.Scan(&dest.Chunk, &dest.Name, &dest.Author)
		if err != nil {
			log.Println(err.Error())
			c.String(http.StatusInternalServerError, "oh no.")
		}

		c.HTML(http.StatusOK, "index.tmpl", dest)
	})
	r.Run() // 8080
}
