package main

import (
	"compress/gzip"
	"database/sql"
	"flag"
	"fmt"
	csv "github.com/JensRantil/go-csv"
	"github.com/JensRantil/go-csv/dialect"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"os"
)

// Queryable interface that matches sql.DB and sql.Tx.
type queryable interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func dump(tables []string, db queryable, outputDir string, compressOut bool, skipHeader bool, csvDialect *csv.Dialect) error {
	for _, table := range tables {
		err := dumpTable(table, db, outputDir, compressOut, skipHeader, csvDialect)
		if err != nil {
			fmt.Printf("Error dumping %s: %s\n", table, err)
		}
	}
	return nil
}

func dumpTable(table string, db queryable, outputDir string, compressOut, skipHeader bool, csvDialect *csv.Dialect) error {
	fname := outputDir + "/" + table + ".csv"
	if compressOut {
		fname = fname + ".gz"
	}

	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	var out io.Writer
	if compressOut {
		gzout := gzip.NewWriter(f)
		defer gzout.Close()
		out = gzout
	} else {
		out = f
	}

	w := csv.NewDialectWriter(out, *csvDialect)

	rows, err := db.Query("SELECT * FROM " + table) // Couldn't get placeholder expansion to work here
	if err != nil {
		return err
	}

	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}
	if !skipHeader {
		err = w.Write(columns) // Header
		if err != nil {
			return err
		}
	}

	for rows.Next() {
		// Shamelessly ripped (and modified) from http://play.golang.org/p/jxza3pbqq9

		// Create interface set
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		// Scan for arbitrary values
		err = rows.Scan(scanArgs...)
		if err != nil {
			return err
		}

		// Print data
		csvData := make([]string, 0, len(values))
		for _, value := range values {
			switch value.(type) {
			default:
				s := fmt.Sprintf("%s", value)
				csvData = append(csvData, string(s))
			}
		}
		err = w.Write(csvData)
		if err != nil {
			return err
		}
	}

	w.Flush()
	err = w.Error()
	if err != nil {
		return err
	}

	return nil
}

func getTables(db queryable) ([]string, error) {
	tables := make([]string, 0, 10)
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}
	return tables, nil
}

func main() {
	dialectBuilder := dialect.FromCommandLine()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [table#1 table#2 ... table#N]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "If no tables are specified, all tables are dumped.\n")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	dbUser := flag.String("user", "root", "database user")
	dbPassword := flag.String("password", "", "database password")
	dbHost := flag.String("hostname", "", "database host")
	dbPort := flag.Int("port", 3306, "database port")
	outputDir := flag.String("outdir", "", "where output will be stored")
	//compressCon := flag.Bool("compress-con", false, "whether compress connection or not")
	compressFiles := flag.Bool("compress-file", false, "whether compress connection or not")
	useTransaction := flag.Bool("single-transaction", true, "whether to wrap everything in a transaction or not.")
	skipHeader := flag.Bool("skip-header", false, "whether column header should be included or not")

	flag.Parse()

	csvDialect, err := dialectBuilder.Dialect()
	if err != nil {
		fmt.Println(err.Error())
		flag.PrintDefaults()
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Database name must be defined.")
		flag.PrintDefaults()
		os.Exit(1)
	}
	dbName := args[0]

	dbUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", *dbUser, *dbPassword, *dbHost, *dbPort, dbName)
	//fmt.Println("DB url:", dbUrl)
	db, err := sql.Open("mysql", dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to server: %s\n", err)
	}
	defer db.Close()

	var q queryable
	if *useTransaction {
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		defer tx.Rollback()
		q = tx
	} else {
		q = db
	}

	var tables []string
	if len(args) > 1 {
		tables = args[1:]
	} else {
		tables, err = getTables(q)
	}

	err = dump(tables, q, *outputDir, *compressFiles, *skipHeader, csvDialect)
	if err != nil {
		panic(err)
	}
}
