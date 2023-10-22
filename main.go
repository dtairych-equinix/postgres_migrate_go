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
		dbName   string
		port     int
		sourceIP string
		dstIP    string
	)

	flag.StringVar(&dbUser, "db_user", "postgres", "Database username")
	flag.StringVar(&dbPass, "db_pass", "", "Database password")
	flag.StringVar(&dbName, "db_name", "mydatabase", "Name of database to migrate")
	flag.IntVar(&port, "port", 5432, "Database port")
	flag.StringVar(&sourceIP, "source_ip", "localhost", "Source IP address")
	flag.StringVar(&dstIP, "dst_ip", "", "Destination IP address")

	flag.Parse()

	// Construct connection strings
	srcDB := fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable", dbUser, sourceIP, port, dbName)
	dstDB := fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable", dbUser, dstIP, port, dbName)

	// Dump source database to a file
	dumpFile := "dump.sql"
	log.Println("Dumping source database...")
	err := dumpDatabase(srcDB, dumpFile, dbPass)
	if err != nil {
		log.Fatalf("Error dumping source database: %v", err)
	}

	// Transfer dump file to remote server
	remoteDumpFile := fmt.Sprintf("/tmp/%s", dumpFile)
	log.Printf("Transferring dump file to remote server: %s\n", dstIP)
	err = transferFile(dumpFile, remoteDumpFile, dstIP, dbUser)
	if err != nil {
		log.Fatalf("Error transferring dump file to remote server: %v", err)
	}

	// Cleanup local dump file
	err = os.Remove(dumpFile)
	if err != nil {
		log.Printf("Warning: failed to delete local dump file: %v", err)
	}

	// Restore dump to destination database
	log.Println("Restoring dump to destination database...")
	err = restoreDatabase(remoteDumpFile, dstDB, dbPass)
	if err != nil {
		log.Fatalf("Error restoring to destination database: %v", err)
	}

	// Cleanup remote dump file
	log.Println("Cleaning up remote dump file...")
	cleanupCmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", fmt.Sprintf("%s@%s", dbUser, dstIP), "rm", remoteDumpFile)
	cleanupCmd.Stdout = os.Stdout
	cleanupCmd.Stderr = os.Stderr
	err = cleanupCmd.Run()
	if err != nil {
		log.Printf("Warning: failed to delete remote dump file: %v", err)
	}

	fmt.Println("Migration completed successfully!")
}

func dumpDatabase(connStr, outputFile, dbPass string) error {
	cmd := exec.Command("pg_dump", "-f", outputFile, "-d", connStr)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func restoreDatabase(inputFile, connStr, dbPass string) error {
	cmd := exec.Command("psql", "-d", connStr, "-f", inputFile)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbPass)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func transferFile(localFile, remoteFile, remoteHost, user string) error {
	cmd := exec.Command("scp", "-o", "StrictHostKeyChecking=no", localFile, fmt.Sprintf("%s@%s:%s", user, remoteHost, remoteFile))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
