// SPDX-License-Identifier: Apache-2.0
// Copyright (C) Arm Ltd. 2022

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pborman/getopt"
)

const RawAuxTraceLen = 62

func parseAuxTraceLine(line string) ([]string, error) {
	lineArray := []rune(line)
	if len(lineArray) < RawAuxTraceLen {
		log.Println(fmt.Sprintf("%d ", len(line)) + "Invalid length of auxtrace, skipping: " + line)
		return nil, errors.New("invalid auxtrace length")
	}
	decodedTrace := string(lineArray[RawAuxTraceLen:])

	return strings.Fields(decodedTrace), nil
}

var count uint64 = 0

var branchRecs *MemoryBranchRecords = &MemoryBranchRecords{}
var ldstRecs *MemoryLoadStoreRecords = &MemoryLoadStoreRecords{}
var WriterConcurrency int64 = 8

func main() {

	inputFile := getopt.StringLong("file", 'f', "", "perf spe trace file from 'perf script -D'")
	debugFlag := getopt.BoolLong("debug", 'd', "false", "Enable debug output")
	outputPrefix := getopt.StringLong("prefix", 'p', "spe", "file prefix for output parquet file")
	WriterConcurrency = *getopt.Int64Long("concurrency", 'c', 8, "Parquet writer concurrency")
	noLdst := getopt.BoolLong("noldst", 'l', "false", "Disable LDST instructions parsing")
	noBr := getopt.BoolLong("nobr", 'b', "false", "Disable Branch instrcutions parsing")

	getopt.Parse()

	if *inputFile == "" {
		getopt.Usage()
		os.Exit(1)
	}

	log.Printf("Processing SPE trace file: %s", *inputFile)
	file, err := os.Open(*inputFile)

	if err != nil {
		log.Fatalf("Fail to open file")
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	HasSpeAuxtrace := false
	// matches "cpu: ID"
	cpuMatched := regexp.MustCompile(`\bcpu\:\s+\d+`)
	// There maybe many ARM SPE data sessions

	var cpu int32 = -1

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			log.Println("End of ARM SPE trace file")
			break
		}

		if cpuMatched.MatchString(line) {
			if !strings.Contains(line, "PERF_RECORD_AUXTRACE") {
				continue
			}
			cpuStr := cpuMatched.FindString(line)
			cpu64, err := strconv.ParseInt(strings.TrimSpace(strings.Split(cpuStr, ":")[1]), 10, 32)
			if err != nil {
				log.Printf("Failed to parse cpu id: %s\n", cpuStr)
				continue
			}
			cpu = int32(cpu64)
		}
		// skips the lines before ARM SPE data line
		if !HasSpeAuxtrace && !strings.Contains(line, "ARM SPE data") {
			continue
		}
		if !HasSpeAuxtrace {
			HasSpeAuxtrace = true
			// skip the ARM SPE data line
			if *debugFlag {
				log.Println("Begin of SPE session")
			}
			continue
		}
		if HasSpeAuxtrace && len(line) <= 1 {
			// Last line of current ARM SPE session. It is an empty line with \n
			// Turn off HasSpeAuxtrace flag to search for next session
			HasSpeAuxtrace = false
			if *debugFlag {
				log.Println("End of SPE session")
			}
			cpu = -1
			log.Printf("%d SPE trace records has been parsed\n", count)
			continue
		}
		// Processing SPE auxtrace records
		tokens, err := parseAuxTraceLine(line)
		if err != nil {
			log.Printf("%s: %s\n", err, line)
			// skip the shortened line
			continue
		}

		if tokens[0] == "PC" {
			// Find the start of an Auxtrace record, which ends with TS
			record := RecordPayload{RecUnknown, make(map[string][]string)}
			record.Data["PC"] = tokens[1:]
			hasTS := false
			for !hasTS {

				recLine, err := reader.ReadString('\n')
				if err == io.EOF {
					log.Println("End of ARM SPE trace file")
					os.Exit(0)
				}
				lineTokens, err := parseAuxTraceLine(recLine)
				if err != nil {
					log.Printf("%s: %s\n", err, recLine)
					continue
				}
				switch lineTokens[0] {
				case "PAD":
					continue
				case "TS":
					record.Data["TS"] = lineTokens[1:]
					hasTS = true
				case "LAT":
					// LAT 259 ISSUE
					if len(lineTokens) != 3 {
						log.Printf("invalid LAT packet: %s", recLine)
						continue
					}
					record.Data[lineTokens[2]] = []string{lineTokens[1]}
				default:
					record.Data[lineTokens[0]] = lineTokens[1:]
				}

			}
			record.UpdateType()

			if record.GetType() == RecUnknown {
				log.Fatalln("invalid auxtrace record" + record.ToString())
			}
			count++
			// get one spe record by now

			var debugx interface{}
			if (record.Type == RecLoad || record.Type == RecStore) && !*noLdst {
				rec, err := record.ToLoadStore(cpu)
				if err != nil {
					log.Fatalln("Invalid Load/Store Record, err: ", string(err.Error()), record.ToString())
				}
				ldstRecs.data = append(ldstRecs.data, rec)
				debugx = rec
			} else if record.Type == RecBranch && !*noBr {
				rec, err := record.ToBranch(cpu)
				if err != nil {
					log.Fatalln("Invalid Branch Record, err: ", string(err.Error()), record.ToString())

				}
				branchRecs.data = append(branchRecs.data, rec)
				debugx = rec
			}

			// debug print
			if *debugFlag && count%10000 == 0 {
				log.Println(record.ToString())
				log.Println("Load/store/Branch: " + fmt.Sprintf("%v", debugx))
			}

		}
	}
	if err == io.EOF && !HasSpeAuxtrace {
		log.Fatalf("Trace contains no Arm spe data!")
	}

	// write out parquet files
	brpfname := *outputPrefix + "-br.parquet"
	ldstpfname := *outputPrefix + "-ldst.parquet"
	if !*noBr {
		if err := branchRecs.Write(brpfname); err != nil {
			log.Fatalln("Fail to write spe branch parquet file: " + brpfname)
		}
	}

	if !*noLdst {
		if err := ldstRecs.Write(ldstpfname); err != nil {
			log.Fatalln("Fail to write spe ldst parquet file: " + ldstpfname)
		}
		log.Println("SPE trace parquet files created successfully")
	}
}
