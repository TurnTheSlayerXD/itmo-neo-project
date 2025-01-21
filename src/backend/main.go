package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"

	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/actor"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/gas"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/nep17"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
)

import (
	solutiontoken "backend/wrappers/solutiontoken"
)

func main() {

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	rpcCli, err := rpcclient.New(ctx, "http://localhost:30333", rpcclient.Options{})
	die(err)
	w, err := wallet.NewWalletFromFile("../wallets/wallet1.json")
	die(err)
	acc := w.GetAccount(w.GetChangeAddress())
	err = acc.Decrypt("", w.Scrypt)
	die(err)

	act, err := actor.NewSimple(rpcCli, acc)
	die(err)
}

func die(err error) {
	if err == nil {
		return
	}

	debug.PrintStack()
	_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
