package main

import(
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "os"
    "io"
    "bufio"
    "strings"
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
    format string
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

    var w io.Writer
    if opts.gzip {
        z := gzip.NewWriter(os.Stdout)
        defer z.Close()
        w = z
    } else {
        w = os.Stdout
    }
    f := bufio.NewWriter(w)
    defer f.Flush()

    var generate generatorFunction
    if opts.format == "tsv" {
        generate = generateTsv
    } else {
        generate = generateJson
    }

    n := 0
    for rows.Next() {
        err := rows.Scan(args...)
        if err != nil {
            errorExit(err.Error())
        }
        generate(f, columns, values)
        f.WriteString("\n")

        n++
        if n % 100000 == 0 {
            info("read %d records...", n)
        }
    }

    info("Total %d records", n)
}

func parseOptions() options {
    tsvOpt := flag.Bool("tsv", false, "Enables TSV output.")
    jsonOpt := flag.Bool("json", false, "Enables JSON output. (default)")
    gzipOpt := flag.Bool("gzip", false, "Enables gzip compression.")
    flag.Parse()
    args := flag.Args()
    if len(args) != 6 {
        usageExit("wrong number of arguments (%v for %v)", len(args), 6)
    }
    opts := options {format: "json"}
    if *jsonOpt {
        opts.format = "json"
    }
    if *tsvOpt {
        opts.format = "tsv"
    }
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

type generatorFunction func (f *bufio.Writer, columns []string, values []sql.RawBytes)

var jsonValueReplacer *strings.Replacer = strings.NewReplacer(
    "\"", "\\\"",
    "\t", "\\t",
    "\r", "\\r",
    "\n", "\\n")

func generateJson(f *bufio.Writer, columns []string, values []sql.RawBytes) {
    f.WriteString("{")
    sep := ""
    for i, val := range values {
        f.WriteString(sep); sep = ","
        f.WriteString("\"")
        name := columns[i]
        f.WriteString(name)
        f.WriteString("\":")
        if val == nil {
            f.WriteString("null")
        } else {
            f.WriteString("\"")
            jsonValueReplacer.WriteString(f, string(val))
            f.WriteString("\"")
        }
    }
    f.WriteString("}")
}

var tsvValueReplacer *strings.Replacer = strings.NewReplacer(
    "\t", "\\t",
    "\r", "\\r",
    "\n", "\\n")

func generateTsv(f *bufio.Writer, columns []string, values []sql.RawBytes) {
    sep := ""
    for _, val := range values {
        f.WriteString(sep); sep = "\t"
        if val != nil {
            tsvValueReplacer.WriteString(f, string(val))
        }
    }
}

func info(format string, params ...interface{}) {
    fmt.Fprintln(os.Stderr, time.Now().String() + ": " + fmt.Sprintf(format, params...))
}

func usageExit(format string, params ...interface{}) {
    printError(format, params...)
    fmt.Fprintln(os.Stderr, "Usage: mysql-jsondump [--tsv] [--gzip] HOST PORT USER PASSWORD DATABASE QUERY > out.json")
    flag.PrintDefaults()
    os.Exit(1)
}

func errorExit(format string, params ...interface{}) {
    printError(format, params...)
    os.Exit(1)
}

func printError(format string, params ...interface{}) {
    fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], fmt.Sprintf(format, params...))
}
