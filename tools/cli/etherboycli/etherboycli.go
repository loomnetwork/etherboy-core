package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/loadimpact/k6/cmd"
	"github.com/loadimpact/k6/core"
	"github.com/loadimpact/k6/core/local"
	"github.com/loadimpact/k6/lib"
	"github.com/loadimpact/k6/stats"
	"github.com/loadimpact/k6/stats/dummy"
	"github.com/loadimpact/k6/ui"
	"github.com/loomnetwork/etherboy-core/txmsg"
	loom "github.com/loomnetwork/go-loom"
	"github.com/loomnetwork/go-loom/auth"
	"github.com/loomnetwork/go-loom/client"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
	null "gopkg.in/guregu/null.v3"
)

var writeURI, readURI, chainID string

func main() {
	var contractHexAddr, contractName string
	var privFile, user string
	var value int
	//var value int
	var iterations, maxuid, concurrency int64

	rootCmd := &cobra.Command{
		Use:   "etherboycli",
		Short: "Etherboy cli tool",
	}
	rootCmd.PersistentFlags().StringVarP(&writeURI, "write", "w", "http://localhost:46658", "URI for sending txs")
	rootCmd.PersistentFlags().StringVarP(&readURI, "read", "r", "http://localhost:9999", "URI for quering app state")
	rootCmd.PersistentFlags().StringVarP(&contractHexAddr, "contract", "", "0x005B17864f3adbF53b1384F2E6f2120c6652F779", "contract address")
	rootCmd.PersistentFlags().StringVarP(&contractName, "name", "n", "etherboycore", "smart contract name")
	rootCmd.PersistentFlags().StringVarP(&chainID, "chain", "", "default", "chain ID")

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
			contract, err := getContract(contractHexAddr, contractName)
			if err != nil {
				return err
			}
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
			contract, err := getContract(contractHexAddr, contractName)
			if err != nil {
				return err
			}
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
			contract, err := getContract(contractHexAddr, contractName)
			if err != nil {
				return err
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

	loadtestCreateCmd := &cobra.Command{
		Use:           "loadtest-create",
		Short:         "loadtesting with create requests",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(ccmd *cobra.Command, args []string) error {
			// setup RPC
			contract, err := getContract(contractHexAddr, contractName)
			if err != nil {
				return err
			}
			privKey, err := getPrivKey(privFile)
			if err != nil {
				return err
			}

			// metrics
			createAcctRTT := stats.New("Create Account RTT", stats.Trend)
			createAcctSuccessRate := stats.New("Create Account Success Rate", stats.Rate)
			// threshold
			thresholds := make(map[string]stats.Thresholds)
			ths, err := stats.NewThresholds([]string{"rate==1"})
			if err != nil {
				return err
			}
			thresholds[createAcctSuccessRate.Name] = ths

			// create runner
			runner := &LoomRunner{
				MaxUID: maxuid,
				Fn: func(ctx context.Context, uid string) (samples []stats.SampleContainer, err error) {
					defer func(begin time.Time) {
						samples = append(samples, stats.Sample{
							Metric: createAcctRTT,
							Time:   begin,
							Value:  time.Now().Sub(begin).Seconds(),
						})

						var succRate float64 = 1
						if err != nil {
							succRate = 0
						}
						samples = append(samples, stats.Sample{
							Metric: createAcctSuccessRate,
							Time:   begin,
							Value:  succRate,
						})
					}(time.Now())

					msg := &txmsg.EtherboyCreateAccountTx{
						Version: 0,
						Owner:   uid,
						Data:    []byte(uid),
					}
					signer := auth.NewEd25519Signer(privKey)
					_, err = contract.Call("CreateAccount", msg, signer, nil)
					if err != nil {
						return
					}
					return
				},
			}

			// executor
			var ex lib.Executor = local.New(runner)

			// config
			var conf cmd.Config
			conf.Iterations = null.IntFrom(iterations)
			conf.VUs = null.IntFrom(1) // VU must be 1 for create-acct
			conf.VUsMax = null.IntFrom(1)
			conf.Thresholds = thresholds

			engine, err := core.NewEngine(ex, conf.Options)
			if err != nil {
				log.Fatal(err)
			}

			// ignore collector
			engine.Collector = &dummy.Collector{}

			ctx, cancel := context.WithCancel(context.Background())
			errC := make(chan error)
			go func() { errC <- engine.Run(ctx) }()

			// Trap Interrupts, SIGINTs and SIGTERMs.
			sigC := make(chan os.Signal, 1)
			signal.Notify(sigC, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			defer signal.Stop(sigC)

			// run the engine
			func() {
				for {
					select {
					case err := <-errC:
						if err != nil {
							fmt.Printf("error: %v\n", err)
						}
						cancel()
						return
					case sig := <-sigC:
						fmt.Printf("got signal %v, exitting the loadtest...\n", sig)
						cancel()
						return
					}
				}
			}()

			fmt.Fprintf(os.Stdout, "\n")
			ui.Summarize(os.Stdout, "", ui.SummaryData{
				Opts:    conf.Options,
				Root:    engine.Executor.GetRunner().GetDefaultGroup(),
				Metrics: engine.Metrics,
				Time:    engine.Executor.GetTime(),
			})
			fmt.Fprintf(os.Stdout, "\n")

			if engine.IsTainted() {
				fmt.Fprintf(os.Stdout, "some thresholds have failed")
				os.Exit(1)
			}

			return nil
		},
	}
	loadtestCreateCmd.Flags().Int64VarP(&iterations, "iteration", "i", 1, "The number of iteration")
	loadtestCreateCmd.Flags().Int64VarP(&maxuid, "maxuid", "m", 1000, "The upperbound of possible UID")
	loadtestCreateCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")

	loadtestSetCmd := &cobra.Command{
		Use:           "loadtest-set",
		Short:         "loadteting with set requests",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(ccmd *cobra.Command, args []string) error {
			// setup RPC
			contract, err := getContract(contractHexAddr, contractName)
			if err != nil {
				return err
			}
			privKey, err := getPrivKey(privFile)
			if err != nil {
				log.Fatal(err)
			}

			// metrics
			setRTT := stats.New("Set RTT", stats.Trend)
			setSuccessRate := stats.New("Create Account Success Rate", stats.Rate)
			// threshold
			thresholds := make(map[string]stats.Thresholds)
			ths, err := stats.NewThresholds([]string{"rate==1"})
			if err != nil {
				return err
			}
			thresholds[setSuccessRate.Name] = ths

			// create runner
			runner := &LoomRunner{
				MaxUID: maxuid,
				Fn: func(ctx context.Context, uid string) (samples []stats.SampleContainer, err error) {
					defer func(begin time.Time) {
						samples = append(samples, stats.Sample{
							Metric: setRTT,
							Time:   begin,
							Value:  time.Now().Sub(begin).Seconds(),
						})

						var succRate float64 = 1
						if err != nil {
							succRate = 0
						}
						samples = append(samples, stats.Sample{
							Metric: setSuccessRate,
							Time:   begin,
							Value:  succRate,
						})
					}(time.Now())

					msgData := struct {
						Value string
					}{Value: uid}
					msgJSON, err := json.Marshal(msgData)
					if err != nil {
						return nil, err
					}
					msg := &txmsg.EtherboyStateTx{
						Version: 0,
						Owner:   uid,
						Data:    msgJSON,
					}
					signer := auth.NewEd25519Signer(privKey)
					_, err = contract.Call("SaveState", msg, signer, nil)
					if err != nil {
						return nil, err
					}
					return
				},
			}

			// executor
			var ex lib.Executor = local.New(runner)

			// config
			var conf cmd.Config
			conf.Iterations = null.IntFrom(iterations)
			conf.VUs = null.IntFrom(1)
			conf.VUsMax = null.IntFrom(1)
			conf.Thresholds = thresholds

			engine, err := core.NewEngine(ex, conf.Options)
			if err != nil {
				log.Fatal(err)
			}

			// ignore collector
			engine.Collector = &dummy.Collector{}

			ctx, cancel := context.WithCancel(context.Background())
			errC := make(chan error)
			go func() { errC <- engine.Run(ctx) }()

			// Trap Interrupts, SIGINTs and SIGTERMs.
			sigC := make(chan os.Signal, 1)
			signal.Notify(sigC, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			defer signal.Stop(sigC)

			// run the engine
			func() {
				for {
					select {
					case err := <-errC:
						if err != nil {
							fmt.Printf("error: %v\n", err)
						}
						cancel()
						return
					case sig := <-sigC:
						fmt.Printf("got signal %v, exitting the loadtest...\n", sig)
						cancel()
						return
					}
				}
			}()

			fmt.Fprintf(os.Stdout, "\n")
			ui.Summarize(os.Stdout, "", ui.SummaryData{
				Opts:    conf.Options,
				Root:    engine.Executor.GetRunner().GetDefaultGroup(),
				Metrics: engine.Metrics,
				Time:    engine.Executor.GetTime(),
			})
			fmt.Fprintf(os.Stdout, "\n")

			if engine.IsTainted() {
				fmt.Fprintf(os.Stdout, "some thresholds have failed")
				os.Exit(1)
			}

			return nil
		},
	}
	loadtestSetCmd.Flags().Int64VarP(&iterations, "iteration", "i", 1, "The number of iteration")
	loadtestSetCmd.Flags().Int64VarP(&maxuid, "maxuid", "m", 1000, "The upperbound of possible UID")
	loadtestSetCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")

	loadtestGetCmd := &cobra.Command{
		Use:           "loadtest-get",
		Short:         "loadteting with get requests",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(ccmd *cobra.Command, args []string) error {
			// setup RPC
			contract, err := getContract(contractHexAddr, contractName)
			if err != nil {
				return err
			}

			// metrics
			getRTT := stats.New("Get RTT", stats.Trend)
			getSuccessRate := stats.New("Get Success Rate", stats.Rate)
			// threshold
			thresholds := make(map[string]stats.Thresholds)
			ths, err := stats.NewThresholds([]string{"rate>0.99"})
			if err != nil {
				return err
			}
			thresholds[getSuccessRate.Name] = ths

			// create runner
			runner := &LoomRunner{
				MaxUID: maxuid,
				Fn: func(ctx context.Context, uid string) (samples []stats.SampleContainer, err error) {
					begin := time.Now()

					var result txmsg.StateQueryResult
					params := &txmsg.StateQueryParams{
						Owner: uid,
					}
					_, err = contract.StaticCall("GetState", params, &result)
					samples = append(samples, stats.Sample{
						Metric: getRTT,
						Time:   begin,
						Value:  time.Now().Sub(begin).Seconds(),
					})

					msgData := struct {
						Value string
					}{}
					json.Unmarshal(result.GetState(), &msgData)
					// fmt.Printf("---> user:%v, %v\n", uid, string(result.GetState()))

					var succRate float64 = 1
					if err != nil {
						succRate = 0
					}
					if msgData.Value != uid {
						succRate = 0
					}
					samples = append(samples, stats.Sample{
						Metric: getSuccessRate,
						Time:   begin,
						Value:  succRate,
					})

					return
				},
			}

			// executor
			var ex lib.Executor = local.New(runner)

			// config
			var conf cmd.Config
			conf.Iterations = null.IntFrom(iterations)
			conf.VUs = null.IntFrom(concurrency)
			conf.VUsMax = null.IntFrom(concurrency)
			conf.Thresholds = thresholds

			engine, err := core.NewEngine(ex, conf.Options)
			if err != nil {
				log.Fatal(err)
			}

			// ignore collector
			engine.Collector = &dummy.Collector{}

			ctx, cancel := context.WithCancel(context.Background())
			errC := make(chan error)
			go func() { errC <- engine.Run(ctx) }()

			// Trap Interrupts, SIGINTs and SIGTERMs.
			sigC := make(chan os.Signal, 1)
			signal.Notify(sigC, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			defer signal.Stop(sigC)

			// run the engine
			func() {
				for {
					select {
					case err := <-errC:
						if err != nil {
							fmt.Printf("error: %v\n", err)
						}
						cancel()
						return
					case sig := <-sigC:
						fmt.Printf("got signal %v, exitting the loadtest...\n", sig)
						cancel()
						return
					}
				}
			}()

			fmt.Fprintf(os.Stdout, "\n")
			ui.Summarize(os.Stdout, "", ui.SummaryData{
				Opts:    conf.Options,
				Root:    engine.Executor.GetRunner().GetDefaultGroup(),
				Metrics: engine.Metrics,
				Time:    engine.Executor.GetTime(),
			})
			fmt.Fprintf(os.Stdout, "\n")

			if engine.IsTainted() {
				fmt.Fprintf(os.Stdout, "some thresholds have failed")
				os.Exit(1)
			}

			return nil
		},
	}
	loadtestGetCmd.Flags().Int64VarP(&iterations, "iteration", "i", 1, "The number of iteration")
	loadtestGetCmd.Flags().Int64VarP(&concurrency, "concurrency", "c", 1, "The number of concurrency")
	loadtestGetCmd.Flags().Int64VarP(&maxuid, "maxuid", "m", 1000, "The upperbound of possible UID")
	loadtestGetCmd.Flags().StringVarP(&privFile, "key", "k", "", "private key file")

	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(createAccCmd)
	rootCmd.AddCommand(setStateCmd)
	rootCmd.AddCommand(getStateCmd)
	rootCmd.AddCommand(loadtestCreateCmd)
	rootCmd.AddCommand(loadtestSetCmd)
	rootCmd.AddCommand(loadtestGetCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func getPrivKey(privKeyFile string) ([]byte, error) {
	return ioutil.ReadFile(privKeyFile)
}

func getContract(contractHexAddr, contractName string) (*client.Contract, error) {
	rpcClient := client.NewDAppChainRPCClient(chainID, writeURI, readURI)
	contractAddr, err := loom.LocalAddressFromHexString(contractHexAddr)
	if err != nil {
		return nil, err
	}
	return client.NewContract(rpcClient, contractAddr, contractName), nil
}
