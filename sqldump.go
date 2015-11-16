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

const ProgramVersion string = "1.0.1"

type options struct {
    driver string
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
    db, err := sql.Open(opts.driver, connString)
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
        if n % 1000000 == 0 {
            info("read %d records...", n)
        }
    }

    info("Total %d records", n)
}

func parseOptions() options {
    opts := options {format: "json"}
    flag.StringVar(&opts.driver, "driver", "mysql", "Database driver name. (default: mysql)")
    tsvOpt := flag.Bool("tsv", false, "Enables TSV output.")
    jsonOpt := flag.Bool("json", false, "Enables JSON output. (default)")
    flag.BoolVar(&opts.gzip, "gzip", false, "Enables gzip compression.")
    versionOpt := flag.Bool("version", false, "Shows version number and quit.")
    flag.Parse()
    if *versionOpt {
        fmt.Println("sqldump version " + ProgramVersion)
        os.Exit(0)
    }
    args := flag.Args()
    if len(args) != 6 {
        usageExit("wrong number of arguments (%v for %v)", len(args), 6)
    }
    if *jsonOpt {
        opts.format = "json"
    }
    if *tsvOpt {
        opts.format = "tsv"
    }
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

var controlCharTranslates []string = []string {
    "\u0000", "",
    "\u0001", "",
    "\u0002", "",
    "\u0003", "",
    "\u0004", "",
    "\u0005", "",
    "\u0006", "",
    "\u0007", "",
    "\u0008", "",
    // "\u0009", "",   // TAB
    // "\u000A", "",   // NL, \n
    "\u000B", "",
    "\u000C", "",
    // "\u000D", "",   // CR, \r
    "\u000E", "",
    "\u000F", "",
    "\u0010", "",
    "\u0011", "",
    "\u0012", "",
    "\u0013", "",
    "\u0014", "",
    "\u0015", "",
    "\u0016", "",
    "\u0017", "",
    "\u0018", "",
    "\u0019", "",
    "\u001A", "",
    "\u001B", "",
    "\u001C", "",
    "\u001D", "",
    "\u001E", "",
    "\u001F", ""}

var jsonValueReplacer *strings.Replacer =
    strings.NewReplacer(
        append(
            []string {
                "\"", "\\\"",
                "\\", "\\\\",
                "\t", "\\t",
                "\r", "\\r",
                "\n", "\\n"},
            controlCharTranslates...)...)

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

var tsvValueReplacer *strings.Replacer =
    strings.NewReplacer(
        append(
            []string {
                "\\", "\\\\",
                "\t", "\\t",
                "\r", "\\r",
                "\n", "\\n" },
            controlCharTranslates...)...)

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
    fmt.Fprintln(os.Stderr, "Usage: sqldump [--tsv] [--gzip] HOST PORT USER PASSWORD DATABASE QUERY > out.json")
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
