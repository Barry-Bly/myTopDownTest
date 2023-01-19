// SPDX-License-Identifier: Apache-2.0
// Copyright (C) Arm Ltd. 2022

package main

import (
	"log"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

type Branch struct {
	Cpu       int32  `parquet:"name=cpu, type=INT32"`
	Type      string `parquet:"name=op, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	Pc        string `parquet:"name=pc, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	Elevel    int32  `parquet:"name=el, type=INT32"`
	BrCond    bool   `parquet:"name=condition, type=BOOLEAN"`
	BrInd     bool   `parquet:"name=indirect, type=BOOLEAN"`
	Event     string `parquet:"name=event, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	IssueLat  int32  `parquet:"name=issue_lat, type=INT32"`
	TotalLat  int32  `parquet:"name=total_lat, type=INT32"`
	Tgt       string `parquet:"name=br_tgt, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN"`
	TgtEl     int32  `parquet:"name=br_tgt_lvl, type=INT32"`
	Timestamp int64  `parquet:"name=ts, type=INT64"`
}

type MemoryBranchRecords struct {
	data []*Branch
}

func (rec *MemoryBranchRecords) Write(filename string) error {
	log.Println("Generating parquet file: " + filename)
	log.Printf("%d Branch records to write\n", len(rec.data))
	fw, err := local.NewLocalFileWriter(filename)

	if err != nil {
		return err
	}

	pw, err := writer.NewParquetWriter(fw, new(Branch), WriterConcurrency)
	if err != nil {
		return err
	}

	pw.CompressionType = parquet.CompressionCodec_ZSTD

	defer fw.Close()

	for i, d := range rec.data {
		if err = pw.Write(d); err != nil {
			return err
		}
		if i%100000 == 0 {
			if err = pw.Flush(true); err != nil {
				return err
			}
			log.Printf("%d Branch records have been flushed\n", i)
		}
	}

	if err = pw.WriteStop(); err != nil {
		return err
	}

	return nil
}
