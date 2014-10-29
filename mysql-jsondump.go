package main

import(
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "encoding/json"
    "os"
    "bufio"
    "fmt"
    "time"
)

func main() {
    if len(os.Args) != 7 {
        errorExit(fmt.Sprintf("%s: wrong number of arguments (%v for %v)", os.Args[0], len(os.Args), 6))
    }
    host := os.Args[1]
    port := os.Args[2]
    user := os.Args[3]
    password := os.Args[4]
    database := os.Args[5]
    query := os.Args[6]

    connString := user + ":" + password + "@tcp(" + host + ":" + port + ")/" + database
    db, err := sql.Open("mysql", connString)
    if err != nil {
        errorExit(err.Error())
    }
    defer db.Close()

    info("[SQL] %s", query)
    rows, err := db.Query(query)
    info("query returned")
    defer rows.Close()
    if err != nil {
        errorExit(err.Error())
    }

    columns, err := rows.Columns()
    if err != nil {
        errorExit(err.Error())
    }
    values := make([]sql.RawBytes, len(columns))
    args := make([]interface{}, len(columns))
    for i := range values {
        args[i] = &values[i]
    }

    f := bufio.NewWriter(os.Stdout)
    rec := make(map[string]interface{})
    n := 0
    for rows.Next() {
        err := rows.Scan(args...)
        if err != nil {
            errorExit(err.Error())
        }

        for i, val := range values {
            rec[columns[i]] = unmarshalValue(val)
        }
        data, err := json.Marshal(rec)
        if err != nil {
            errorExit(err.Error())
        }
        _, err = f.Write(data)
        if err != nil {
            errorExit(err.Error())
        }
        f.WriteString("\n")

        n++
        if n % 100000 == 0 {
            info("read %d records...", n)
        }
    }
    f.Flush()

    info("Total %d records", n)
}

func unmarshalValue(data sql.RawBytes) interface{} {
    if data == nil {
        return nil
    } else {
        // FIXME
        return string(data)
    }
}

func info(format string, values ...interface{}) {
    fmt.Fprintln(os.Stderr, time.Now().String() + ": " + fmt.Sprintf(format, values))
}

func errorExit(msg string) {
    // FIXME
    panic(msg)
}
