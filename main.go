package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func usage(name string) {
	fmt.Printf("Usage:\n\t%s dir1 dir2\n\nCompares the files that are present in dir1 and dir2\n", name)
	os.Exit(1)
}

func list_files(path string) map[string]bool {
	results := make(map[string]bool)

	rootLen := len(path)

	filepath.Walk(path, func(curPath string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Unable to process %s", curPath)
			os.Exit(1)
		}
		relPath := curPath
		if strings.HasPrefix(curPath, path) {
			relPath = curPath[rootLen:]
		}
		results[relPath] = true
		return nil
	})
	return results
}

func process_diffs(root1 string, set1 map[string]bool, set2 map[string]bool) int {
	changes := 0

	printed_header := false
	for entry, _ := range set1 {
		_, exists := set2[entry]
		if !exists {
			if !printed_header {
				fmt.Println("\nOnly in", root1)
				printed_header = true
			}
			fmt.Println(entry)
			changes++
		}
	}
	return changes
}

func main() {
	var count int
	if len(os.Args) != 3 {
		usage(os.Args[0])
	}
	if os.Args[1] != os.Args[2] {
		l1 := list_files(os.Args[1])
		l2 := list_files(os.Args[2])

		count = process_diffs(os.Args[1], l1, l2)
		count = count + process_diffs(os.Args[2], l2, l1)
	}
	if count == 0 {
		fmt.Println("There were no differences")
	} else {
		fmt.Println("\nThere were", count, "differences")
	}
}
