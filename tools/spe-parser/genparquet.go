// SPDX-License-Identifier: Apache-2.0
// Copyright (C) Arm Ltd. 2022

package main

type GenParquet interface {
	Write(filename string) error
}
