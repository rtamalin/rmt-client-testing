# A simulated client testing framework for SUSE/rmt

This repo is expected to be used in conjunction with either:
* a [SUSE/rmt](github.com/SUSE/rmt) docker-compose based development
environment.
* a deployed RMT instance that is SSH accessible

## Need the .env from a SUSE/rmt repo

Some of the Makefile rules also expect that an appropriately configured
.env file from an active development deployment RMT has been copied to
this repository clone's top-level directory and is available as the
.env file in this repo, with an additional REG_CODE entry added that
specifies the registration code to use when registering clients.

# Getting Started

The first requirement is to ensure that an appropriate `.env` has been
generated and is available in the top-level directory of the cloned repo.

If using a local docker based RMT then you need to ensure that a valid
REG_CODE value is specified in the `.env` file, either by specifying it
to the helper script when generating the env file, or by manually adding
it later.

If targetting a PubCloud RMT then you need to ensure that a file that
contains a valid instance data XML document exists, and it's path must
be specified as the INST_DATA setting in the `.env` file. This can be
achieved by specifying the path as the second argument when calling the
helper script to generate the env file, or by manually adding it later.

Next the RMT should be setup appropriately to mirror the appropriate
product, defaulting to SLES/15.7/x86_64, which will be used by the
simulated client registrations.

Simulating client registration requires an appropriate set of simulated
client hardware info.

## Generating the appropriate .env file

The following helper scripts are available in the bin directory to
assist with generating a viable `.env` file:

* `bin/setup-docker-env`
  - Generates a `.env` file targetting a local docker compose based RMT
    dev env deployment if called with the path to the deployment's
    `.env` file.
  - An optional second argument can be used to specifying a client
    registration code that will be stored as the REG_CODE setting in
    the generated `.env` file.

* `bin/setup-pcrmt-env`
  - Generates a `.env` file targetting the specified, SSH accessible,
    RMT instance, when called with the appropriate `<user>@<host>` values.
  - An optional second argument specifying the path to a file containing
    an appropriate client instance data XML document which will be used
    instead of a REG_CODE when registering client systems against a
    PubCloud RMT. This path will be available as the INST_DATA setting
    in the the generated `.env` file.
  - NOTE: SSH access must already have been configured for the current
    user.

## Ensure the RMT is setup appropriately

You can run `make rmt-setup` to ensure that the RMT is setup to support
registering clients with your configured registration code.

Note that this is only required if the RMT hasn't already been setup
appropriately, and can take a long time if the RMT hasn't been setup
yet or hasn't mirrored updates recently.

## Generate simulated client hardware info

You can run `make generate-hwinfo` to generate the simulated client
hardware info using the `rmt-hwinfo-generator` tool. By default this
will generate info for 1000 simulated clients, resulting in the creation
of a `_ClientDataStore-1000` hierarchy hosting the simulated client
hardware info.

To generate simulated client hardware info for a different number client
systems, an override value can be specified for the `NUM_CLIENTS` variable
when running the make commane, e.g. `make NUM_CLIENTS=100 client-register`,
which will generate a `_ClientDataStore-100` hierarchy.

### Hardware Info Stats Details

When a client datastore hierarchy is generated a `HwInfoStats.json` file
will be created in the top-level directory that provides details about
the set of simulated clients.

## Simulating client registrations

You can run `make client-register` to register a number of simulated
clients, controlled by the NUM_CLIENTS Makefile variable, defaulting
to 1000.

You can override the number of clients by specifying the desired value
on the make command line, e.g. `make NUM_CLIENTS=100 client-register`.

If the simulated hardware info to support the registering the specified
number of clients doesn't exist yet, it will be automatically generated
using the `rmt-hwinfo-generator` tool via a dependency on the associated
`generate-hwinfo` target.

## Simulating client keepalive heartbeat updates

Note that it is only possible to simulate client keepalive heartbeat
updates if the clients have previously been registered appropriately
using `make client-register`.

You can simulate clients sending keepalive heartbeat updates using 
the `client-update` Makefile target, again with the number of clients
being controlled via the `NUM_CLIENTS` Makefile variable, which
defaults to 1000.

You can override the number of clients by specifying the desired value
on the make command line, e.g. `make NUM_CLIENTS=100 client-update`.

## Simulating client deregistration

Note that it is only possible to simulate client deregistration if
the clients have previously been registered appropriately using
`make client-register`.

You can simulate clients deregistering using the `client-deregister`
Makefile target, again with the number of clients being controlled
via the `NUM_CLIENTS` Makefile variable, which defaults to 1000.

You can override the number of clients by specifying the desired value
on the make command line, e.g. `make NUM_CLIENTS=100 client-deregister`.

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
