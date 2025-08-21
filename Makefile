.DEFAULT_GOAL := build
CNTR_MGR = docker
REPO_BASE_DIR := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

# golang rules
include Makefile.golang

# optional rules that may not exist in all cases
-include Makefile.docker
-include Makefile.tester
