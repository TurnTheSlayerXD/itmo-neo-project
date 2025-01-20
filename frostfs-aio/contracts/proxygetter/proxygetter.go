package proxygetter

import (
	"github.com/nspcc-dev/neo-go/pkg/interop"
	"github.com/nspcc-dev/neo-go/pkg/interop/contract"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/management"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/oracle"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/std"
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
)

const itemKey = "Counter"

func Set(val int) {
	ctx := storage.GetContext()
	storage.Put(ctx, itemKey, val)
}

func Get() {
	get()
}

func get() int {
	ctx := storage.GetReadOnlyContext()
	itemValue := storage.Get(ctx, itemKey)
	return itemValue.(int)
}

func SetRemote(hash interop.Hash160, val int) {
	contract.Call(hash, "set", contract.AllowCall, val)
}
func GetRemote(hash interop.Hash160, val int) {
	contract.Call(hash, "get", contract.AllowCall, val)
}

func SetToTestsNumber(groupID int) {
	filter := []byte("$.groups[" + std.Itoa10(groupID) + "]")
	oracle.Request("https://git.frostfs.info/TrueCloudLab/s3-tests-parser/src/branch/master/internal/s3/resources/tests-struct.json",
		filter,
		"cbSetToTestsNumber",
		groupID,
		2*oracle.MinimumResponseGas)
}

func CbSetToTestsNumber(url string, userData any, code int, result []byte) {
	callingHash := runtime.GetCallingScriptHash()
	if !callingHash.Equals(oracle.Hash) {
		panic("not called from the oracle contract")
	}
	if code != oracle.Success {
		panic("not called from the oracle contract")
	}
	runtime.Log("result for " + url + " is: " + string(result))
	resultLen := len(result)
	data := std.JSONDeserialize(result[1 : resultLen-1]).(map[string]any)
	tests := data["tests"].([]string)
	groupName := data["name"].(string)

	groupID := userData.(int)

	runtime.Notify("setToTestNumber", groupID, groupName, len(tests))
	ctx := storage.GetContext()
	storage.Put(ctx, itemKey, len(tests))
}

func OnNep17Payment(from interop.Hash160, amount int, data interface{}) {

}

func Update(script []byte, manifest []byte, data any) {
	management.UpdateWithData(script, manifest, data)
	runtime.Log("proxygetter contract updated")
}
