package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/SeogyuGim/eth-testnet-smartcontract/constants"
	"github.com/SeogyuGim/eth-testnet-smartcontract/contract"
	"github.com/SeogyuGim/eth-testnet-smartcontract/modules/clients"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"net/http"
)

func main() {

	// get health status
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Internal server error"))
			log.Fatal(err)
		}
	})

	// get chain id
	http.HandleFunc("/chain", func(w http.ResponseWriter, r *http.Request) {
		client := clients.GetClient()
		defer client.Close()
		cid, _ := client.ChainID(context.Background())
		w.Write([]byte(cid.String()))
	})

	// call contract
	http.HandleFunc("/contract", func(w http.ResponseWriter, r *http.Request) {
		client := clients.GetClient()
		defer client.Close()
		address := common.HexToAddress(constants.ContractAddress)
		instance, _ := contract.NewMyContractCaller(address, client)
		if instance != nil {
			w.Write([]byte("success"))
		} else {
			w.Write([]byte("failed"))
		}
	})

	// call contract method name
	http.HandleFunc("/contract/name", func(w http.ResponseWriter, r *http.Request) {
		client := clients.GetClient()
		defer client.Close()
		address := common.HexToAddress(constants.ContractAddress)
		instance, _ := contract.NewMyContractCaller(address, client)
		name, err := instance.Name(nil)
		if err != nil {
			log.Fatal(err)
		}
		w.Write([]byte(name))
		w.WriteHeader(200)
	})

	// call contract method transfer
	http.HandleFunc("/contract/transfer", func(w http.ResponseWriter, r *http.Request) {
		client := clients.GetClient()
		defer client.Close()
		privateKey, err := crypto.HexToECDSA(constants.PrivateKey)
		if err != nil {
			log.Fatal(err)
		}

		toAddr := common.HexToAddress(r.URL.Query().Get("to_address"))

		amount := new(big.Int)
		chainId, _ := client.ChainID(context.Background())

		cont, err := contract.NewMyContract(common.HexToAddress(constants.ContractAddress), client)
		if err != nil {
			log.Fatal(err)
		}

		opts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
		if err != nil {
			log.Fatal(err)
		}

		transfer, err := cont.Transfer(opts, toAddr, amount)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(transfer.Hash().Bytes())
	})

	// call raw transaction
	http.HandleFunc("/transfer", func(w http.ResponseWriter, r *http.Request) {
		client := clients.GetClient()
		defer client.Close()

		privateKey, err := crypto.HexToECDSA(constants.PrivateKey)
		if err != nil {
			log.Fatal(err)
		}

		pubKey := privateKey.Public()
		pubKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
		if !ok {
			log.Fatal("Error occurred while casting public key to ECDSA")
		}

		fromAddr := crypto.PubkeyToAddress(*pubKeyECDSA)
		nonce, err := client.PendingNonceAt(context.Background(), fromAddr)
		if err != nil {
			log.Fatal(err)
		}

		balance, _ := client.BalanceAt(context.Background(), fromAddr, nil)
		fmt.Printf("FromAddress balance: %s", balance.String())

		value := big.NewInt(0)
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		toAddr := common.HexToAddress(r.URL.Query().Get("to_address"))
		tokenAddr := common.HexToAddress(constants.ContractAddress)

		transferFnSig := []byte("transfer(address,uint256)")
		hash := crypto.NewKeccakState()
		hash.Write(transferFnSig)
		methodID := hash.Sum(nil)[:4]
		fmt.Printf("Method ID: %s\n", hexutil.Encode(methodID))

		paddedAddr := common.LeftPadBytes(toAddr.Bytes(), 32)

		amount := new(big.Int)
		amount.SetString("10", 10)
		paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

		var data []byte
		data = append(data, methodID...)
		data = append(data, paddedAddr...)
		data = append(data, paddedAmount...)

		gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
			To:   &toAddr,
			Data: data,
		})
		if err != nil {
			log.Fatal(err)
		}

		chainId, err := client.ChainID(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		tx := types.NewTransaction(nonce, tokenAddr, value, gasLimit, gasPrice, data)
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
		if err != nil {
			log.Fatal(err)
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Fatal(err)
		}

		w.Write(signedTx.Hash().Bytes())
	})

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("Server is not started")
	} else {
		fmt.Println("Server is listening on 8080 port ...")
	}
}
