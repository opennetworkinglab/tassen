#!/bin/bash

source /p4pktgen/my-venv/bin/activate

p4pktgen --allow-uninitialized-reads --allow-unimplemented-primitives -d $1
