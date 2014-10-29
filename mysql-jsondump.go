package main

import(
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "encoding/json"
    "os"
    "bufio"
    "strings"
    "fmt"
)

func main() {
    host := os.Args[1]
    port := os.Args[2]
    user := os.Args[3]
    password := os.Args[4]
    database := os.Args[5]
    query := os.Args[6]

    connString := user + ":" + password + "@tcp(" + host + ":" + port + ")/" + database
    db, err := sql.Open("mysql", connString)
    if err != nil {
        panic(err.Error())
    }
    defer db.Close()

    rows, err := db.Query("select * from users")
    fmt.Fprintln(os.Stderr, "query end")
    defer rows.Close()
    if err != nil {
        panic(err.Error())
    }

    columns, err := rows.Columns()
    if err != nil {
        panic(err.Error())
    }
    values := make([]sql.RawBytes, len(columns))
    args := make([]interface{}, len(values))
    for i := range args {
        args[i] = &values[i]
    }

    f := bufio.NewWriter(os.Stdout)
    n := 0
    for rows.Next() {
        err := rows.Scan(args...)
        if err != nil {
            panic(err.Error())
        }

        sep := "{"
        for i, col := range values {
            fmt.Fprint(f, sep)
            sep = ","
            if col == nil {
                fmt.Fprint(f, columns[i])
                fmt.Fprint(f, ":null")
            } else {
                val := strings.Replace(string(col), "\"", "\\\"", -1)
                fmt.Fprint(f, columns[i])
                fmt.Fprint(f, ":\"")
                fmt.Fprint(f, val)
                fmt.Fprint(f, "\"")
            }
        }
        fmt.Fprintln(f, "}")
        n++
        if n % 100000 == 0 {
            fmt.Fprintln(os.Stderr, n, "lines")
        }
    }
}
