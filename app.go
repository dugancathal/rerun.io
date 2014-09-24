package main

import (
    "fmt"
    "net/http"
	"encoding/json"
	"bytes"
	"os"
	_ "github.com/lib/pq"
	"database/sql"
)

var port = os.Getenv("PORT")

func handler(w http.ResponseWriter, r *http.Request) {
	c, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("ERROR ON ADD %s", err)
	}

	r.ParseForm()
	headers, _ := json.Marshal(r.Header)
	body, _ := json.Marshal(r.Form)
	if string(body) == "{}" {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body = buf.Bytes()
	}

	_, err = c.Exec("INSERT INTO requests(method, path, body, headers) VALUES ($1, $2, $3, $4)", r.Method, r.URL.Path, body, headers)
	if err != nil {
		fmt.Printf("ERROR ON ADD %s", err)
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	c, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("ERROR ON LIST %s", err)
	}

	fmt.Fprint(w, "<table><tr><th>Method</th><th>Path</th><th>Body</th><th>Headers</th></tr>")

	rows, err := c.Query("SELECT method, path, body, headers FROM requests")
	defer rows.Close()
	for rows.Next() {
		var method, path, body, headers string
		err = rows.Scan(&method, &path, &body, &headers)
		if err == nil {
			htmlString := "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>"
			fmt.Fprintf(w, htmlString, method, path, body, headers)
		}
	}

	if err := rows.Err(); err != nil {
        fmt.Println(err)
    }

	fmt.Fprintf(w, "</table>")
}

func main() {
	c, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Printf("ERROR ON CONNECT %s", err)
	}

	fmt.Println("Creating DB")
	_, err = c.Exec("CREATE TABLE IF NOT EXISTS requests(id SERIAL, method TEXT, path TEXT, body TEXT, headers TEXT)")
	if err != nil {
		fmt.Printf("ERROR ON CREATE %s", err)
	}
	fmt.Println("DB Created")

    http.HandleFunc("/favicon.ico", nil)
    http.HandleFunc("/requests", listHandler)
    http.HandleFunc("/", handler)

	fmt.Println("Funcs bound")
	fmt.Println("Listening on :" + port)
	err = http.ListenAndServe(":" + port, nil)
	if err != nil {
		fmt.Printf("ERROR ON LISTEN %s", err)
	}
}
