package main

import (
  "database/sql"
  "encoding/csv"
  "flag"
  "fmt"
  "os"
  _ "github.com/go-sql-driver/mysql"
)

func dump(tables []string, db *sql.DB, outputDir string) error {
  for _, table := range tables {
    err := dumpTable(table, db, outputDir)
    if err != nil {
      fmt.Printf("Error dumping %s: %s\n", table, err)
    }
  }
  return nil
}

func dumpTable(table string, db *sql.DB, outputDir string) error {
  f, err := os.Create(outputDir + "/" + table + ".csv")
  if err != nil {
    return err
  }
  defer f.Close()

  w := csv.NewWriter(f)

  rows, err := db.Query("SELECT * FROM " + table)   // Couldn't get placeholder expansion to work here
  if err != nil {
    return err
  }

  columns, err := rows.Columns()
  if err != nil {
    panic(err.Error())
  }
  err = w.Write(columns)   // Header
  if err != nil {
    return err
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

func getTables(db *sql.DB) ([]string, error) {
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
  dbUser := flag.String("user", "root", "database user")
  dbPassword := flag.String("password", "", "database password")
  dbHost := flag.String("hostname", "", "database host")
  dbPort := flag.Int("port", 3306, "database port")
  outputDir := flag.String("outdir", "", "where output will be stored")
  //csvSep := flag.String("fields-terminated-by", "\t", "character to terminate fields by")
  //csvOptEncloser := flag.String("fields-optionally-enclosed-by", "\"", "character to enclose fields with when needed")
  //csvEscape := flag.String("fields-escaped-by", "\\", "character to escape special characters with")
  //compressCon := flag.Bool("compress-con", false, "whether compress connection or not")
  //compressFiles := flag.Bool("compress-file", false, "whether compress connection or not")

  flag.Parse()
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

  tx, err := db.Begin()
  if err != nil {
    panic(err)
  }
  defer tx.Rollback()

  var tables []string
  if len(args) > 1 {
    tables = args[1:]
  } else {
    tables, err = getTables(db)
  }

  err = dump(tables, db, *outputDir)
  if err != nil {
    panic(err)
  }
}
