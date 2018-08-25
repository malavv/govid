package main

import (
	"encoding/json"
	"flag"
	"github.com/malavv/govid-xml"
	"io"
	"os"
)

func main() {
	inFileName := flag.String("xml", "", "XML filename")
	outFileName := flag.String("f", "", "Output filename")

	flag.Parse()

	var in io.Reader = os.Stdin
	var out io.Writer = os.Stdout

	// Infile
	if *inFileName != "" {
		infile, err := os.Open(*inFileName)
		if err != nil { panic(err) }
		defer infile.Close()
		in = infile
	}

	// Outfile
	if *outFileName != "" {
		outfile, err := os.Create(*outFileName)
		if err != nil { panic(err) }
		defer outfile.Close()
		out = outfile
	}

	// Read
	records, err := read(in)
	if err != nil { panic(err) }

	// Print
	err = write(out, records)
	if err != nil { panic(err) }
}

func read(reader io.Reader) ([]govid.TRecord, error) {
	return govid.ParseXML(reader)
}
func write(writer io.Writer, records []govid.TRecord) error {
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "    ")
	return enc.Encode(records)
}
