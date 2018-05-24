package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/loomnetwork/etherboy-core/txmsg"
	loom "github.com/loomnetwork/go-loom"
	"github.com/loomnetwork/go-loom/auth"
	ctypes "github.com/loomnetwork/go-loom/builtin/types/coin"
	"github.com/loomnetwork/go-loom/client"
	types "github.com/loomnetwork/go-loom/types"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

var writeURI = fmt.Sprintf("http://%s:%d/rpc", "localhost", 46658)
var readURI = fmt.Sprintf("http://%s:%d/query", "localhost", 46658)

// var writeURI = fmt.Sprintf("http://%s:%d/rpc", "etherboy-stage.loomapps.io", 80)
// var readURI = fmt.Sprintf("http://%s:%d/query", "etherboy-stage.loomapps.io", 80)

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

	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "send a transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := getPrivKey(privFile)
			if err != nil {
				log.Fatal(err)
			}
			msg := &txmsg.EtherboyEndGameTx{
				Version: 0,
				Owner:   user,
				Data:    []byte(user),
			}
			signer := auth.NewEd25519Signer(privKey)
			encoder := base64.StdEncoding
			addr := loom.LocalAddressFromPublicKey(signer.PublicKey()[:])
			fmt.Println(addr)
			fmt.Println(encoder.EncodeToString(addr))
			resp, err := contract.Call("EndGame", msg, signer, nil)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(resp)

			return nil
		},
	}
	txCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")
	txCmd.Flags().StringVarP(&user, "user", "u", "", "user name")

	balCmd := &cobra.Command{
		Use:   "bal",
		Short: "Balance",
		RunE: func(cmd *cobra.Command, args []string) error {
			encoder := base64.StdEncoding
			addr := loom.LocalAddressFromPublicKey([]byte("0xe288d6eec7150D6a22FDE33F0AA2d81E06591C4d"))
			fmt.Println("==")
			fmt.Println(addr)
			fmt.Println(encoder.EncodeToString([]byte("0xe288d6eec7150D6a22FDE33F0AA2d81E06591C4d")))
			fmt.Println("==")
			rpcClient := client.NewDAppChainRPCClient("default", writeURI, readURI)
			contractAddr, err := loom.LocalAddressFromHexString("0x01D10029c253fA02D76188b84b5846ab3D19510D")
			if err != nil {
				log.Fatalf("Cannot generate contract address: %v", err)
			}
			contract := client.NewContract(rpcClient, contractAddr)
			addr1 := loom.MustParseAddress("default:" + user)
			// "0xe9CF9552A580c7A79667f7F81D541Ecd6af2EBf9"

			msg := &ctypes.BalanceOfRequest{
				Owner: addr1.MarshalPB(),
			}
			var result ctypes.BalanceOfResponse
			if _, err := contract.StaticCall("BalanceOf", msg, &result); err != nil {
				return err
			}
			fmt.Println(result.Balance.Value)
			return nil
		},
	}
	balCmd.Flags().StringVarP(&user, "user", "u", "", "user name")

	trnTokenCmd := &cobra.Command{
		Use:   "transfer_token",
		Short: "send a transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, err := getPrivKey(privFile)
			if err != nil {
				log.Fatal(err)
			}
			toAddr := &types.Address{}
			msg := &txmsg.EtherboyTransferTokenTx{
				Version: 0,
				Owner:   user,
				ToAddr:  toAddr,
			}
			signer := auth.NewEd25519Signer(privKey)
			encoder := base64.StdEncoding
			addr := loom.LocalAddressFromPublicKey(signer.PublicKey()[:])
			fmt.Println(addr)
			fmt.Println(encoder.EncodeToString(addr))
			resp, err := contract.Call("TransferToken", msg, signer, nil)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(resp)

			return nil
		},
	}
	trnTokenCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")
	trnTokenCmd.Flags().StringVarP(&user, "user", "u", "", "user name")

	transferCmd := &cobra.Command{
		Use:   "transfer",
		Short: "Transfer to Contract",
		RunE: func(cmd *cobra.Command, args []string) error {
			rpcClient := client.NewDAppChainRPCClient("default", writeURI, readURI)
			contractAddr, err := loom.LocalAddressFromHexString("0x01D10029c253fA02D76188b84b5846ab3D19510D")
			if err != nil {
				log.Fatalf("Cannot generate contract address: %v", err)
			}
			contract := client.NewContract(rpcClient, contractAddr)
			privKey, err := getPrivKey(privFile)
			if err != nil {
				log.Fatal(err)
			}
			//Address of etherboy contract
			addr1 := loom.MustParseAddress("default:0xe288d6eec7150D6a22FDE33F0AA2d81E06591C4d")
			amount := loom.NewBigUIntFromInt(10)
			msg := &ctypes.TransferRequest{
				To:     addr1.MarshalPB(),
				Amount: &types.BigUInt{Value: *amount},
			}
			signer := auth.NewEd25519Signer(privKey)
			resp, err := contract.Call("Transfer", msg, signer, nil)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(resp)
			return nil
		},
	}
	transferCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")

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
			privKey, err := getPrivKey(privFile)
			if err != nil {
				log.Fatal(err)
			}

			var result txmsg.StateQueryResult
			params := &txmsg.StateQueryParams{
				Owner: user,
			}

			signer := auth.NewEd25519Signer(privKey)

			callerAddr := loom.Address{
				ChainID: rpcClient.GetChainID(),
				Local:   loom.LocalAddressFromPublicKey(signer.PublicKey()),
			}
			if _, err := contract.StaticCall("GetState", params, callerAddr, &result); err != nil {
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
	rootCmd.AddCommand(txCmd)
	rootCmd.AddCommand(balCmd)
	rootCmd.AddCommand(transferCmd)
	rootCmd.AddCommand(trnTokenCmd)
	rootCmd.Execute()
}
