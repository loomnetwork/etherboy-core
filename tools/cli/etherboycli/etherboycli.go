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
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

var writeURI = fmt.Sprintf("http://%s:%d/rpc", "localhost", 46658)
var readURI = fmt.Sprintf("http://%s:%d/query", "localhost", 46658)

func getPrivKey(privKeyFile string) ([]byte, error) {
	return ioutil.ReadFile(privKeyFile)
}

func main() {
	var privFile, user string
	var value int
	//var value int

	rpcClient := client.NewDAppChainRPCClient("default", writeURI, readURI)

	contractAddr, err := loom.LocalAddressFromHexString("0xe288d6eec7150D6a22FDE33F0AA2d81E06591C4d")
	if err != nil {
		log.Fatalf("Cannot generate contract address: %v", err)
	}
	contract := client.NewContract(rpcClient, contractAddr)

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
				Owner:   user,
				Data:    []byte(user),
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
	createAccCmd.Flags().StringVarP(&user, "user", "u", "", "user name")

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
				Owner:   user,
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
	setStateCmd.Flags().StringVarP(&user, "user", "u", "", "user")

	getStateCmd := &cobra.Command{
		Use:   "get",
		Short: "get state",
		RunE: func(cmd *cobra.Command, args []string) error {
			var result txmsg.StateQueryResult
			params := &txmsg.StateQueryParams{
				Owner: user,
			}
			if _, err := contract.StaticCall("GetState", params, &result); err != nil {
				return err
			}
			fmt.Println(string(result.State))
			return nil
		},
	}

	getStateCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")
	getStateCmd.Flags().StringVarP(&user, "user", "u", "loom", "user")

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
	rootCmd.AddCommand(getStateCmd)
	rootCmd.Execute()
}
