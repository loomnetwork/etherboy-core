package main

import (
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
	"log"
	"io/ioutil"
	"github.com/loomnetwork/etherboy-core/gen"
	"github.com/loomnetwork/loom/client"
	lp "github.com/loomnetwork/loom-plugin"
	"github.com/gogo/protobuf/proto"
	"encoding/json"
	"fmt"
)

func main() {
	var publicFile string
	var privFile string
	var value int
	//var value int
	rootCmd := &cobra.Command{
		Use:   "etherboy",
		Short: "Etherboy cli tool",
	}
	createAccCmd := &cobra.Command{
		Use:   "create-acct",
		Short: "send a transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := ioutil.ReadFile(privFile)
			if err != nil {
				log.Fatalf("Cannot read priv key: %s", privFile)
			}
			addr, err := ioutil.ReadFile(publicFile)
			if err != nil {
				log.Fatalf("Cannot read address file: %s", publicFile)
			}
			msg := &txmsg.EtherboyAppTx{
				Version: 0,
				Type: "CreateAccount",
				Owner: "aditya",
				Data: &txmsg.EtherboyAppTx_CreateAccount{
					CreateAccount: &txmsg.EtherboyCreateAccountTx{
						Data: []byte("dummy"),
					},
				},
			}
			msgBytes, err := proto.Marshal(msg)
			signer := lp.NewEd25519Signer(privKey)
			rpcclient := client.NewDAppChainRPCClient("localhost:46657")
			resp, err := rpcclient.CommitCallTx(addr, []byte("etherboycore"), signer, lp.VMType_PLUGIN, msgBytes)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(resp)

			return nil
		},
	}
	createAccCmd.Flags().StringVarP(&publicFile, "address", "a", "", "address file")
	createAccCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")

	setStateCmd := &cobra.Command{
		Use:   "set",
		Short: "set the state",
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := ioutil.ReadFile(privFile)
			if err != nil {
				log.Fatalf("Cannot read priv key: %s", privFile)
			}
			addr, err := ioutil.ReadFile(publicFile)
			if err != nil {
				log.Fatalf("Cannot read address file: %s", publicFile)
			}
			log.Printf("running send with %d", value)
			msgData := struct {
				Val int
			}{ Val: value}
			msgJson, err := json.Marshal(msgData)
			if err != nil {
				log.Fatal("Cannot generate state json")
			}
			msg := &txmsg.EtherboyAppTx{
				Version: 0,
				Type: "SetState",
				Owner: "aditya",
				Data: &txmsg.EtherboyAppTx_State{
					State: &txmsg.EtherboyStateTx{
						Data: []byte(msgJson),
					},

				},
			}
			msgBytes, err := proto.Marshal(msg)
			signer := lp.NewEd25519Signer(privKey)
			rpcclient := client.NewDAppChainRPCClient("localhost:46657")
			rpcclient.CommitCallTx(addr, []byte("etherboycore"), signer, lp.VMType_PLUGIN, msgBytes)

			return nil
		},
	}
	setStateCmd.Flags().StringVarP(&publicFile, "address", "a", "", "address file")
	setStateCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")
	setStateCmd.Flags().IntVarP(&value, "value", "v", 0, "integer state value")

	keygenCmd := &cobra.Command{
		Use:   "genkey",
		Short: "generate a public and private key pair",
		RunE: func(cmd *cobra.Command, args []string) error {

			pub, priv, err := ed25519.GenerateKey(nil)
			if err != nil {
				log.Fatalf("Error generating key pair: %v", err)
			}
			if err := ioutil.WriteFile(publicFile, pub, 0664); err != nil {
				log.Fatalf("Unable to write public key: %v", err)
			}
			if err := ioutil.WriteFile(privFile, priv, 0664); err != nil {
				log.Fatalf("Unable to write private key: %v", err)
			}
			return nil
		},
	}
	keygenCmd.Flags().StringVarP(&publicFile, "address", "a", "", "address file")
	keygenCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")
	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(createAccCmd)
	rootCmd.AddCommand(setStateCmd)
	rootCmd.Execute()
}
