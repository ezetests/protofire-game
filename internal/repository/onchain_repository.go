package repository

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"bytes"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"protofire-game/internal/domain"
)

const (
	GasLimit          = 80000
	maxBlocksPerQuery = 1000
)

type OnChainRepository struct {
	client       *ethclient.Client
	abi          abi.ABI
	contractAddr common.Address
	signer       string
}

func NewOnChainRepository() (*OnChainRepository, error) {
	nodeRPC := os.Getenv("RPC_ENDPOINT")
	if nodeRPC == "" {
		return nil, fmt.Errorf("RPC_ENDPOINT environment variable is not set")
	}

	contractAddr := os.Getenv("CONTRACT_ADDRESS")
	if contractAddr == "" {
		return nil, fmt.Errorf("CONTRACT_ADDRESS environment variable is not set")
	}

	signer := os.Getenv("SIGNER")
	if signer == "" {
		return nil, fmt.Errorf("SIGNER environment variable is not set")
	}

	abiFile, err := os.ReadFile("internal/repository/abi/protofire-game.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read ABI file: %w", err)
	}

	contractABI, err := abi.JSON(bytes.NewReader(abiFile))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	client, err := ethclient.Dial(nodeRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}

	return &OnChainRepository{
		client:       client,
		abi:          contractABI,
		contractAddr: common.HexToAddress(contractAddr),
		signer:       signer,
	}, nil
}

func (r *OnChainRepository) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

func (r *OnChainRepository) GetGameResult(ctx context.Context, index uint64) (*domain.Game, error) {
	data, err := r.abi.Pack("getGameResult", big.NewInt(int64(index)))
	if err != nil {
		return nil, err
	}

	result, err := r.client.CallContract(ctx, ethereum.CallMsg{
		To:   &r.contractAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var (
		player1 [15]byte
		player2 [15]byte
		winner  uint8
	)

	err = r.abi.UnpackIntoInterface(&[]interface{}{&player1, &player2, &winner}, "getGameResult", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %w", err)
	}

	var winnerName string
	switch winner {
	case 0:
		winnerName = "Draw"
	case 1:
		winnerName = string(bytes.TrimRight(player1[:], "\x00"))
	case 2:
		winnerName = string(bytes.TrimRight(player2[:], "\x00"))
	}

	return &domain.Game{
		ID:       fmt.Sprintf("game_%d", index),
		Player1:  string(bytes.TrimRight(player1[:], "\x00")),
		Player2:  string(bytes.TrimRight(player2[:], "\x00")),
		Winner:   winnerName,
		PlayedAt: time.Now().Format(time.RFC3339Nano),
	}, nil
}

func (r *OnChainRepository) GetGameHistory() ([]*domain.Game, error) {
	ctx := context.Background()

	latestBlock, err := r.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	// Create a map to store block timestamps to avoid fetching the same block multiple times
	blockTimestamps := make(map[common.Hash]time.Time)

	var results []*domain.Game

	eventSig := r.abi.Events["GameResultStored"].ID

	// Find the first block with an event using binary search
	firstEventBlock := uint64(0)
	left := uint64(0)
	right := latestBlock

	for left <= right {
		mid := (left + right) / 2
		startBlock := mid
		if mid > maxBlocksPerQuery {
			startBlock = mid - maxBlocksPerQuery
		}

		query := ethereum.FilterQuery{
			Addresses: []common.Address{r.contractAddr},
			FromBlock: big.NewInt(int64(startBlock)),
			ToBlock:   big.NewInt(int64(mid)),
			Topics:    [][]common.Hash{{eventSig}},
		}

		logs, err := r.client.FilterLogs(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs: %w", err)
		}

		if len(logs) > 0 {
			firstEventBlock = logs[0].BlockNumber
			right = firstEventBlock - 1
		} else {
			left = mid + 1
		}
	}

	// Query logs in chunks of maxBlocksPerQuery blocks
	for fromBlock := firstEventBlock; fromBlock <= latestBlock; fromBlock += maxBlocksPerQuery {
		toBlock := fromBlock + maxBlocksPerQuery - 1
		if toBlock > latestBlock {
			toBlock = latestBlock
		}

		query := ethereum.FilterQuery{
			Addresses: []common.Address{r.contractAddr},
			FromBlock: big.NewInt(int64(fromBlock)),
			ToBlock:   big.NewInt(int64(toBlock)),
			Topics:    [][]common.Hash{{eventSig}},
		}

		logs, err := r.client.FilterLogs(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs from block %d to %d: %w", fromBlock, toBlock, err)
		}

		for _, log := range logs {
			event := struct {
				Winner uint8
			}{}

			err = r.abi.UnpackIntoInterface(&event, "GameResultStored", log.Data)
			if err != nil {
				continue
			}

			// Extract indexed parameters from topics
			// Topics[0] is the event signature
			// Topics[1] is player1 (indexed)
			// Topics[2] is player2 (indexed)
			if len(log.Topics) != 3 {
				continue
			}

			player1Bytes := log.Topics[1].Bytes()
			player2Bytes := log.Topics[2].Bytes()

			// Get time from the block timestamp where the event was emitted(from cache if available)
			playedAt, exists := blockTimestamps[log.BlockHash]
			if !exists {
				block, err := r.client.BlockByHash(ctx, log.BlockHash)
				if err != nil {
					return nil, fmt.Errorf("failed to get block: %w", err)
				}
				playedAt = time.Unix(int64(block.Time()), 0)
				blockTimestamps[log.BlockHash] = playedAt
			}

			result := &domain.Game{
				ID:       log.TxHash.Hex(),
				Player1:  string(bytes.TrimRight(player1Bytes[:15], "\x00")),
				Player2:  string(bytes.TrimRight(player2Bytes[:15], "\x00")),
				PlayedAt: playedAt.Format(time.RFC3339Nano),
			}

			switch event.Winner {
			case 0:
				result.Winner = "Draw"
			case 1:
				result.Winner = result.Player1
			case 2:
				result.Winner = result.Player2
			}

			results = append(results, result)
		}
	}

	return results, nil
}

func (r *OnChainRepository) SaveGame(result *domain.Game) error {
	ctx := context.Background()

	var player1Bytes [15]byte
	var player2Bytes [15]byte
	copy(player1Bytes[:], []byte(result.Player1))
	copy(player2Bytes[:], []byte(result.Player2))

	var winnerNum uint8
	switch result.Winner {
	case "Draw":
		winnerNum = 0
	case result.Player1:
		winnerNum = 1
	case result.Player2:
		winnerNum = 2
	default:
		return fmt.Errorf("invalid winner value")
	}

	return r.StoreGameResult(ctx, player1Bytes, player2Bytes, winnerNum)
}

func (r *OnChainRepository) StoreGameResult(ctx context.Context, player1, player2 [15]byte, winner uint8) error {
	data, err := r.abi.Pack("storeGameResult", player1, player2, winner)
	if err != nil {
		return fmt.Errorf("failed to pack data: %w", err)
	}

	chainID, err := r.client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	privateKeyECDSA, err := crypto.HexToECDSA(strings.TrimPrefix(r.signer, "0x"))
	if err != nil {
		return fmt.Errorf("failed to parse signer key: %w", err)
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := r.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := r.client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKeyECDSA, chainID)
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = GasLimit
	auth.GasPrice = gasPrice

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &r.contractAddr,
		Value:     big.NewInt(0),
		Gas:       GasLimit,
		GasFeeCap: gasPrice,
		GasTipCap: big.NewInt(1),
		Data:      data,
	})

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), privateKeyECDSA)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = r.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	_, err = bind.WaitMined(ctx, r.client, signedTx)
	if err != nil {
		return fmt.Errorf("failed to wait for transaction to be mined: %w", err)
	}

	return nil
}
