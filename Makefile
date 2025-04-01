DATA_DIR ?= "$(HOME)/.local/state/protofire-game"

ifneq (,$(wildcard .env))
    include .env
    export
endif

build:
	@ go build -ldflags="-s -w -X 'main.dataDir=$(DATA_DIR)'" -o protofire-game cmd/main.go

run:
	@ go run -ldflags="-X 'main.dataDir=data'" cmd/main.go

test/client:
	@ go test -v ./...

test/contract:
	@ cd contract && forge test --fork-url $(NODE_RPC) -vvvv --match-contract ProtofireGame

generate/abi:
	@ cd contract && forge inspect ProtofireGame abi --json > ../internal/repository/abi/protofire-game.json

run/anvil:
	@ NODE_RPC="http://localhost:8545" anvil --fork-url $(NODE_RPC) --port 8545 --block-time 1

stop/anvil:
	@ pkill anvil

deploy/contract/dev:
	@ cd contract && forge script script/protofire-game.s.sol:ProtofireGameScript --fork-url "http://localhost:8545" --broadcast --legacy -vvvv

deploy/contract:
	@ cd contract && forge script script/protofire-game.s.sol:ProtofireGameScript --rpc-url $(NODE_RPC) --broadcast --legacy -vvvv

