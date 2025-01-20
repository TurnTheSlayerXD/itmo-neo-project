package counter

import (
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
)

const itemKey = "Counter"

func _deploy(_ interface{}, isUpdate bool) {
	if !isUpdate {
		ctx := storage.GetContext()
		runtime.Notify("info", []byte("Counter not set in Storage. Setting to 0"))
		itemValue := 0
		storage.Put(ctx, itemKey, itemValue)
		runtime.Notify("info", []byte("Counter in Storage is now initialised"))
	}
}

func Main() interface{} {
	ctx := storage.GetContext()
	itemValue := storage.Get(ctx, itemKey)
	runtime.Notify("info", []byte("Value read from Storage"))

	runtime.Notify("info", []byte("Incrementing Counter by 1"))
	itemValue = itemValue.(int) + 1

	storage.Put(ctx, itemKey, itemValue)
	runtime.Notify("info", []byte("New counter value written into Storage"))
	return itemValue
}
