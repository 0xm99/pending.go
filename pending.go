package main

import (
        "context"
        "fmt"
        "log"
        "os"
        "os/signal"
        "reflect"
        "syscall"
        "time"
        "unsafe"

        "github.com/ethereum/go-ethereum/common"
        "github.com/ethereum/go-ethereum/ethclient"
        "github.com/ethereum/go-ethereum/rpc"
)

type TransactionStream struct {
        transactionHash   common.Hash
        isPending         bool
}

func main() {
        // Handle ctrl+c routine
        c := make(chan os.Signal)
        signal.Notify(c, os.Interrupt, syscall.SIGTERM)
        go handleCtrlC(c)

        // Connect to Ethereum client
        client, err := ethclient.Dial("/home/geth/mainnet/geth.ipc")
        if err != nil {
                log.Fatal("Error connecting to client: ", err)
        }
        fmt.Println("Successfully connected to Ethereum client")

        // Initialize RPC client
        rpcClient := initRPCClient(client)

        // Stream new transactions from the mempool
        streamNewTxs(client, rpcClient)
}

func handleCtrlC(c chan os.Signal) {
        startTime := time.Now()

        // Will be executed after Ctrl+C
        <-c

        endTime := time.Now()
        fmt.Println("\nStart: ", startTime.Format("2006-01-02 15:04:05.000000000"))
        fmt.Println("End:   ", endTime.Format("2006-01-02 15:04:05.000000000"))
        os.Exit(0)
}

func initRPCClient(client *ethclient.Client) *rpc.Client {
        clientValue := reflect.ValueOf(client).Elem()
        fieldStruct := clientValue.FieldByName("c")
        clientPointer := reflect.NewAt(fieldStruct.Type(), unsafe.Pointer(fieldStruct.UnsafeAddr())).Elem()
        finalClient, _ := clientPointer.Interface().(*rpc.Client)
        return finalClient
}

func streamNewTxs(client *ethclient.Client, rpcClient *rpc.Client) {
        newTxsChannel := make(chan common.Hash)

        _, err := rpcClient.EthSubscribe(
                context.Background(), newTxsChannel, "newPendingTransactions", // no additional args
        )
        if err != nil {
                fmt.Println("Error while subscribing: ", err)
                return
        }

        fmt.Println("\nSubscribed to mempool txs!\n")

        for {
                select {
                case transactionHash := <-newTxsChannel:
                        txStream := TransactionStream{
                                transactionHash: transactionHash,
                        }

                        go processTransaction(client, &txStream)
                }
        }
}

func processTransaction(client *ethclient.Client, txStream *TransactionStream) {
        transactionHash := txStream.transactionHash

        _, isPending, _ := client.TransactionByHash(context.Background(), transactionHash)
        currentTime := time.Now()

        if isPending {
                fmt.Println(currentTime.Format("2006-01-02 15:04:05.000000000"), transactionHash, "[PENDING]")
        } else {
                fmt.Println(currentTime.Format("2006-01-02 15:04:05.000000000"), transactionHash)
        }
}
