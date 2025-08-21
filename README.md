# A simulated client testing framework for SUSE/rmt

This repo is expected to be used in conjunction with the
[SUSE/rmt](github.com/SUSE/rmt) that has at least one product mirrored
and available for clients to register against, which by default is
expected to be the base SLES/15.7/x86_64 product with 

## Need the .env from a SUSE/rmt repo
Some of the Makefile rules also expect that an appropriately configured
.env file from an active development deployment RMT is available as the
.env file in this repo, with a REG_CODE entry setup in the .env file.

# Tools available in this repo

The repo provides a number of tools for use with testing client
registrations with an RMT.

Running `make build` will build these tools, and make the available
under the `out/` directory by default, overridable via the `OUT_DIR`
make variable.

## rmt-hwinfo-generator

This tool can be used to generate hardware system information JSON
blobs for number of clients, which will be stored in the specified
directory hierarchy.

## rmt-hwinfo-clientctl

This tool can be used to simulate registration of clients with an
RMT using the provided hardware system information JSON blobs to
register those clients.

# Helper Scripts
The `bin/` directory contains some helper scripts for querying the
RMT DB.

These scripts also require than a copy of an appropriately configured
SUSE/rmt .env file be available as the .env file in this repo.

## rmt-db-query
Simple wrapper script to exec the mariadb command in the RMT Server
container that is expected to be running locally, defaulting to
`rmt-rmt-1` per the SUSE/rmt docker compose deployment defaults.
It can take as argument an SQL query that will be executed, otherwise
it drops the user into an interactive mariadb session, connected to
the RMT DB.

## rmt-systems-table-size
A wrapper script for the rmt-db-query helper that calls it with an

## rmt-size-of-system-information
A wrapper script for the rmt-db-query helper that calls it with an
SQL query that calculates the sum of the sizes of the `systems`
table's `system_information` column entries.