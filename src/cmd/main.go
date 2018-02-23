package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/kofan/goblockchain/src/gochain"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	log.WithFields(log.Fields{
		"hostname": hostname,
		"pid":      pid,
	}).Info("Blockchain demo has been started")

	blockchain := gochain.NewBlockchain()
	setup(&blockchain)

	blockchain.PushCoinbase("Nickolay", 100)
	blockchain.PushCoinbase("Anna", 100)
	process(&blockchain)

	blockchain.PushTransaction("Anna", "Nickolay", 50)
	process(&blockchain)

	blockchain.PushTransaction("Nickolay", "Anna", 10)
	blockchain.PushTransaction("Nickolay", "Anna", 200)

	fmt.Printf("%v", &blockchain)
}

func setup(bc *gochain.Blockchain) {
	fmt.Printf("Enter the blockchain difficulty: __\b\b")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		difficulty, err := strconv.ParseInt(scanner.Text(), 10, 32)
		if err != nil {
			fmt.Printf("You've entered invalid number. Try again... __\b\b")
			continue
		}
		err = bc.SetDifficulty(int(difficulty))
		if err != nil {
			fmt.Printf("Error: %v. Try again... __\b\b", err)
			continue
		}
		break
	}
}

func process(bc *gochain.Blockchain) {
	duration, err := bc.ProcessPendingTransactions()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Mining time spent: %.3fs", duration.Seconds())
}
