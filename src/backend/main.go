package main

import (
	"encoding/json"
	"math/big"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/actor"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/gas"
	"github.com/nspcc-dev/neo-go/pkg/wallet"

	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"backend/wrappers/solutiontoken"
	"backend/wrappers/tasktoken"
)

var (
	solutionContract solutiontoken.Invoker
)

func main() {

	ctx, _ := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM)
	rpcCli, err := rpcclient.New(ctx, "http://localhost:30333", rpcclient.Options{})

	if err != nil {
		die(err)
	}

	Listen(rpcCli)

}

func Listen(rpc *rpcclient.Client) error {

	http.HandleFunc("/createaccount", func(w http.ResponseWriter, r *http.Request) {
		f, err := requestToDict(r)
		smt := *f
		if err != nil {
			http_die(w, "json parsing err", err)
			return
		}

		name, ok := smt["name"].(string)
		if !ok {
			http_die(w, "name parsing err", err)
			return
		}
		passphrase, ok := smt["password"].(string)
		if !ok {
			http_die(w, "password parsing err", err)
			return
		}

		wal := wallet.NewInMemoryWallet()

		acc, err := wallet.NewAccount()
		if err != nil {
			http_die(w, "New account", err)
			return
		}
		acc.Label = name
		if err := acc.Encrypt(passphrase, wal.Scrypt); err != nil {
			http_die(w, "encrypt", err)
			return
		}
		wal.AddAccount(acc)
		fmt.Printf("account=%s\npassword=%s\n", name, passphrase)
		if err != nil {
			http_die(w, "CreateAccount", err)
			return
		}
		json, err := wal.JSON()
		if err != nil {
			http_die(w, "wallet json", err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "%s\n\n", json)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {

		dict, err := requestToDict(r)
		if err != nil {
			http_die(w, "requestToDict", err)
			return
		}
		_, _, err = checkAuthentication(dict)
		if err != nil {
			http_die(w, "checkAuthentication", err)
		}
		w.WriteHeader(http.StatusAccepted)
	})

	http.HandleFunc("/add_task", func(w http.ResponseWriter, r *http.Request) {

		smt, err := requestToDict(r)
		dict := *smt
		if err != nil {
			http_die(w, "requestToDict", err)
			return
		}
		_, acc, err := checkAuthentication(&dict)

		if err != nil {
			http_die(w, "checkAuthentication", err)
			return
		}

		act, err := actor.NewSimple(rpc, &acc)
		if err != nil {
			http_die(w, "newSimpleActor", err)
			return
		}

		data := &struct {
			Name        string
			Tests       []byte
			Description string
		}{
			Name:        dict["name"].(string),
			Tests:       []byte(dict["tests"].(string)),
			Description: dict["description"].(string),
		}

		json_dt, err := json.Marshal(data)

		if err != nil {
			die(err)
			return
		}

		contractGas := gas.New(act)
		_, _, err = contractGas.Transfer(act.Sender(),
			tasktoken.Hash, big.NewInt(10_0000_0000), json_dt)

		if err != nil {
			http_die(w, "Transfer", err)
			die(err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Successful")

	})

	http.ListenAndServe("localhost:8040", nil)
	return nil
}

func requestToDict(r *http.Request) (*map[string]any, error) {

	smt := map[string]any{}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger("reading body", err)
		return nil, err
	}
	err = json.Unmarshal(bytes, &smt)
	if err != nil {
		logger("json unmarshall", err)
		return nil, err
	}

	return &smt, nil
}

func checkAuthentication(f *map[string]any) (wallet.Wallet, wallet.Account, error) {
	smt := *f
	wallet_dt, ok := smt["wallet"]
	if !ok {
		err := errors.New("missing field in json")
		logger("wallet", err)
		return wallet.Wallet{}, wallet.Account{}, err
	}
	wallet_dt, err := json.Marshal(wallet_dt)
	if err != nil {
		logger("Marshal", err)
		return wallet.Wallet{}, wallet.Account{}, err
	}

	password, ok := smt["password"].(string)
	if !ok {
		logger("passwordParsing", err)
		return wallet.Wallet{}, wallet.Account{}, err
	}
	wal, err := wallet.NewWalletFromBytes(wallet_dt.([]byte))

	if err != nil {
		logger("NewWalletFromBytes", err)
		return wallet.Wallet{}, wallet.Account{}, err
	}

	acc := wal.GetAccount(wal.GetChangeAddress())

	err = acc.Decrypt(password, wal.Scrypt)

	if err != nil {
		logger("Decrypt", err)
		return wallet.Wallet{}, wallet.Account{}, err
	}
	return *wal, *acc, nil
}

func http_die(writer http.ResponseWriter, error_type string, err error) {
	writer.WriteHeader(http.StatusInternalServerError)
	msg := fmt.Sprintf("%s err: \n", error_type)
	if err != nil {
		msg = fmt.Sprintf("%s err: %s\n", error_type, err.Error())
	}
	fmt.Fprintf(writer, "%s", msg)
	logger(msg, err)

}

func logger(msg string, err error) {

	println(msg)
}
func die(err error) {
	if err == nil {
		return
	}

	debug.PrintStack()
	_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
