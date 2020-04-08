package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Kalinin-Andrey/ipindexer/pkg/ipwriter"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"regexp"
)

var sourceFile		string
var sourceDate		string
var dstPath			string
var isHelp			bool

func init() {
	flag.StringVar(&sourceFile, "file", "", "path to an access log file to read from")
	flag.StringVar(&sourceDate, "date", "", "date for an index file name to write into")
	flag.StringVar(&dstPath, "dst", "example/", "path to an index file to write into")
	flag.BoolVar(&isHelp, "help", false, "Command usage")
}

func help() {
	fmt.Printf(`Command usage:
		--file (string; requared) path to an access log file to read from
		--date (string; requared) date for an index file name to write into
		--dst (string; requared) path to an index file to write into
		--help (bool) this help`)
	return
}

func main() {
	flag.Parse()

	if isHelp {
		help()
		return
	}

	if sourceFile == "" || sourceDate == "" {
		fmt.Println("file or date params can not be empty")
		help()
		return
	}
	srcFile, err := os.OpenFile(sourceFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("Can not open a source file")
	}
	defer srcFile.Close()
	dstFileName := dstPath + sourceDate + ".txt"
	dstFile, err := os.OpenFile(dstFileName , os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("Can not open a source file")
	}
	defer dstFile.Close()

	fileInfo, err := dstFile.Stat()
	if err != nil {
		log.Fatalf("Can not get a stat for a source file, error: %v", err)
	}

	w := ipwriter.New(dstFile)

	if size := uint(fileInfo.Size()); size > 0 {
		w.SetCurrentByte(size)
	}

	reader := bufio.NewReader(srcFile)

	for {
		ln, err := reader.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				if err = lineProcessing(w, ln); err != nil {
					log.Fatalf("Error occurred while processing a line: %q; error: %q", ln, err)
				}
				break
			}
			log.Fatalf("Error occurred while reading source file: %q", err)
		}
		if err = lineProcessing(w, ln); err != nil {
			log.Fatalf("Error occurred while processing a line: %q; error: %q", ln, err)
			break
		}
	}

}

func lineProcessing(w *ipwriter.IPWriter,ln string) error {
	ip := extractIP(ln)
	if ip == "" {
		return errors.Errorf("Can not extract an IP from a line: %q", ln)
	}
	return w.Write(ip)
}

func extractIP(s string) string {
	re := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)
	return re.FindString(s)
}

