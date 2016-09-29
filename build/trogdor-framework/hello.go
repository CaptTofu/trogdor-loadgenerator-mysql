package main // import "hello"

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
        "database/sql"
        "github.com/go-sql-driver/mysql"
)

type Message struct {
	Message string `json:"message"`
}

var (
	debug   = flag.Bool("debug", false, "debug logging")
	message = flag.String("message", "Hello, World!", "message to return")
	port    = flag.Int("port", 9080, "port to serve on")
        mysql_port = flag.Int("mysql-port", 3306, "mysql db port")
        mysql_user = flag.String("mysql-user", "test", "mysql user")
        mysql_password = flag.String("mysql-password", "test", "mysql password")
        mysql_host = flag.String("mysql-host", "localhost", "mysql host")
        mysql_db = flag.String("mysql-db", "test", "mysql schema")
)

func main() {
	flag.Parse()
        db := connectMysql()
	runtime.GOMAXPROCS(runtime.NumCPU())

	http.HandleFunc("/json", jsonHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), Log(http.DefaultServeMux))
}

func connectMysql() {
        dsn := fmt.sprintf("%s:%s@/%s", mysql_user, mysql_password, mysql_db)
        db, err := sql.Open("mysql", "user:password@/dbname")
        return db
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if *debug == true {
			log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		}
		handler.ServeHTTP(w, r)
	})
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Message{*message})
}
