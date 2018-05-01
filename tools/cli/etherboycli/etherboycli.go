package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/loomnetwork/etherboy-core/txmsg"
	loom "github.com/loomnetwork/go-loom"
	"github.com/loomnetwork/go-loom/auth"
	"github.com/loomnetwork/go-loom/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

var writeURI = fmt.Sprintf("http://%s:%d", "localhost", 46657)
var readURI = fmt.Sprintf("http://%s:%d", "localhost", 9999)

func getKeys(privFile, publicFile string) ([]byte, []byte, error) {
	privKey, err := getPrivKey(privFile)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Cannot read priv key: %s", privFile)
	}
	addr, err := ioutil.ReadFile(publicFile)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Cannot read public key: %s", publicFile)
	}
	return privKey, addr, nil
}

func getPrivKey(privKeyFile string) ([]byte, error) {
	return ioutil.ReadFile(privKeyFile)
}

func main() {
	var privFile string
	var value int
	//var value int

	rpcClient := client.NewDAppChainRPCClient("default", writeURI, readURI)

	contractAddr, err := loom.LocalAddressFromHexString("0x005B17864f3adbF53b1384F2E6f2120c6652F779")
	if err != nil {
		log.Fatalf("Cannot generate contract address: %v", err)
	}
	contract := client.NewContract(rpcClient, contractAddr, "etherboycore")

	createAccCmd := &cobra.Command{
		Use:   "create-acct",
		Short: "send a transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := getPrivKey(privFile)
			if err != nil {
				log.Fatal(err)
			}
			msg := &txmsg.EtherboyCreateAccountTx{
				Version: 0,
				Owner:   "aditya",
				Data:    []byte("dummy"),
			}
			signer := auth.NewEd25519Signer(privKey)
			resp, err := contract.Call("CreateAccount", msg, signer, nil)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(resp)

			return nil
		},
	}
	createAccCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")

	setStateCmd := &cobra.Command{
		Use:   "set",
		Short: "set the state",
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := getPrivKey(privFile)
			if err != nil {
				log.Fatal(err)
			}

			msgData := struct {
				Value int
			}{Value: value}
			msgJSON, err := json.Marshal(msgData)
			if err != nil {
				log.Fatal("Cannot generate state json")
			}
			msg := &txmsg.EtherboyStateTx{
				Version: 0,
				Owner:   "aditya",
				Data:    msgJSON,
			}

			signer := auth.NewEd25519Signer(privKey)
			resp, err := contract.Call("SaveState", msg, signer, nil)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(resp)

			return nil
		},
	}
	setStateCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")
	setStateCmd.Flags().IntVarP(&value, "value", "v", 0, "integer state value")

	keygenCmd := &cobra.Command{
		Use:   "genkey",
		Short: "generate a public and private key pair",
		RunE: func(cmd *cobra.Command, args []string) error {

			_, priv, err := ed25519.GenerateKey(nil)
			if err != nil {
				log.Fatalf("Error generating key pair: %v", err)
			}
			if err := ioutil.WriteFile(privFile, priv, 0664); err != nil {
				log.Fatalf("Unable to write private key: %v", err)
			}
			return nil
		},
	}
	keygenCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")

	rootCmd := &cobra.Command{
		Use:   "etherboycli",
		Short: "Etherboy cli tool",
	}
	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(createAccCmd)
	rootCmd.AddCommand(setStateCmd)
	rootCmd.Execute()
}
