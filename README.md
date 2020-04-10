# Tassen

This project aims at defining a BNG Control and User Plane Separation (CUPS) API
based on P4Runtime, gNMI, and OpenConfig.

Tassen is a German word for "cups".

## Requirements

To build and test the Tassen API you will need the following software to be
installed on your machine:

* Docker
* make

Docker is used to run the necessary without worrying about additional
dependencies. Before starting, make sure to fetch all the required Docker
images:

    make deps

## Content

### Reference BNG-UP P4 Implementation

The directory `p4src` contains the P4 program defining the reference packet
forwarding pipeline of a BNG user plane (BNG-UP) abstracted by the Tassen API.

The goal of this P4 program is twofold:

1. formally define the forwarding model of a BNG-UP;
2. implicitly define the runtime API that a BNG control plane (BNG-CP) can use
   to manipulate the forwarding state of the BNG-UP.

To build the P4 program:

    make build

### Packet-based Unit Tests

The directory `ptf` contains unit tests for the P4 program. Tests use PTF, a
Python-based framework for data plane testing, and `stratum_bmv2`, the reference
P4 software switch ([BMv2 simple_switch][bmv2]) built with [Stratum][stratum]
support to provide a P4Runtime and gNMI server interface.

To run all test cases:

    make check

`ptf/tests` contains the actual test case implementation, organized in logical
groups, e.g., `routing.py` for all test cases pertaining the routing
functionality, `packetio.py` for control packet I/O, etc.

To run all tests in a group:

    make check TEST=<GROUP>

To run a specific test case:

    make check TEST=<GROUP>.<TEST NAME>

For example:

    make check TEST=packetio.PacketOutTest
  
`ptf/lib` contains the test runner as well as libraries useful to simplify
the test case implementations (e.g., `helper.py` provides a P4Info helper with
methods convenient to construct P4Runtime table entries)

### Mapping to Target-specific BNG-UP implementations

The directory `mapping` is currently empty. The goal is to provide here a
reference implementation of the runtime mapping logic to translate P4Runtime
RPCs for the logical P4 program (`bng.p4`) to target-specific ones, e.g., for
Tofino, FPGA, Broadcom Q2C, etc.


[bmv2]: https://github.com/p4lang/behavioral-model
[stratum]: https://github.com/stratum/stratum
