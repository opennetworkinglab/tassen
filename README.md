# Tassen: Next-Generation BNG CUPS API

[![CircleCI](https://circleci.com/gh/opennetworkinglab/tassen.svg?style=svg&circle-token=1192ef25b712aaf3f6e5e54fb65b3aad27ad1f57)](https://app.circleci.com/pipelines/github/opennetworkinglab/tassen)

This project aims at defining an API for BNG Control and User Plane Separation
(CUPS) based on next-generation SDN interfaces such as P4Runtime, gNMI, and
OpenConfig.

Tassen is a German word for "cups".

## Requirements

To build and test the Tassen API you will need the following software to be
installed on your machine:

* Docker
* make

Docker is used to run the necessary tools without worrying about additional
dependencies. Before starting, make sure to fetch all the required Docker
images:

    make deps

## Quick Instructions

To build everything:

    make build

To test everything:

    make check

## Content

### Reference BNG-UP P4 Implementation

The directory `p4src` contains `bng.p4`, the P4 program defining the reference
packet forwarding pipeline of a BNG user plane (BNG-UP) abstracted by the Tassen
API.

The goal of this P4 program is twofold:

1. formally define the forwarding model of a BNG-UP;
2. implicitly define the runtime API that a BNG control plane (BNG-CP) can use
   to manipulate the forwarding state of the BNG-UP.

To build the P4 program:

    make p4

To generate the P4 graphs:

    make graph

### Packet-based Unit Tests

The directory `ptf` contains unit tests for the P4 program. Tests use PTF, a
Python-based framework for data plane testing, and `stratum_bmv2`, the reference
P4 software switch ([BMv2 simple_switch][bmv2]) built with [Stratum][stratum]
support to provide a P4Runtime and gNMI server interface.

To run all test cases:

    make check

`ptf/tests` contains the actual test case implementation, organized in
modules, e.g., `upstream.py` for all test cases pertaining the upstream
functionality, `packetio.py` for control packet I/O, etc.

To run all tests in a module:

    make check TEST=<MODULE>

To run a specific test case:

    make check TEST=<MODULE>.<TEST NAME>

For example:

    make check TEST=packetio.PacketOutTest

To run all tests, except that of a specific module (e.g., `accounting`)

    make check TEST="all ^accounting"

`ptf/lib` contains the test runner as well as libraries useful to simplify
the test case implementations (e.g., `helper.py` provides a P4Info helper with
methods convenient to construct P4Runtime table entries).

### Mapping to Target-specific BNG-UP Implementations

The directory `mapr` contains the reference implementation of the runtime
logic to translate P4Runtime RPCs for the logical P4 program (`bng.p4`) to 
target-specific ones, e.g., for Tofino, FPGA, Broadcom Q2C, etc.

`mapr` is written in Go.

To build `mapr`:

    make mapr

This command will produce a binary in `mapr/mapr` that can be used as part of
the PTF tests.

`mapr` currently provides the translation logic for different targets, such as:

* `dummy`: for testing purposes only, where the target device runs with
  the same Tassen logical pipeline, and P4Runtime RPCs are relayed as-is,
  with no translation. 
* `fabric`: for a switch running ONF's fabric.p4 (`fabric-bng` profile).

To run PTF tests on a given target together with `mapr`:

    make check-<target> TEST=<filters>

For example, to run all PTF tests on the `dummy` target:

    make check-dummy TEST=all

## Continuous Integration

This repository is configured with a CircleCI job that checks code changes by:

* building logical P4 program (`bng.p4`) for BMv2;
* running tests on different targets.

The job configuration can be found in [.circleci/config.yml](.circleci/config.yml).

Currently, we check the following targets:

| Target   | Tests             | Notes                                       |
|--------- |-------------------|---------------------------------------------|
| `self`   | `all`             | Tests executed without `mapr`               |
| `dummy`  | `all`             |                                             |
| `fabric` | `all ^accounting` | Accounting not supported on fabric.p4, yet. |

[bmv2]: https://github.com/p4lang/behavioral-model
[stratum]: https://github.com/stratum/stratum
