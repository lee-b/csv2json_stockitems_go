//
// Converts a CSV-formatted stock item file to JSON format.
//
// This will handle large data files, bigger than memory.
//
// Build it, and then run it with no arguments for usage
// information.
//
// Copyright (c) 2014 by Lee Braiden (leebraid@gmail.com)
//

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"lbgotest.com/lb/stockdatalib"
	"os"
	"strings"
)

// Don't know if there's a better way to get the proper
// exit codes for the os in question in Go.  I figure this
// will do for all intents and purposes.
const (
	EXIT_SUCCESS = 0
	EXIT_ERROR   = 1
)

// small struct to hold a few configuration variables
type Config struct {
	SrcFile string
	DstFile string
	Verbose bool
}

func exitUsage(msg *string) {
	if msg != nil {
		fmt.Fprintf(os.Stderr, "ERROR:\n\n\t%s\n\n", *msg)
	}

	fmt.Fprintf(os.Stderr, "Usage:\n\n")
	fmt.Fprintf(os.Stderr, "\t%s srcFile dstFile\n\n", os.Args[0])

	os.Exit(EXIT_ERROR)
}

func exitError(err error) {
	fmt.Fprintf(os.Stderr, "ERROR:\n\n\t%s\n\n", err)
	os.Exit(EXIT_ERROR)
}

// read the config / operational arguments from the command line
func getConfig() *Config {
	// NOTE: Exits with usage information if config is wrong,
	//       rather than returning, so no need for an error
	//       type to be returned along with the config

	var conf Config

	fileArgsSeen := 0

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--verbose":
			conf.Verbose = true

		default:
			if strings.HasPrefix(arg, "--") {
				err_msg := fmt.Sprintf("Unrecognised command-line option flag: '%s'", arg)
				exitUsage(&err_msg)
			}

			if fileArgsSeen == 0 {
				conf.SrcFile = arg
			} else if fileArgsSeen == 1 {
				conf.DstFile = arg
			} else {
				err_msg := fmt.Sprintf("Too many file arguments given.  extra arg is '%s'", arg)
				exitUsage(&err_msg)
			}

			fileArgsSeen = fileArgsSeen + 1
		}
	}

	if fileArgsSeen < 2 {
		err_msg := "Not enough file arguments given."
		exitUsage(&err_msg)
	}

	return &conf
}

// Main record-conversion loop, which opens files, reads items, writes them,
// and closes files again.
func doConversion(srcFile string, dstFile string, verbose bool) error {
	srcFp, err := os.Open(srcFile)
	if err != nil {
		exitError(err)
	} else {
		defer srcFp.Close()
	}

	dstFp, err := os.Create(dstFile)
	if err != nil {
		exitError(err)
	} else {
		defer dstFp.Close()
	}

	// create the CSV reader
	csvReader := csv.NewReader(srcFp)

	// make sure the basic file format is correct (has title row, valid field names)
	columns_err := stockdatalib.VerifyCsvFields(*csvReader)
	if columns_err != nil {
		exitError(columns_err)
	}

	// begin the json output file
	dstFp.WriteString("[\n    ")

	at_least_one_item_written := false

	// read rows one at a time (streaming, low mem usage), and write them out as json
	for {
		//
		// Create a stock item.  Normally you'd do this outside the loop and re-use it,
		// but I want to be sure I get a new, clean item each time before loading data
		// into it.  Go will (should!) take care of optimising this to use the same
		// memory each time, with minimal data clearing.
		//
		var stock_item stockdatalib.StockItem

		// read the item
		read_err := stock_item.ReadItem(*csvReader)
		if read_err != nil {
			if read_err == io.EOF {
				break
			} else {
				exitError(read_err)
			}
		}

		// Let the user know what we're currently processing, if verbose mode
		// is enabled.
		if verbose {
			fmt.Fprintf(os.Stderr, "   Item %10d: %-50s... ", stock_item.Item_id, stock_item.Description)
		}

		// write the json
		json_bytes, encode_err := json.MarshalIndent(stock_item, "    ", "    ")
		if encode_err != nil {
			exitError(encode_err)
		}

		_, write_err := dstFp.Write(json_bytes)
		if write_err != nil {
			exitError(write_err)
		}

		// finish the current status line, if in verbose mode
		if verbose {
			fmt.Fprintf(os.Stderr, "converted.\n")
		}

		// write the record separator if needed, and if not,
		// flag that we'll need one next time
		if at_least_one_item_written {
			// separate items
			dstFp.WriteString(",\n    ")
		} else {
			at_least_one_item_written = true
		}
	}

	// end the json output file
	dstFp.WriteString("\n]\n")

	return nil
}

func main() {
	conf := getConfig()

	if conf.Verbose {
		fmt.Printf("Reading from CSV file %s\n", conf.SrcFile)
		fmt.Printf("Writing to JSON file %s\n", conf.DstFile)
	}

	err := doConversion(conf.SrcFile, conf.DstFile, conf.Verbose)

	if err != nil {
		exitError(err)
	} else {
		if conf.Verbose {
			fmt.Printf("done.\n")
		}
	}

	// if we reach here, all went well
	os.Exit(EXIT_SUCCESS)
}
