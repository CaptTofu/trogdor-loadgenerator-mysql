package main

import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import (
    "encoding/json"
    "flag"
    "fmt"
    "math/rand"
    "log"
    "net/http"
    "runtime"
)

type Message struct {
    Message string `json:"message"`
}
type rres struct {
    Id int
    Msg string
    Time string
}

var (
    debug   = flag.Bool("debug", false, "debug logging")
    port    = flag.Int("port", 9080, "port to serve on")
    mysql_port = flag.Int("mysql-port", 3306, "mysql db port")
    mysql_user = flag.String("mysql-user", "test", "mysql user")
    mysql_password = flag.String("mysql-password", "test", "mysql password")
    mysql_host = flag.String("mysql-host", "localhost", "mysql host")
    mysql_db = flag.String("mysql-db", "test", "mysql schema")
    writes = flag.Int("writes", 1, "write ops")
    updates = flag.Int("updates", 1, "update ops")
    reads = flag.Int("reads", 1, "read ops")
    col_len = flag.Int("col-len", 16, "length of string to insert max 5000")
)
var db *sql.DB

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Intn(len(letterBytes))]
    }
    return string(b)
}

func dbsetup() {
    dsn := fmt.Sprintf("%s:%s@/%s", *mysql_user, *mysql_password, *mysql_db)
    //db, err := sql.Open("mysql", "user:password@/dbname")
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Panic(err)
    }

    stmt, err := db.Prepare("drop table if exists t1")
    res, err := stmt.Exec()
    checkErr(err)

    stmt, err = db.Prepare("create table t1 (id int auto_increment, message varchar(5000) not null default '', created datetime, primary key (id)) engine=innodb")
    res, err = stmt.Exec()
    checkErr(err)

    id, err := res.LastInsertId()
    checkErr(err)

    fmt.Println(id)
}

func random(min int, max int) int {
    // panic if <= 1
    if max < 1 {
        max = 1
    }
    //rand.Seed(time.Now().Unix())
    rnd := rand.Intn(max)
    if rnd == 0 {
        rnd = min
    }
    return rnd
}

func do_writes(db *sql.DB) {
    for i :=0; i < *writes; i++ {
        // insert
        stmt, err := db.Prepare("INSERT into t1 SET message=?,created=now()")
        checkErr(err)

        txt := RandStringBytes(*col_len)
        _, err = stmt.Exec(txt)
        checkErr(err)
    }
}

func do_updates(db *sql.DB) {
    for i :=0; i < *updates; i++ {
        // insert
        stmt, err := db.Prepare("UPDATE t1 SET message = ? WHERE id = ?")
        checkErr(err)

        txt := RandStringBytes(*col_len)
        id := random(get_min_max(db))
        _, err = stmt.Exec(txt, id)
        checkErr(err)
    }
}

func get_min_max(db *sql.DB) (int,int) {
    var min int
    var max int
    // obtain max and min
    err := db.QueryRow("SELECT min(id), max(id) FROM t1").Scan(&min, &max)
    checkErr(err)
    return min,max
}

func do_reads(db *sql.DB) []rres {
    var id int
    var created string
    var message string
    var resstruct rres
    var resultarr []rres
    for i :=0; i < *reads ; i++ {

        rnd := random(get_min_max(db))
        fmt.Printf("rnd: %d\n", rnd)
        //fmt.Println("random num: %d", rnd)
        err := db.QueryRow("SELECT * FROM t1 where id = ?", rnd).Scan(&id, &message, &created)
        checkErr(err)

        resstruct = rres{id, message, created}
        resultarr = append(resultarr, resstruct)
    }

    return resultarr
}
func dbprocess() []rres {
    dsn := fmt.Sprintf("%s:%s@/%s", *mysql_user, *mysql_password, *mysql_db)
    //db, err := sql.Open("mysql", "user:password@/dbname")
    db, err := sql.Open("mysql", dsn)

    if err != nil {
        log.Panic(err)
    }
    // writes
    do_writes(db)
    // updates
    do_updates(db)
    // reads
    resultarr := do_reads(db)

    db.Close()
    return resultarr
}

func main() {
    flag.Parse()

    runtime.GOMAXPROCS(runtime.NumCPU())

    http.HandleFunc("/json", jsonHandler)
    http.ListenAndServe(fmt.Sprintf(":%d", *port), Log(http.DefaultServeMux))
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
    // keeping for example
    //msg := r.FormValue("msg")
    resarr := dbprocess()
    json.NewEncoder(w).Encode(resarr)
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
