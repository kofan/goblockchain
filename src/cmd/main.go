package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kofan/goblockchain/src/common/appflag"
	"github.com/kofan/goblockchain/src/gochain"

	log "github.com/sirupsen/logrus"
)

var dataDir string

var dataFile = flag.String("datafile", "blockchain.dat", "file where blockchain gets persisted")
var difficulty = appflag.Difficulty("difficulty", 0, "difficulty of the blockchain [0-255]")
var logLevel = appflag.LogLevel("loglevel", log.DebugLevel, "application log level")

func init() {
	dataDir = filepath.Join(os.Getenv("GOPATH"), "github.com/kofan/goblockchain", "data")
	flag.Parse()
	log.SetLevel(*logLevel)
}

func main() {
	hostname, _ := os.Hostname()
	pid := os.Getpid()

	log.WithFields(log.Fields{
		"hostname": hostname,
		"pid":      pid,
	}).Info("Blockchain demo has been started")

	stream, err := openFile(*dataFile)
	if err != nil {
		log.Fatalf(`Cannot open/create the file "%s"`, *dataFile)
	}
	blockchain := gochain.NewBlockchain(stream, *difficulty)
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

func openFile(path string) (*os.File, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(dataDir, path)
	}
	return os.OpenFile("path", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
}

func setup(bc *gochain.Blockchain) {
	fmt.Printf("Enter the blockchain difficulty: __\b\b")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		difficulty, err := strconv.ParseInt(scanner.Text(), 10, 8)
		if err != nil {
			fmt.Printf("You've entered invalid number. Try again... __\b\b")
			continue
		}
		err = bc.SetDifficulty(uint8(difficulty))
		if err != nil {
			fmt.Printf("Error: %v. Try again... __\b\b", err)
			continue
		}
		break
	}
}

func process(bc *gochain.Blockchain) {
	duration, err := bc.ProcessPendingTrasactions()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Mining time spent: %.3fs", duration.Seconds())
}
