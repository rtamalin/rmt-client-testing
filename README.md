# A simulated client testing framework for SUSE/rmt

This repo is expected to be used in conjunction with the
[SUSE/rmt](github.com/SUSE/rmt) docker-compose based development
environment.

## Need the .env from a SUSE/rmt repo

Some of the Makefile rules also expect that an appropriately configured
.env file from an active development deployment RMT has been copied to
this repository clone's top-level directory and is available as the
.env file in this repo, with an additional REG_CODE entry added that
specifies the registration code to use when registering clients.

# Getting Started

Ensure that you have setup an RMT development environment using the
[SUSE/rmt](github.com/SUSE/rmt) and that you have created an appropriate
.env file as outlined above.

You can run `make rmt-setup` to ensure that the RMT is setup to support
registering clients with your configured registration code.

## Simulating client registrations

You can run `make client-register` to register a number of simulated
clients, controlled by the NUM_CLIENTS Makefile variable, defaulting
to 1000.

You can override the number of clients by specifying the desired value
on the make command line, e.g. `make NUM_CLIENTS=100 client-register`.

If the required hardware info to simulate the specified number of
clients doesn't exist, it will be automatically generated using the
`rmt-hwinfo-generator` tool. This data can also be manually generated
using the `make NUM_CLIENTS generated-hwinfo`. The `HwInfoStats.json`
file in the top-level directory of the generated client datastore
provides details about the generated set of simulated clients.

## Simulating client keepalive heartbeat updates

You can simulate clients sending keepalive heartbeat updates using 
the `client-update` Makefile target, again with the number of clients
being controlled via the `NUM_CLIENTS` Makefile variable.

To simulate triggering keepalive heartbeat updates for 100 clients,
you can run `make NUM_CLIENTS=100 client-update`. You will need to
have registered those simulated clients first though.

## Simulating client deregistration.

You can simulate clients deregistering using the `client-deregister`
Makefile target, again with the number of clients being controlled
via the `NUM_CLIENTS` Makefile variable.

To deregister 100 simulated clients that were previously registered
you can run `make NUM_CLIENTS-100 client-deregister`.

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

Also generates a `HwInfoStats.json` file in the top-level directory of
the specified data store directory that summarizes the generated clients
and the potential sizes and savings associated with the proposed data
profile storage handling scheme.

## rmt-hwinfo-clientctl

This tool can be used to simulate the registration, update (keepalive
heartbeat), and deregistration actions of clients with an RMT using the
provided hardware system information JSON blobs to register those clients.

# Helper Scripts

The `bin/` directory contains some helper scripts for querying the
RMT DB and working with the local developer deployment of RMT.

These scripts also require than a copy of an appropriately configured
SUSE/rmt .env file be available as the .env file in this repo.

## rmt-cli

Simple wrapper script to call rmt-cli command with provided arguments
within the locally running RMT Server containers, which defaults to
`rmt-rmt-1`.

## rmt-db-query

Simple wrapper script to exec the mariadb command in the RMT Server
container that is expected to be running locally, defaulting to
`rmt-rmt-1` per the SUSE/rmt docker compose deployment defaults.
It can take as argument an SQL query that will be executed, otherwise
it drops the user into an interactive mariadb session, connected to
the RMT DB.

## rmt-db-get-table-size

A helper script that leverages `rmt-db-query` to query the sizes of
tables in the RMT's DB.

## rmt-systems-table-size

A wrapper script for the `rmt-db-get-table-size` helper script that
calls it with the name of appropriate table name.

## rmt-system_data_profiles-table-size

A wrapper script for the `rmt-db-get-table-size` helper script that
calls it with the name of appropriate table name.

## rmt-size-of-system-information

A wrapper script for the rmt-db-query helper that calls it with an
SQL query that calculates the sum of the sizes of the `systems`
table's `system_information` column entries.
