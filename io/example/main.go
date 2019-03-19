package main

import (
	"fmt"
	"io"
	"os"

	structs "github.com/fatih/structs"
	atomicio "github.com/multiverse-os/atomic/io"
)

func main() {
	f, err := atomicio.Create("test-atomic-io", 0644)
	fmt.Println("File name is:", f.Name())
	if err != nil {
		// ...
	}
	defer f.Close()

	fmt.Println("Original file name is:", f.OriginalName())
	fmt.Println("File is:", f.Name())

	//fmt.Println("File path is:", f.Path())
	fileMap := structs.New(f)
	fmt.Println("Struct data is:", fileMap.Names())
	fmt.Println("Struct data is:", fileMap.Fields())
	fmt.Println("Struct data is:", fileMap.Map())

	_, err = io.WriteString(f, "Hello world")
	if err != nil {
		// ...
	}
	fmt.Println("[PRE COMMIT] File name is:", f.Name())

	err = f.Commit()
	if err != nil {
		// ...
	}
	fmt.Println("[POST COMMIT] File name is:", f.Name())
	fmt.Println("Attempting to clean up the atomic file")
	os.Remove(f.Name())
	if _, err := os.Stat(f.Name()); os.IsNotExist(err) {
		fmt.Println("Success! The file has been cleaned up")
	} else {
		fmt.Println("FAIL! For some reason the file was not able to be removed.")
	}
}
