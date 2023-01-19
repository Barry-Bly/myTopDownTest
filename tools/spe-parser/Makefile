# SPDX-License-Identifier: Apache-2.0
# Copyright (C) Arm Ltd. 2022

build:
	CGO_ENABLED=0 GOARCH=arm64 go build -o spe.arm64
	CGO_ENABLED=0 GOARCH=amd64 go build -o spe.amd64

clean:
	rm -rf spe.arm64 spe.amd64 *.parquet
all: build

