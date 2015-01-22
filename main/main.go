package main

import (
	"bufio"
	"flag"
	"github.com/jacobhaven/sleepsort"
	"os"
	"strconv"
	"time"
)

func main() {
	var timeStep time.Duration
	var filename string
	flag.StringVar(&filename, "", value, usage)
	flag.DurationVar(&timeStep, "speed", 0, "Custom rate at which")
	numbersFile := flag.String("file", "", "File containing numbers to sort")
	if numbersFile != "" {
		if numbers, err := os.Open(*numbersFile); err != nil {
			log.Fatalf("Couldn't read input file: %s\n", err)
		}
		defer close(numbers)
		scanner := bufio.NewScanner(numbers)
		scanner.Split(bufio.ScanWords)
		array = make([]int64, len(numbers)/2)
		for scanner.Scan() {
			n, err := strconv.ParseInt(scanner.Text(), 10, 64)
			if err != nil {
				append(array, n)
			}
		}
		if scanner.Err(); err != nil {
			log.Fatalf("Couldn't parse input file: %s\n", err)
		}
	}

	NewSleepSorter(
		NewArrayIterator(array),
		time.Nanosecond*1<<timeStep,
	).Run()
}
