package parsy_test

import (
	"fmt"
	"log"

	"github.com/odeke-em/parsy"
)

func ExampleParse() {
	parser, err := parsy.NewParser("a/b/notes", "music", "--md5sum", "--recursive", "true", "--quiet=true", "interest/2014")
	if err != nil {
		log.Fatalf("initializing parser: %v", err)
	}

	if err := parser.Add("md5sum", parsy.TBool, true, "turn on md5 checksumming"); err != nil {
		log.Fatal(err)
	}

	if err := parser.Add("depth", parsy.TInt, 2, "the traversal depth"); err != nil {
		log.Fatal(err)
	}

	if err := parser.Add("name", parsy.TString, "", "the name to use"); err != nil {
		log.Fatal(err)
	}

	if err := parser.Add("recursive", parsy.TBool, true, "traverse recursively?"); err != nil {
		log.Fatal(err)
	}

	if err := parser.Add("frequency", parsy.TFloat32, 10.0, "the frequency of the throttle"); err != nil {
		log.Fatal(err)
	}

	if err := parser.Parse(); err != nil {
		log.Fatalf("Parse failed err: %v", err)
	}

	fmt.Printf("Recursive: %v\n", parser.Get("recursive"))
	fmt.Printf("Can Md5Sum it? %v\n", parser.Get("md5sum"))
	fmt.Printf("Depth: %d\n", parser.Get("depth"))
	fmt.Printf("Frequency: %.2f\n", parser.Get("frequency"))

	// Output:
	// Recursive: true
	// Can Md5Sum it? true
	// Depth: 2
	// Frequency: 10.00
}
