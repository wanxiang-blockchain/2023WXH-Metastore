export package=github.com/MetaDataLab/web3-console-backend/cmd
export TAG=v0.0.1
export COMMIT=$(shell git rev-parse --short HEAD)
export GOVER=$(shell go version|awk '{print $$3}')
export BLDTIME=$(shell date +%Y-%m-%d:%H:%M:%S)
export LDFLAGS=-X $(package).TAG=$(TAG) \
		-X '$(package).COMMIT=$(COMMIT)' \
		-X '$(package).GOVER=$(GOVER)' \
		-X '$(package).BLDTIME=$(BLDTIME)'

.PHONY: all clean

all: 
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/web3-console-backend ./

clean: 
	rm -rf ./bin/web3-console-backend
