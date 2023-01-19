# SPE Parser Tool

spe-parser is a tool used to parse SPE data from `perf` raw data that contains SPE records. It will save parsed SPE records into files with parquet format.

The `perf` raw data should be collected by `perf record` with SPE extension and converted into raw format with `perf report -D`.

# Usage

## Build

Before building the tool, you need to install `go` on your machine. And then simply run `make`.
```
make
```
The output file `spe.arm64` is the executable tool.

## Run

### Enable SPE
Before collecting SPE data, you need to enable SPE on your machine.
```
1. Enable CONFIG_ARM_SPE_PMU and CONFIG_PID_IN_CONTEXTIDR in kernel config
2. Add "kpti=off" option in kernel boot cmdline
3. If compiled as module, may need run "modprobe arm_spe_pmu"
4. Check if spe supported in perf: "perf list | grep arm_spe"
```

### Collect SPE data
Run `perf` to collect SPE data with root privilege, and convert the result into raw format for spe-parser.

Add `--all-kernel` or `--all-user` for specific exception level if you want.  Add `-c xxxx` if you want specify sampling interval (kernel 5.17+).
```
perf record -e arm_spe_0/branch_filter=1,ts_enable=1,pct_enable=1,pa_enable=1,load_filter=1,jitter=1,store_filter=1,min_latency=0/ -- test_program
```
Or, you can run perf against a running application like:
```
perf record -p 300526 -e arm_spe_0/branch_filter=1,ts_enable=1,pct_enable=1,pa_enable=1,load_filter=1,jitter=1,store_filter=1,min_latency=0/
```
Convert perf.data file into raw format.
```
perf report -D -i perf.data > perf.raw
```

### Run SPE parser
Run spe.arm64 with `-p`, which defines the prefix of output files' names.

```
./spe.arm64 -f ./perf.raw -p testspe
```
Output is `testspe-br.parquet` and `testspe-ldst.parquet` which is in parquet file format. The first one is all the SPE records for branch operation. And the second one is SPE records for load and store operations.

For post processing, you can use whatever tool that you are familiar with to process these parquet files.

### Usage info
Here is the usage info for spe-parser tool.
```
Usage: spe.arm64 [-bdl] [-c value] [-f value] [-p value] [parameters ...]
 -b, --nobr        false
 -c, --concurrency=value
                   Parquet writer concurrency
 -d, --debug       false
 -f, --file=value  perf spe trace file from 'perf script -D'
 -l, --noldst      false
 -p, --prefix=value
                   file prefix for output parquet file
```
