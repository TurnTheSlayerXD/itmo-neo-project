package main

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"

	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/actor"
	"github.com/nspcc-dev/neo-go/pkg/rpcclient/gas"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
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

type NFTTask struct {
	ID            string
	Owner         util.Uint160
	Name          string
	Tests         string
	Description   string
	NSolutions    int
	AverAssesment int
}
type NFTSolution struct {
	ID            string
	TaskId        string
	TaskAssesment int
	Owner         util.Uint160
	SrcCode       string
	NAssesments   int
	AverAssesment int
	Description   string
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

		f, err := json.Marshal(dict["tests"])
		tests := string(f)

		die(err)
		data := map[string]any{
			"name":        dict["name"],
			"tests":       tests,
			"description": dict["description"]}

		json_dt, err := json.Marshal(data)
		die(err)
		if err != nil {
			die(err)
			return
		}

		contractGas := gas.New(act)
		_, err = act.WaitSuccess(contractGas.Transfer(act.Sender(),
			tasktoken.Hash, big.NewInt(10_0000_0000), json_dt))
		die(err)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Successful")

	})

	http.HandleFunc("/add_solution", func(w http.ResponseWriter, r *http.Request) {
		smt, err := requestToDict(r)
		if err != nil {
			http_die(w, "requestToDict", err)
			return
		}
		dict := *smt

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

		contrTask := tasktoken.New(act)
		die(err)

		taskNFT, err := getTaskById(contrTask, dict["taskid"].(string))
		die(err)

		ok := checkIfPassesTests(string(taskNFT.Tests), dict["srccode"].(string))
		if !ok {
			die(errors.New("didn't pass tests"))
		}

		taskid, err := hex.DecodeString(dict["taskid"].(string))
		die(err)
		data := map[string]any{
			"taskid":        taskid,
			"srccode":       dict["srccode"],
			"taskassesment": dict["taskassesment"],
			"description":   dict["description"],
		}

		json_dt, err := json.Marshal(data)

		if err != nil {
			die(err)
			return
		}
		contractGas := gas.New(act)
		_, err = act.WaitSuccess(contractGas.Transfer(act.Sender(),
			solutiontoken.Hash, big.NewInt(10_0000_0000), json_dt))
		die(err)
		if err != nil {
			http_die(w, "Transfer", err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Successful")

	})

	http.HandleFunc("/get_all_tasks", func(w http.ResponseWriter, r *http.Request) {

		smt, err := requestToDict(r)
		if err != nil {
			http_die(w, "requestToDict", err)
			return
		}
		dict := *smt

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

		c := tasktoken.New(act)
		die(err)

		res := listTasks(c)

		bytes, err := json.Marshal(res)
		die(err)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "%s", string(bytes))

	})
	http.HandleFunc("/get_all_solutions", func(w http.ResponseWriter, r *http.Request) {

		smt, err := requestToDict(r)
		if err != nil {
			http_die(w, "requestToDict", err)
			return
		}
		dict := *smt

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

		c := solutiontoken.New(act)
		die(err)
		res := listSolutions(c)
		for t := range res {
			println(t)
		}

		bytes, err := json.Marshal(res)
		die(err)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "%s", string(bytes))

	})

	http.ListenAndServe("localhost:8040", nil)
	return nil
}

func checkIfPassesTests(tests string, srcCode string) bool {

	err := os.WriteFile("./solutionCheck/main.cpp", []byte(srcCode), 0755)
	die(err)
	if err != nil {
		return false
	}
	err = os.WriteFile("./solutionCheck/tests.json", []byte(tests), 0755)
	die(err)
	if err != nil {
		return false
	}
	cmd := exec.Command("python3", "./solutionCheck/solution_check.py")
	stdout, err := cmd.CombinedOutput()
	println(string(stdout))
	die(err)
	if !strings.HasSuffix(string(stdout), "success\n") || err != nil {
		return false
	}

	return true
}

func listTasks(c *tasktoken.Contract) []NFTTask {
	res, err := c.TokensExpanded(10)
	die(err)

	var list []NFTTask
	for _, re := range res {
		prop, err := c.Properties(re)
		die(err)

		list = append(list, parseTask(prop.Value().([]stackitem.MapElement)))
	}

	return list
}
func listSolutions(c *solutiontoken.Contract) []NFTSolution {
	res, err := c.TokensExpanded(10)
	die(err)

	var list []NFTSolution
	for _, re := range res {
		prop, err := c.Properties(re)
		die(err)
		list = append(list, parseSolution(prop.Value().([]stackitem.MapElement)))
	}
	return list
}

func getTaskById(c *tasktoken.Contract, taskId string) (NFTTask, error) {
	list := listTasks(c)
	for _, task := range list {
		if task.ID == taskId {
			return task, nil
		}
	}
	return NFTTask{}, errors.New("no such task")
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

func parseTask(items []stackitem.MapElement) NFTTask {
	var res NFTTask

	for _, item := range items {
		k, err := item.Key.TryBytes()
		die(err)
		v, err := item.Value.TryBytes()
		die(err)

		kStr := string(k)
		switch kStr {
		case "id":
			res.ID = hex.EncodeToString(v)
		case "owner":
			res.Owner, err = address.StringToUint160(string(v))
			die(err)
		case "name":
			res.Name = string(v)
		case "nsolutions":
			res.NSolutions, err = strconv.Atoi(string(v))
			die(err)
		case "tests":
			res.Tests = string(v)
		case "averassesment":
			res.AverAssesment, err = strconv.Atoi(string(v))
			die(err)
		case "description":
			res.Description = string(v)
		}
	}

	return res
}

func parseSolution(items []stackitem.MapElement) NFTSolution {
	var res NFTSolution

	for _, item := range items {
		k, err := item.Key.TryBytes()
		die(err)
		v, err := item.Value.TryBytes()
		die(err)
		kStr := string(k)
		switch kStr {
		case "id":
			id := hex.EncodeToString(v)
			die(err)
			res.ID = id
		case "ownerid":
			res.Owner, err = address.StringToUint160(string(v))
			die(err)
		case "taskid":
			res.TaskId = hex.EncodeToString(v)
		case "taskassesment":
			res.TaskAssesment, err = strconv.Atoi(string(v))
			die(err)
		case "srccode":
			res.SrcCode = string(v)
		case "description":
			res.Description = string(v)
		case "nassesments":
			res.TaskAssesment, err = strconv.Atoi(string(v))
			die(err)
		case "averassesment":
			res.AverAssesment, err = strconv.Atoi(string(v))
			die(err)
		}
	}

	return res
}
