# protofire-game

Assumptions:

- The max length of player's name is 15 characters since I set in the smart contract as a name 15 bytes to be able to store each game result in a single slot.
- To fetch the results from the contract I used event logs which is better because makes less rpc requests, when there are just few transactions fetching directly the contract is faster but since there is no multicall contract deployed in the testnet I decided to move forward using event logs.
- For prod I store the local db in "$HOME/.local/state/protofire-game" since storing data in /.local/state/ is an standard but can be changed.
- At the beginning, it is possible to choose between storing the results in SQLite or Onchain.
- For games stored in SQLite, the id is a UUID, and Onchain is the tx hash.

Issues:

- The contract works and can be tested and used in a local fork, but I couldn't deploy the contract on the Harmony testnet because the RPC available doesn't support the RPC method (error -32000) used by forge to deploy contracts. I tried to search for other RPC URLs, but there are none. Anyways, I will give the step-by-step instructions on how to do it.

Requirements:

- Golang 1.24.1
- Forge 1.0.0

Environment variables:

- `NODE_RPC`: the RPC endpoint used to deploy the contract
- `RPC_ENDPOINT`: the RPC endpoint used by the client to interact with the on-chain contract.
- `PRIVATE_KEY`: private key of your address used to deploy the contract.
- `CONTRACT_ADDRESS`: contract address used by the client to store the games.
- `SIGNER`: private key of your address used as a signer in the client

How to run it locally:

1. Type `cp .env.example .env`
2. Run in an isolated terminal `make anvil`, which will create a local fork of the network set up in `NODE_RPC` env var and copy 1 private key and add it to the `.env` file in `SIGNER`
3. Run `make deploy/contract/dev` to deploy the contract to anvil, copy the contract address and set it in `CONTRACT_ADDRESS` env var.
4. Run `make run` to run the client locally.

How to deploy for prod:

1. In your `.env`, set `NODE_RPC` and `PRIVATE_KEY` to deploy the contract.
2. Run `make deploy/contract` to deploy the contract and copy the address and paste it in `CONTRACT_ADDRESS`.
3. Set `SIGNER` and run `make build`, it will build a protofire-game binary.
