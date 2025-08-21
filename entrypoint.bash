#!/bin/bash

cat - << __EOF__
RMT Client Tester
=================

The rmt-client-tester repo is mounted as /app/tester
The generated ClientDataStore is mounted as /app/ClientDataStore
The SUSE/connect-ng and rmt-client-tester built commands are available
in /app/bin.
__EOF__

exec bash -i
