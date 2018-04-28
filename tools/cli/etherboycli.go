package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/loomnetwork/etherboy-core/txmsg"
	loom "github.com/loomnetwork/go-loom"
	"github.com/loomnetwork/go-loom/plugin"
	"github.com/loomnetwork/go-loom/types"
	"github.com/loomnetwork/loomchain/client"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

func decodeHexString(s string) ([]byte, error) {
	if !strings.HasPrefix(s, "0x") {
		return nil, errors.New("string has no hex prefix")
	}

	return hex.DecodeString(s[2:])
}

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

			msg := &txmsg.EtherboyCreateAccountTx{
				Version: 0,
				Owner:   "aditya",
				Data:    []byte("dummy"),
			}
			msgBytes, err := proto.Marshal(msg)
			if err != nil {
				return err
			}
			contractTx := &types.ContractMethodCall{
				Method: "etherboycore.CreateAccount",
				Args:   msgBytes,
			}
			contractTxBytes, err := proto.Marshal(contractTx)
			if err != nil {
				return err
			}
			req := &plugin.Request{
				ContentType: plugin.EncodingType_PROTOBUF3,
				Body:        contractTxBytes,
			}
			reqBytes, err := proto.Marshal(req)
			if err != nil {
				return err
			}
			contractAddrS, err := decodeHexString("0x005B17864f3adbF53b1384F2E6f2120c6652F779")
			if err != nil {
				return err
			}
			contractAddr := loom.Address{
				ChainID: "default",
				Local:   loom.LocalAddress(contractAddrS),
			}

			localAddr := loom.LocalAddressFromPublicKey(addr)
			log.Println(localAddr)
			clientAddr := loom.Address{
				ChainID: "default",
				Local:   localAddr,
			}
			signer := loom.NewEd25519Signer(privKey)
			rpcclient := client.NewDAppChainRPCClient("tcp://localhost", 46657, 9999)
			resp, err := rpcclient.CommitCallTx(clientAddr, contractAddr, signer, loom.VMType_PLUGIN, reqBytes)
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
			}{Val: value}
			msgJson, err := json.Marshal(msgData)
			if err != nil {
				log.Fatal("Cannot generate state json")
			}
			msg := &txmsg.EtherboyStateTx{
				Version: 0,
				Owner:   "aditya",
				Data:    msgJson,
			}
			msgBytes, err := proto.Marshal(msg)
			if err != nil {
				return err
			}
			contractTx := &types.ContractMethodCall{
				Method: "etherboycore.SaveState",
				Args:   msgBytes,
			}
			contractTxBytes, err := proto.Marshal(contractTx)
			if err != nil {
				return err
			}
			req := &plugin.Request{
				ContentType: plugin.EncodingType_PROTOBUF3,
				Body:        contractTxBytes,
			}
			reqBytes, err := proto.Marshal(req)
			if err != nil {
				return err
			}
			contractAddrS, err := decodeHexString("0x005B17864f3adbF53b1384F2E6f2120c6652F779")
			if err != nil {
				return err
			}
			contractAddr := loom.Address{
				ChainID: "default",
				Local:   loom.LocalAddress(contractAddrS),
			}
			localAddr := loom.LocalAddressFromPublicKey(addr)
			clientAddr := loom.Address{
				ChainID: "default",
				Local:   localAddr,
			}
			signer := loom.NewEd25519Signer(privKey)
			rpcclient := client.NewDAppChainRPCClient("tcp://localhost", 46657, 9999)
			resp, err := rpcclient.CommitCallTx(clientAddr, contractAddr, signer, loom.VMType_PLUGIN, reqBytes)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(resp)

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
