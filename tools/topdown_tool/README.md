Topdown Tool
============

Tool to collect and compute metric data for Arm Neoverse CPUs, including metrics for Topdown performance analysis.

Data is collected via Linux `perf stat` with metric information stored in per-CPU JSON files.

Install
=======

This tool can optionally be installed as a python package by running the following from the project directory:

```
pip3 install .
```

or

```
pip3 install --user .
```

Usage
=====

When installed, a script will be created to execute the tool.

If pip's `<install>/bin` directory is in your PATH, you can execute the tool as follows:

```
topdown-tool --help
```

Alternatively, you can execute the `topdown-tool` script from the project directory:
```
./topdown-tool --help
```

Choosing what to monitor
------------------------

### Launch and monitor an application

```
topdown-tool ./a.out
```

### Monitor a running application

You can specify one or more process IDs to monitor:
```
$ topdown-tool -p 289156
Monitoring PID(s) 289156. Hit Ctrl-C to stop.
...
```
```
$ topdown-tool --pid 289156,289153
Monitoring PID(s) 289156,289153. Hit Ctrl-C to stop.
...
```

### System-wide monitoring

If no application or process ID is specified, then system-wide monitoring will be performed (for all CPUs)

```
$ topdown-tool
Starting system-wide profiling. Hit Ctrl-C to stop. (See --help for usage information.)
...
```

Choosing which metrics to measure
---------------------------------

### What is available?

The metrics (and metric groups) available will depend on the Arm CPU used.

Examples below were collected on a Neoverse N1 system.

To query the available metric groups:

```
$ topdown-tool --list-groups
Cycle_Accounting (Cycle Accounting)
General (General)
MPKI (Misses Per Kilo Instructions)
Miss_Ratio (Miss Ratio)
Branch_Effectiveness (Branch Effectiveness)
ITLB_Effectiveness (Instruction TLB Effectiveness)
DTLB_Effectiveness (Data TLB Effectiveness)
L1I_Cache_Effectiveness (L1 Instruction Cache Effectiveness)
L1D_Cache_Effectiveness (L1 Data Cache Effectiveness)
L2_Cache_Effectiveness (L2 Unified Cache Effectiveness)
LL_Cache_Effectiveness (Last Level Cache Effectiveness)
Operation_Mix (Speculative Operation Mix)
```

To query metrics according to the Arm Topdown Performance Analysis Methodology:

```
$ topdown-tool --list-metrics
Stage 1 (Topdown metrics)
=========================
[Cycle Accounting]
Frontend Stalled Cycles
Backend Stalled Cycles

Stage 2 (uarch metrics)
=======================
[Branch Effectiveness]
  (follows Frontend Stalled Cycles)
Branch Misprediction Ratio
Branch MPKI

[Data TLB Effectiveness]
  (follows Backend Stalled Cycles)
DTLB MPKI
DTLB Walk Ratio
L1 Data TLB Miss Ratio
L1 Data TLB MPKI
L2 Unified TLB Miss Ratio
L2 Unified TLB MPKI
...
```

### Topdown metrics

By default, metrics from the Arm topdown performance analysis methodology will be selected, and grouped by stage:

```
$ topdown-tool ./a.out
Stage 1 (Topdown metrics)
=========================
[Cycle Accounting]
Frontend Stalled Cycles............. 0.02% cycles
Backend Stalled Cycles.............. 42.59% cycles

Stage 2 (uarch metrics)
=======================
[Branch Effectiveness]
  (follows Frontend Stalled Cycles)
Branch Misprediction Ratio.......... 0.001 per branch
Branch MPKI......................... 0.372 misses per 1,000 instructions

[Data TLB Effectiveness]
  (follows Backend Stalled Cycles)
DTLB MPKI........................... 0.000 misses per 1,000 instructions
DTLB Walk Ratio..................... 0.000 per TLB access
L1 Data TLB Miss Ratio.............. 0.000 per TLB access
L1 Data TLB MPKI.................... 0.002 misses per 1,000 instructions
L2 Unified TLB Miss Ratio........... 0.000 per TLB access
L2 Unified TLB MPKI................. 0.006 misses per 1,000 instructions
...
```

A specific stage can also be specified by number:

```
$ topdown-tool -s 1 ./a.out
Stage 1 (Topdown metrics)
=========================
...
```

or by name:

```
$ topdown-tool -s uarch ./a.out
Stage 2 (uarch metrics)
=======================
...
```

These metrics can also be combined into a single hierarchy:

```
$ topdown-tool -s combined ./a.out
[Cycle Accounting]                     [Topdown group]
Frontend Stalled Cycles............... 0.00% cycles
  [Branch Effectiveness]               [uarch group]
  Branch MPKI......................... 0.371 misses per 1,000 instructions
  Branch Misprediction Ratio.......... 0.001 per branch

  [Instruction TLB Effectiveness]      [uarch group]
  ...
Backend Stalled Cycles................ 42.78% cycles
  [Data TLB Effectiveness]             [uarch group]
  DTLB MPKI........................... 0.000 misses per 1,000 instructions
  L1 Data TLB MPKI.................... 0.002 misses per 1,000 instructions
  L2 Unified TLB MPKI................. 0.000 misses per 1,000 instructions
  DTLB Walk Ratio..................... 0.000 per TLB access
  L1 Data TLB Miss Ratio.............. 0.000 per TLB access
  L2 Unified TLB Miss Ratio........... 0.002 per TLB access

  [L1 Data Cache Effectiveness]        [uarch group]
  ...
```

### Collecting metric groups

It is also possible to collect specific specific metric groups (as show in `topdown-tool --list-groups`):

```
$ topdown-tool --metric-group MPKI,Miss_Ratio ./a.out
[Misses Per Kilo Instructions] [uarch group]
Branch MPKI................... 0.399 misses per 1,000 instructions
ITLB MPKI..................... 0.000 misses per 1,000 instructions
L1 Instruction TLB MPKI....... 0.001 misses per 1,000 instructions
DTLB MPKI..................... 0.000 misses per 1,000 instructions
L1 Data TLB MPKI.............. 0.013 misses per 1,000 instructions
L2 Unified TLB MPKI........... 0.000 misses per 1,000 instructions
L1I Cache MPKI................ 0.001 misses per 1,000 instructions
L1D Cache MPKI................ 0.002 misses per 1,000 instructions
L2 Cache MPKI................. 0.000 misses per 1,000 instructions
LL Cache Read MPKI............ 0.000 misses per 1,000 instructions

[Miss Ratio]                   [uarch group]
Branch Misprediction Ratio.... 0.001 per branch
ITLB Walk Ratio............... 0.000 per TLB access
DTLB Walk Ratio............... 0.000 per TLB access
L1 Instruction TLB Miss Ratio. 0.000 per TLB access
L1 Data TLB Miss Ratio........ 0.000 per TLB access
L2 Unified TLB Miss Ratio..... 0.015 per TLB access
L1I Cache Miss Ratio.......... 0.000 per cache access
L1D Cache Miss Ratio.......... 0.000 per cache access
L2 Cache Miss Ratio........... 0.065 per cache access
LL Cache Read Miss Ratio...... 0.435 per cache access
```

Collecting a branch of the hierarchy
------------------------------------

It is also possible to collect a specific branch of the combined hierarchy:

```
$ topdown-tool --node backend_stalled_cycles ./a.out
[Cycle Accounting]                     [Topdown group]
Backend Stalled Cycles................ 42.42% cycles
  [Data TLB Effectiveness]             [uarch group]
  DTLB MPKI........................... 0.000 misses per 1,000 instructions
  L1 Data TLB MPKI.................... 0.002 misses per 1,000 instructions
  L2 Unified TLB MPKI................. 0.000 misses per 1,000 instructions
  DTLB Walk Ratio..................... 0.000 per TLB access
  L1 Data TLB Miss Ratio.............. 0.000 per TLB access
  L2 Unified TLB Miss Ratio........... 0.002 per TLB access

  [L1 Data Cache Effectiveness]        [uarch group]
  ...
```


Other options
-------------

See `topdown-tool --help` for full usage information.