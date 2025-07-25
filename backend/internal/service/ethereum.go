package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"paperplay/config"
	"paperplay/internal/model"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthereumService handles Ethereum blockchain operations
type EthereumService struct {
	config     *config.EthereumConfig
	client     *ethclient.Client
	privateKey *ecdsa.PrivateKey
	enabled    bool
}

// NewEthereumService creates a new Ethereum service
func NewEthereumService(config *config.EthereumConfig) (*EthereumService, error) {
	if !config.Enabled {
		return &EthereumService{
			config:  config,
			enabled: false,
		}, nil
	}

	// Connect to Ethereum network
	client, err := ethclient.Dial(config.NetworkURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum network: %w", err)
	}

	// Load private key if provided
	var privateKey *ecdsa.PrivateKey
	if config.PrivateKey != "" {
		privateKey, err = crypto.HexToECDSA(config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
	}

	return &EthereumService{
		config:     config,
		client:     client,
		privateKey: privateKey,
		enabled:    true,
	}, nil
}

// IsEnabled checks if Ethereum service is enabled
func (s *EthereumService) IsEnabled() bool {
	return s.enabled
}

// GenerateWallet generates a new Ethereum wallet
func (s *EthereumService) GenerateWallet() (address, privateKey string, err error) {
	if !s.enabled {
		return "", "", fmt.Errorf("Ethereum service is not enabled")
	}

	// Generate new private key
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Get address from private key
	addr := crypto.PubkeyToAddress(key.PublicKey)

	// Convert private key to hex string (without 0x prefix)
	privateKeyHex := fmt.Sprintf("%x", crypto.FromECDSA(key))

	return addr.Hex(), privateKeyHex, nil
}

// GetBalance returns the ETH balance for an address
func (s *EthereumService) GetBalance(address string) (*big.Int, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Ethereum service is not enabled")
	}

	addr := common.HexToAddress(address)
	balance, err := s.client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// NFTMintRequest represents an NFT minting request
type NFTMintRequest struct {
	ToAddress string         `json:"to_address"`
	TokenURI  string         `json:"token_uri"`
	Metadata  map[string]any `json:"metadata"`
}

// NFTMintResult represents the result of NFT minting
type NFTMintResult struct {
	TokenID         string `json:"token_id"`
	TransactionHash string `json:"transaction_hash"`
	ContractAddress string `json:"contract_address"`
}

// MintNFT mints an NFT for a user achievement
func (s *EthereumService) MintNFT(request NFTMintRequest) (*NFTMintResult, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Ethereum service is not enabled")
	}

	if s.privateKey == nil {
		return nil, fmt.Errorf("private key not configured for minting")
	}

	// For this implementation, we'll simulate NFT minting
	// In a real implementation, you would:
	// 1. Load the NFT contract ABI
	// 2. Create a contract instance
	// 3. Call the mint function
	// 4. Wait for transaction confirmation

	// Get the public address from private key
	publicKey := s.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get nonce for transaction
	nonce, err := s.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Create transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, big.NewInt(s.config.ChainID))
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // Transfer 0 ETH
	auth.GasLimit = s.config.GasLimit
	auth.GasPrice = gasPrice

	// For simulation, generate a mock token ID and transaction hash
	tokenID := generateMockTokenID()
	txHash := generateMockTxHash()

	// In a real implementation, you would call the contract's mint function here
	// Example:
	// contract, err := NewNFTContract(common.HexToAddress(s.config.ContractAddress), s.client)
	// tx, err := contract.Mint(auth, common.HexToAddress(request.ToAddress), tokenID, request.TokenURI)

	return &NFTMintResult{
		TokenID:         tokenID,
		TransactionHash: txHash,
		ContractAddress: s.config.ContractAddress,
	}, nil
}

// CreateNFTAsset creates an NFT asset record in the database
func (s *EthereumService) CreateNFTAsset(userID string, achievementID *string, metadataURI string) (*model.NFTAsset, error) {
	if !s.enabled {
		return nil, fmt.Errorf("Ethereum service is not enabled")
	}

	nftAsset := &model.NFTAsset{
		UserID:          userID,
		AchievementID:   achievementID,
		ContractAddress: s.config.ContractAddress,
		TokenID:         "", // Will be set after minting
		MetadataURI:     metadataURI,
		Status:          model.NFTStatusPending,
	}

	return nftAsset, nil
}

// UpdateNFTAsset updates an NFT asset with minting results
func (s *EthereumService) UpdateNFTAsset(nftAsset *model.NFTAsset, result *NFTMintResult) {
	nftAsset.TokenID = result.TokenID
	nftAsset.MintTxHash = result.TransactionHash
	nftAsset.MarkMinted(result.TransactionHash)
}

// GenerateMetadataURI generates a metadata URI for an NFT
func (s *EthereumService) GenerateMetadataURI(achievement *model.Achievement, userID string) (string, error) {
	// In a real implementation, you would upload metadata to IPFS or another storage service
	// and return the URI. For now, we'll generate a mock URI.

	mockURI := fmt.Sprintf("https://api.paperplay.com/nft/metadata/%s/%s", achievement.ID, userID)
	return mockURI, nil
}

// GetNetworkInfo returns information about the connected network
func (s *EthereumService) GetNetworkInfo() (map[string]any, error) {
	if !s.enabled {
		return map[string]any{
			"enabled": false,
		}, nil
	}

	chainID, err := s.client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get network ID: %w", err)
	}

	latestBlock, err := s.client.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	return map[string]any{
		"enabled":          true,
		"network_url":      s.config.NetworkURL,
		"chain_id":         chainID.String(),
		"contract_address": s.config.ContractAddress,
		"latest_block":     latestBlock,
		"gas_limit":        s.config.GasLimit,
		"gas_price":        s.config.GasPrice,
	}, nil
}

// HealthCheck checks the health of Ethereum service
func (s *EthereumService) HealthCheck() map[string]any {
	if !s.enabled {
		return map[string]any{
			"status":  "disabled",
			"enabled": false,
		}
	}

	// Check if we can connect to the network
	_, err := s.client.NetworkID(context.Background())
	if err != nil {
		return map[string]any{
			"status":  "unhealthy",
			"enabled": true,
			"error":   err.Error(),
		}
	}

	return map[string]any{
		"status":  "healthy",
		"enabled": true,
	}
}

// Close closes the Ethereum client connection
func (s *EthereumService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

// Helper functions for simulation

func generateMockTokenID() string {
	// In a real implementation, this would be generated by the smart contract
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func generateMockTxHash() string {
	// In a real implementation, this would be the actual transaction hash
	return fmt.Sprintf("0x%x", time.Now().UnixNano())
}
