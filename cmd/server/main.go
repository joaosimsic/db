package main

import (
	"fmt"
	"os"

	"db/internal/storage"
)

func main() {
	dbFile := "mydb.db"

	dm, err := storage.NewDiskManager(dbFile)
	if err != nil {
		panic(err)
	}

	defer dm.Close()
	defer os.Remove(dbFile)

	p1 := &storage.Page{ID: 0}

	copy(p1.Data[:], "Row 1: Alice, Row 2: Bob")

	if err := dm.WritePage(p1); err != nil {
		panic(err)
	}

	fmt.Printf("Wrote page %d on disk\n", p1.ID)

	fetchedPage, err := dm.ReadPage(0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully read Page %d from disk.\n", fetchedPage.ID)
	fmt.Printf("Data inside page: %s\n", string(fetchedPage.Data[:24]))
}
