package parsy_test

import (
	"testing"

	"github.com/odeke-em/parsy"
)

func TestParserWithArgs(t *testing.T) {
	parser, err := parsy.NewParser("influx", "--outfile", "", "outf.txt", "--depth", "-1", "--name=gopher")
	if err != nil {
		t.Fatalf("initialization failed. err: %v", err)
	}

	if err := parser.AddCommand("o", "outfile", parsy.TString, "", "where to place the generate binary"); err != nil {
		t.Fatal(err)
	}

	if err := parser.AddCommand("d", "depth", parsy.TInt, 2, "the traversal depth"); err != nil {
		t.Fatal(err)
	}

	if err := parser.AddCommand("", "name", parsy.TString, "", "the name to use"); err != nil {
		t.Fatal(err)
	}

	if err := parser.AddCommand("", "frequency", parsy.TInt, 10, "the frequency of the throttle"); err != nil {
		t.Fatal(err)
	}

	if err := parser.AddCommand("", "frequency", parsy.TFloat32, 10, "the frequency of the throttle"); err == nil {
		t.Fatalf("adding a duplicate command should fail")
	}

	if err := parser.Parse(); err != nil {
		t.Fatalf("Parse failed err: %v", err)
	}

	tests := [...]struct {
		key  string
		want interface{}
	}{
		0: {"frequency", 10}, // Ensure that we have the default set
		1: {"outfile", "outf.txt"},
		2: {"depth", 2},
		3: {"name", "gopher"},
	}

	for i, tt := range tests {
		got, err := parser.Value(tt.key)
		if err != nil {
			t.Errorf("#%d: key=%q got err: %v\n", i, tt.key, err)
			continue
		}

		if tt.want != got {
			t.Errorf("#%d: key=%q got: %v want: %v", i, tt.key, got, tt.want)
		}
	}
}
