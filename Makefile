# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

GOPEERS=''
GOPARAMETERS=''

ifeq (,$(subst ,,$(PEERS)))
	GOPEERS=''
else
	GOPARAMETERS := $(GOPARAMETERS) '-peers='$(PEERS)
endif


ifeq (,$(subst ,,$(PORT)))
	GOPARAMETERS := $(GOPARAMETERS) '-port=3001'
else
	GOPARAMETERS := $(GOPARAMETERS) '-port='$(PORT)
endif



ifeq (,$(subst ,,$(ENV)))
	GOPARAMETERS := $(GOPARAMETERS) '-env=dev'
else
	GOPARAMETERS := $(GOPARAMETERS) '-env='$(ENV)
endif

ifeq (,$(subst ,,$(SELFIP)))
	GOPARAMETERS := $(GOPARAMETERS) '-selfip=127.0.0.1'
else
	GOPARAMETERS := $(GOPARAMETERS) '-selfip='$(SELFIP)
endif

install:
	$(GOGET) ./...

delete-db-dirs:
	@ rm -R ./herdius

create_db_dirs:
	@ mkdir -p mkdir ./herdius/chaindb/ ./herdius/statedb/ ./herdius/syncdb/ ./herdius/blockdb/

build:
	$(GOBUILD) ./...

build-node:
	$(GOBUILD) -o ./node ./cmd/validator

run-test:
	@$(GOTEST) -v ./...

all: install run-test create_db_dirs

start-validator:
	@echo "Starting validator node"
	@./node$(GOPARAMETERS)

start-node: build-node start-validator
