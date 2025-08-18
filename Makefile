.DEFAULT_GOAL := build
CNTR_MGR = docker
REPO_BASE_DIR := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

-include Makefile.docker
include Makefile.golang
