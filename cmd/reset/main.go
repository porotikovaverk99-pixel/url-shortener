package main

import (
	"flag"
	"log"
	"path/filepath"
)

func main() {
	flag.Parse()
	root := "."
	if len(flag.Args()) > 0 {
		root = flag.Args()[0]
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		log.Fatalf("failed to get absolute path: %v", err)
	}

	pkgs, err := scanPackages(absRoot)
	if err != nil {
		log.Fatalf("failed to scan packages: %v", err)
	}

	for _, pkg := range pkgs {
		if err := generateResetFile(pkg); err != nil {
			log.Printf("failed to generate reset for %s: %v", pkg.Path, err)
		}
	}
}
