// Package test provides ...
package test

import (
	"database/sql"
	"net"
	"os"
)

func noResourceOpened() {} // OK

func fileOpenedAndClosed() error { // OK
	f, err := os.Open("file.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func fileOpenedNotClosed() error { 
	f, err := os.Open("file.txt") // MATCH /check if the resource is closed: f, err := os.Open("file.txt")/
	if err != nil {
		return err
	}
	return nil
}

func networkConnOpenedAndClosed() error { // OK
	conn, err := net.Dial("tcp", "example.com:80")
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func networkConnOpenedNotClosed() error { 
	conn, err := net.Dial("tcp", "example.com:80") // MATCH /check if the resource is closed: conn, err := net.Dial("tcp", "example.com:80")/
	if err != nil {
		return err
	}
	return nil
}

func multipleResourcesOpenedAllClosed() error { // OK
	f, err := os.Open("file.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	conn, err := net.Dial("tcp", "example.com:80")
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

func multipleResourcesOpenedOneNotClosed() error { 
	f, err := os.Open("file.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	f, err := os.ReadFile("file.txt") // MATCH /check if the resource is closed: f, err := os.ReadFile("file.txt")/

	conn, err := net.Dial("tcp", "example.com:80") // MATCH /check if the resource is closed: conn, err := net.Dial("tcp", "example.com:80")/
	if err != nil {
		return err
	}
	// Missing defer conn.Close()

	return nil
}

func resourceOpenedInNestedFunction() error {
	var db *sql.DB
	openDB := func() error {
		var err error
		db, err = sql.Open("mysql", "user:password@/dbname") // MATCH /check if the resource is closed: db, err = sql.Open("mysql", "user:password@/dbname")/
		return err
	}
	
	if err := openDB(); err != nil {
		return err
	}
	// Missing defer db.Close()

	return nil
}

func resourceOpenedAndClosedWithoutDefer() error {
	f, err := os.Open("file.txt") 
	if err != nil {
		return err
	}
	
	// Do something with f
	
	f.Close() // Closed, but not with defer, should pass
	return nil
}

func resourceOpenedInIfStatementAndClosed() error { // OK
	if f, err := os.Open("file.txt"); err == nil {
		defer f.Close()
		// Do something with f
	} else {
		return err
	}
	return nil
}

func resourceOpenedInIfStatementNotClosed() error {
	if f, err := os.Open("file.txt"); err == nil { // MATCH /check if the resource is closed: f, err := os.Open("file.txt")/
		// Do something with f
		// Missing defer f.Close()
	} else {
		return err
	}
	return nil
}