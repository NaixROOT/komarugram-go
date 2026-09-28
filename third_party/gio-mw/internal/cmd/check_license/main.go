// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const licenseHeader = "// SPDX-License-Identifier: Unlicense OR MIT"

func main() {
	log.Println("Checking license headers...")
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		hasLicense, err := fileContainsLicense(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
		}
		if !hasLicense {
			fmt.Printf("%s\n", path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
		os.Exit(1)
	}
}

func fileContainsLicense(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file %s: %v\n", filename, err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), licenseHeader) {
			return true, nil
		}
	}
	return false, scanner.Err()
}
