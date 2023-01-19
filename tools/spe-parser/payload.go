// SPDX-License-Identifier: Apache-2.0
// Copyright (C) Arm Ltd. 2022

package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type RecordType uint8

const (
	RecLoad RecordType = iota
	RecStore
	RecBranch
	RecUnknown
)

type RecordPayload struct {
	Type RecordType
	Data map[string][]string
	// each auxtrace record maps to spe packet. Each spe record has several packets.
	// The data is an array of map. The key is the packet type, and value is the
	// array of tokens. For example, event packet map: ['RETIRED', 'L1D-ACCESS', 'TLB-ACCESS']
}

type Payload interface {
	GetType() RecordType
	UpdateType()
	ToString() string
	ToLoadStore() LoadStore
}

func (r RecordPayload) GetType() RecordType {
	for k := range r.Data {
		switch k {
		case "LD":
			return RecLoad
		case "ST":
			return RecStore
		case "B":
			return RecBranch
		}
	}
	return RecUnknown
}

func (r *RecordPayload) UpdateType() {
	r.Type = r.GetType()
}

func (r RecordPayload) ToString() string {
	var x interface{} = r
	return fmt.Sprintf("%v", x)
}

func translateDataSource(values []string) (string, error) {
	if len(values) > 1 {
		return "", errors.New("invalid data source packet: " + strings.Join(values, ","))
	}
	ds := values[0]
	switch ds {
	case "0":
		return "L1D", nil
	case "8":
		return "L2D", nil
	case "9":
		return "PEER-CPU", nil
	case "10":
		return "LOCAL-CLUSTER", nil
	case "11":
		return "LL-CACHE", nil
	case "12":
		return "PEER-CLUSTER", nil
	case "13":
		return "REMOTE", nil
	case "14":
		return "DRAM", nil
	default:
		return "", errors.New("invalid data source value: " + ds)
	}
}

func (r *RecordPayload) ToLoadStore(cpu int32) (*LoadStore, error) {
	var err error
	if r.GetType() != RecLoad && r.GetType() != RecStore {
		return nil, errors.New("not a Load/Store instruction")
	}
	ldst := LoadStore{}
	ldst.Cpu = cpu

	for k, v := range r.Data {
		switch k {
		case "DATA-SOURCE":
			ldst.Source, err = translateDataSource(v)
			if err != nil {
				return &ldst, err
			}
		case "EV":
			// event
			ldst.Event = strings.Join(v, ":")
		case "ISSUE":
			ilat, err := strconv.ParseUint(v[0], 10, 64)
			if err != nil {
				return &ldst, err
			}
			ldst.IssueLat = int32(ilat)
		case "TOT":
			tlat, err := strconv.ParseUint(v[0], 10, 64)
			if err != nil {
				return &ldst, err
			}
			ldst.TotalLat = int32(tlat)
		case "XLAT":
			xlat, err := strconv.ParseUint(v[0], 10, 64)
			if err != nil {
				return &ldst, err
			}
			ldst.XlatLat = int32(xlat)
		case "PA":
			ldst.Paddr = v[0] // phys address
		case "PC":
			// v : 0xffffab47fdb0 el0 ns=1
			ldst.Pc = v[0]
			if err != nil {
				return &ldst, err
			}
			el, err := strconv.ParseUint(v[1][2:], 10, 64)
			if err != nil {
				return &ldst, err
			}
			ldst.Elevel = int32(el)
		case "TS":
			ts, err := strconv.ParseUint(v[0], 10, 64)
			if err != nil {
				return &ldst, err
			}
			ldst.Timestamp = int64(ts)

		case "VA":
			ldst.Vaddr = v[0]
		case "ST":
			fallthrough
		case "LD":
			// tools/perf/util/arm-spe-decoder/arm-spe-pkt-decoder.c
			ldst.Type = k
			str := strings.Join(v, " ")
			// defaul to GP-REG load
			ldst.Acqrel = false
			ldst.Atomic = false
			ldst.Exclusive = false
			ldst.Subclass = "GP-REG"
			if strings.Contains(str, "AT") {
				ldst.Atomic = true
				ldst.Subclass = ""
			}
			if strings.Contains(str, "EXCL") {
				ldst.Exclusive = true
				ldst.Subclass = ""
			}
			if strings.Contains(str, "AR") {
				ldst.Acqrel = true
				ldst.Subclass = ""
			}
			if !ldst.Atomic && !ldst.Acqrel && !ldst.Exclusive {
				ldst.Subclass = v[0]
			}
		default:
			return &ldst, errors.New("invalid spe packet for Load/store: " + strings.Join(v, " "))
		}
	}
	if ldst.Elevel == 2 {
		// The PC and Vaddr are missing the 0xff from the highest bits
		ldst.Pc = ldst.Pc[:2] + "ff" + ldst.Pc[2:]
		ldst.Vaddr = ldst.Vaddr[:2] + "ff" + ldst.Vaddr[2:]
	}
	return &ldst, nil
}

func (r *RecordPayload) ToBranch(cpu int32) (*Branch, error) {
	if r.GetType() != RecBranch {
		return nil, errors.New("not a Branch instruction")
	}
	br := Branch{}
	br.Cpu = cpu

	for k, v := range r.Data {
		switch k {
		case "B":
			br.Type = k
			switch l := len(v); l >= 0 {
			case l == 0:
				br.BrCond = false
				br.BrInd = false
			case l == 1:
				br.Type = k
				if v[0] == "COND" {
					br.BrCond = true
					br.BrInd = false
				} else if v[0] == "IND" {
					br.BrCond = false
					br.BrInd = true
				} else {
					return &br, errors.New("invalid br ops: " + string(v[0]))
				}
			default:
				return &br, errors.New("invalid other br ops: " + strings.Join(v, " "))

			}
		case "EV":
			// event
			br.Event = strings.Join(v, ":")
		case "ISSUE":
			ilat, err := strconv.ParseUint(v[0], 10, 64)
			if err != nil {
				return &br, err
			}
			br.IssueLat = int32(ilat)
		case "TOT":
			tlat, err := strconv.ParseUint(v[0], 10, 64)
			if err != nil {
				return &br, err
			}
			br.TotalLat = int32(tlat)
		case "PC":
			// v : 0xffffab47fdb0 el0 ns=1
			br.Pc = v[0]
			brel, err := strconv.ParseUint(v[1][2:], 10, 64)
			if err != nil {
				return &br, err
			}
			br.Elevel = int32(brel)
		case "TS":
			ts, err := strconv.ParseUint(v[0], 10, 64)
			if err != nil {
				return &br, err
			}
			br.Timestamp = int64(ts)
		case "TGT":
			br.Tgt = v[0]
			tgtel, err := strconv.ParseUint(v[1][2:], 10, 64)
			if err != nil {
				return &br, err
			}
			br.TgtEl = int32(tgtel)
		default:
			return &br, errors.New("invalid spe packet for branch: " + k + " " + strings.Join(v, " "))
		}
	}
	if br.Elevel == 2 {
		br.Pc = br.Pc[:2] + "ff" + br.Pc[2:]
	}
	if br.TgtEl == 2 {
		br.Tgt = br.Tgt[:2] + "ff" + br.Tgt[2:]
	}
	return &br, nil

}
