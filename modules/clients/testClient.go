package clients

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
)

func GetClient() *ethclient.Client {
	client, err := ethclient.Dial("https://ropsten.infura.io/v3/66cd8456047a4527af2703f9ebd26c0e")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("client created")
	return client
}
