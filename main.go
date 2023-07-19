package main

import (
	"html/template"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"strings"

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

type payload struct {
	ID     int64
	MaxID  int
	Chunk  string
	Tokens []string
	Name   string
	Author string
}

func main() {
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"upper": strings.ToUpper,
	})
	r.LoadHTMLFiles("templates/index.tmpl")
	r.StaticFile("/favicon.ico", "./assets/favicon.ico")
	r.StaticFile("/html2canvas.min.js", "./assets/html2canvas.min.js")

	randMax := big.NewInt(maxID)

	spaceRE := regexp.MustCompile(`[\t\v\f\r ]+`)
	r.HEAD("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	r.GET("/", func(c *gin.Context) {
		db, err := connectDB()
		if err != nil {
			log.Println(err.Error())
			c.String(http.StatusInternalServerError, "oh no.")
			return
		}
		defer db.Close()

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
		var dest payload
		err = row.Scan(&dest.Chunk, &dest.Name, &dest.Author)
		if err != nil {
			log.Println(err.Error())
			c.String(http.StatusInternalServerError, "oh no.")
		}

		if dest.Author == "" {
			dest.Author = "Unknown"
		}

		dest.MaxID = maxID
		dest.ID = id.Int64()

		dest.Tokens = []string{}
		for _, t := range spaceRE.Split(dest.Chunk, -1) {
			if t == "" {
				continue
			}
			if strings.Contains(t, "\n") {
				sp := strings.Split(t, "\n")
				for x, s := range sp {
					nl := "\n"
					if x == len(sp)-1 {
						nl = ""
					}
					dest.Tokens = append(dest.Tokens, s+nl)
				}
			} else {
				dest.Tokens = append(dest.Tokens, t)
			}
		}

		c.HTML(http.StatusOK, "index.tmpl", dest)
	})
	r.Run() // 8080
}
