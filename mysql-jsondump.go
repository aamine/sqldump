package main

import(
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "encoding/json"
    "os"
    "bufio"
    "compress/gzip"
    "fmt"
    "time"
    "flag"
)

type options struct {
    host string
    port string
    user string
    password string
    database string
    query string
    gzip bool
}

func main() {
    opts := parseOptions()
    connString := opts.user + ":" + opts.password + "@tcp(" + opts.host + ":" + opts.port + ")/" + opts.database
    db, err := sql.Open("mysql", connString)
    if err != nil {
        errorExit(err.Error())
    }
    defer db.Close()

    info("[SQL] %s", opts.query)
    rows, err := db.Query(opts.query)
    if err != nil {
        errorExit(err.Error())
    }
    defer rows.Close()
    info("query returned")

    columns, err := rows.Columns()
    if err != nil {
        errorExit(err.Error())
    }
    values := make([]sql.RawBytes, len(columns))
    args := make([]interface{}, len(columns))
    for i := range values {
        args[i] = &values[i]
    }

    z := gzip.NewWriter(os.Stdout)
    f := bufio.NewWriter(z)
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
    z.Close()

    info("Total %d records", n)
}

func parseOptions() options {
    gzipOpt := flag.Bool("gzip", false, "Enables gzip compression.")
    flag.Parse()
    args := flag.Args()
    if len(args) != 6 {
        usageExit("wrong number of arguments (%v for %v)", len(args), 6)
    }
    opts := options {}
    opts.gzip = *gzipOpt
    i := 0
    opts.host = args[i]; i++
    opts.port = args[i]; i++
    opts.user = args[i]; i++
    opts.password = args[i]; i++
    opts.database = args[i]; i++
    opts.query = args[i]; i++
    return opts
}

func unmarshalValue(data sql.RawBytes) interface{} {
    if data == nil {
        return nil
    } else {
        // FIXME: better way?
        return string(data)
    }
}

func info(format string, params ...interface{}) {
    fmt.Fprintln(os.Stderr, time.Now().String() + ": " + fmt.Sprintf(format, params...))
}

func usageExit(format string, params ...interface{}) {
    printError(format, params...)
    fmt.Fprintln(os.Stderr, "Usage: mysql-jsondump HOST PORT USER PASSWORD DATABASE QUERY")
    os.Exit(1)
}

func errorExit(format string, params ...interface{}) {
    printError(format, params...)
    os.Exit(1)
}

func printError(format string, params ...interface{}) {
    fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], fmt.Sprintf(format, params...))
}
