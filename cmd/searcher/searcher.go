package main

import (
	"flag"
	"fmt"
	"github.com/Kalinin-Andrey/ipindexer/pkg/apperror"
	"github.com/Kalinin-Andrey/ipindexer/pkg/ipwriter"
	"log"
	"os"

	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

var indexFile		string
var ip				string
var isHelp			bool

func init() {
	flag.StringVar(&indexFile, "file", "", "path to an index file to search in")
	flag.StringVar(&ip, "ip", "", "IP for searching")
	flag.BoolVar(&isHelp, "help", false, "Command usage")
}

func help() {
	fmt.Printf(`Command usage:
		--file (string; requared) path to an index file to search in
		--ip (string; requared) IP for searching
		--help (bool) this help`)
	return
}

func main() {
	flag.Parse()

	if isHelp {
		help()
		return
	}

	if indexFile == "" || ip == "" {
		fmt.Println("file or ip params can not be empty")
		help()
		return
	}
	err := validation.Validate(ip,
		is.IPv4,                    // is a valid IPv4
	)
	if err != nil {
		log.Fatalf("ip not valid: %s\n",err.Error())
	}

	iFile, err := os.OpenFile(indexFile , os.O_RDONLY, 0755)
	if err != nil {
		log.Fatalf("Can not open an index file")
	}
	defer iFile.Close()

	fileInfo, err := iFile.Stat()
	if err != nil {
		log.Fatalf("Can not get a stat for a index file, error: %v", err)
	}
	if size := uint(fileInfo.Size()); size < ipwriter.BytesInLn {
		log.Fatalf("Index file is empty, size: %v", fileInfo.Size())
	}

	w := ipwriter.New(iFile)

	err = w.Find(ip)
	if err != nil {
		if err == apperror.ErrNotFound {
			fmt.Printf("IP %q was not found.\n", ip)
		} else {
			fmt.Printf("An error has occurred: %q.\n", ip)
		}
	} else {
		fmt.Printf("IP %q was found!\n", ip)
	}
}