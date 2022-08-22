GO      := go
GOBUILD := $(GO) build
GOTEST  := $(GO) test
GOMODINIT	:= $(GO) mod init
MODBUILDER := ${GOPATH}/src/github.com/lyesteven/common/modMakefile/modbuilder.sh
PROCNAME := $(notdir $(shell pwd))

.PHONY: build initmod cleanmod clearall newmod depend initbase

default: build

initmod: newmod depend

cleanmod:
	@rm -f go.mod go.sum

build: depend
	$(GOBUILD)

clean:
	rm -f $(PROCNAME)

initbase:
	@$(MODBUILDER) init

clearall:
	@$(MODBUILDER) cleanmod

newmod:
ifeq (, $(wildcard go.mod))
	@$(GOMODINIT)
endif

depend:
	@$(MODBUILDER) depends
