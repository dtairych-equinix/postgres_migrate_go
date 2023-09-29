package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	// Command-line arguments
	var (
		dbUser   string
		dbPass   string
		port     int
		sourceIP string
		dstIP    string
	)

	flag.StringVar(&dbUser, "db_user", "", "Database username")
	flag.StringVar(&dbPass, "db_pass", "", "Database password")
	flag.IntVar(&port, "port", 5432, "Database port")
	flag.StringVar(&sourceIP, "source_ip", "", "Source IP address")
	flag.StringVar(&dstIP, "dst_ip", "", "Destination IP address")

	flag.Parse()

	// Construct connection strings
	srcDB := fmt.Sprintf("postgres://%s:%s@%s:%d/template1?sslmode=disable", dbUser, dbPass, sourceIP, port)
	dstDB := fmt.Sprintf("postgres://%s:%s@%s:%d/template1?sslmode=disable", dbUser, dbPass, dstIP, port)

	// Dump source database to a file
	dumpFile := "dump.sql"
	err := dumpDatabase(srcDB, dumpFile)
	if err != nil {
		log.Fatalf("Error dumping source database: %v", err)
	}

	// Restore dump to destination database
	err = restoreDatabase(dumpFile, dstDB)
	if err != nil {
		log.Fatalf("Error restoring to destination database: %v", err)
	}

	// Cleanup
	err = os.Remove(dumpFile)
	if err != nil {
		log.Printf("Warning: failed to delete dump file: %v", err)
	}

	fmt.Println("Migration completed successfully!")
}

func dumpDatabase(connStr, outputFile string) error {
	cmd := exec.Command("pg_dump", "-f", outputFile, "-d", connStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func restoreDatabase(inputFile, connStr string) error {
	cmd := exec.Command("pg_restore", "-d", connStr, inputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
