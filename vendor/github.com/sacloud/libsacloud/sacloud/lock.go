package sacloud

import (
	"fmt"

	"github.com/sacloud/libsacloud/utils/mutexkv"
)

var resourceMu = mutexkv.NewMutexKV()

// LockByKey 任意のキーでのMutexロック
func LockByKey(key string) {
	resourceMu.Lock(key)
}

// UnlockByKey 任意のキーでのMutexアンロック
func UnlockByKey(key string) {
	resourceMu.Unlock(key)
}

// LockByResourceID リソース単位でのMutexロック
func LockByResourceID(resourceID int64) {
	resourceMu.Lock(fmt.Sprintf("%d", resourceID))
}

// UnlockByResourceID リソース単位でのMutexアンロック
func UnlockByResourceID(resourceID int64) {
	resourceMu.Unlock(fmt.Sprintf("%d", resourceID))
}
