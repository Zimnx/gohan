package etcdv3

import (
	"testing"
	"time"

	gohan_sync "github.com/cloudwan/gohan/sync"
	etcd "github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
)

var endpoints = []string{"localhost:2379"}

func TestNewSyncTimeout(t *testing.T) {
	done := make(chan struct{})
	go func() {
		_, err := NewSync([]string{"invalid:1000"}, time.Millisecond*100)
		if err == nil {
			t.Errorf("nil returned for error")
		}
		close(done)
	}()
	select {
	case <-time.NewTimer(time.Millisecond * 200).C:
		t.Errorf("timeout didn't work")
	case <-done:
	}
}

func TestNonEmptyUpdate(t *testing.T) {
	sync := newSync(t)
	sync.etcdClient.Delete(context.Background(), "/", etcd.WithPrefix())

	path := "/path/to/somewhere"
	data := "blabla"
	err := sync.Update(path, data)
	if err != nil {
		t.Errorf("unexpected error")
	}

	node, err := sync.Fetch(path)
	if err != nil {
		t.Errorf("unexpected error")
	}
	if node.Key != path || node.Value != data || len(node.Children) != 0 {
		t.Errorf("unexpected node: %+v", node)
	}

	err = sync.Delete(path)
	if err != nil {
		t.Errorf("unexpected error")
	}

	node, err = sync.Fetch(path)
	if err == nil {
		t.Errorf("unexpected non error")
	}
}

func TestEmptyUpdate(t *testing.T) {
	sync := newSync(t)
	sync.etcdClient.Delete(context.Background(), "/", etcd.WithPrefix())

	path := "/path/to/somewhere"
	data := ""
	err := sync.Update(path, data)
	if err != nil {
		t.Errorf("unexpected error")
	}

	// not found because v3 doesn't support directories
	_, err = sync.Fetch(path)
	if err == nil {
		t.Errorf("unexpected error")
	}
}

func TestRecursiveUpdate(t *testing.T) {
	sync := newSync(t)
	sync.etcdClient.Delete(context.Background(), "/", etcd.WithPrefix())

	base := "/path/to/somewhere"
	items := map[string]string{
		base:                 "",
		base + "/inside":     "inside",
		base + "/else":       "",
		base + "/else/child": "child",
	}

	for path, data := range items {
		err := sync.Update(path, data)
		if err != nil {
			t.Errorf("unexpected error")
		}
	}
	err := sync.Update(base+"invalid", "should not be included")
	if err != nil {
		t.Errorf("unexpected error")
	}

	// not found because v3 doesn't support directories
	node, err := sync.Fetch(base)
	if err != nil {
		t.Errorf("unexpected error")
	}

	if node.Key != base || node.Value != items[base] || len(node.Children) != 2 {
		t.Errorf("unexpected node: %+v", node)
	}
	if node.Children[0].Key != base+"/else" || node.Children[0].Value != items[base+"/else"] || len(node.Children[0].Children) != 1 {
		t.Errorf("unexpected node: %+v", node.Children[0])
	}
	if node.Children[0].Children[0].Key != base+"/else/child" || node.Children[0].Children[0].Value != items[base+"/else/child"] || len(node.Children[0].Children[0].Children) != 0 {
		t.Errorf("unexpected node: %+v", node.Children[0].Children[0])
	}
	if node.Children[1].Key != base+"/inside" || node.Children[1].Value != items[base+"/inside"] || len(node.Children[1].Children) != 0 {
		t.Errorf("unexpected node: %+v", node.Children[1])
	}
}

func TestLockUnblocking(t *testing.T) {
	sync0 := newSync(t)
	sync1 := newSync(t)
	sync0.etcdClient.Delete(context.Background(), "/", etcd.WithPrefix())

	path := "/path/lock"
	err := sync0.Lock(path, false)
	if err != nil {
		t.Errorf("unexpected error")
	}
	err = sync1.Lock(path, false)
	if err == nil {
		t.Errorf("unexpected non error")
	}

	if sync0.HasLock(path) != true {
		t.Errorf("unexpected false")
	}
	if sync1.HasLock(path) != false {
		t.Errorf("unexpected true")
	}

	err = sync0.Unlock(path)
	if err != nil {
		t.Errorf("unexpected error")
	}
	err = sync1.Lock(path, false)
	if err != nil {
		t.Errorf("unexpected  error")
	}

	if sync0.HasLock(path) != false {
		t.Errorf("unexpected true")
	}
	if sync1.HasLock(path) != true {
		t.Errorf("unexpected false")
	}
}

func TestLockBlocking(t *testing.T) {
	sync0 := newSync(t)
	sync1 := newSync(t)
	sync0.etcdClient.Delete(context.Background(), "/", etcd.WithPrefix())

	path := "/path/lock"
	err := sync0.Lock(path, true)
	if err != nil {
		t.Errorf("unexpected error")
	}
	locked1 := make(chan struct{})
	go func() {
		err := sync1.Lock(path, true)
		if err != nil {
			t.Errorf("unexpected error")
		}
		close(locked1)
	}()

	time.Sleep(time.Millisecond * 100)
	select {
	case <-locked1:
		t.Errorf("blocking failed")
	default:
	}

	if sync0.HasLock(path) != true {
		t.Errorf("unexpected false")
	}
	if sync1.HasLock(path) != false {
		t.Errorf("unexpected true")
	}

	err = sync0.Unlock(path)
	if err != nil {
		t.Errorf("unexpected error")
	}
	time.Sleep(time.Millisecond * 200)
	<-locked1

	if sync0.HasLock(path) != false {
		t.Errorf("unexpected true")
	}
	if sync1.HasLock(path) != true {
		t.Errorf("unexpected false")
	}
}

func TestWatch(t *testing.T) {
	sync := newSync(t)
	sync.etcdClient.Delete(context.Background(), "/", etcd.WithPrefix())

	path := "/path/to/watch"
	responseChan := make(chan *gohan_sync.Event)
	stopChan := make(chan bool)

	sync.etcdClient.Put(context.Background(), path+"/existing", `{"existing": true}`)

	go func() {
		err := sync.Watch(path, responseChan, stopChan, gohan_sync.RevisionCurrent)
		if err != nil {
			t.Errorf("failed to watch")
		}
	}()

	resp := <-responseChan
	if resp.Action != "get" || resp.Key != path+"/existing" || resp.Data["existing"].(bool) != true {
		t.Errorf("mismatch response: %+v", resp)
	}

	sync.etcdClient.Put(context.Background(), path+"/new", `{"existing": false}`)
	resp = <-responseChan
	if resp.Action != "set" || resp.Key != path+"/new" || resp.Data["existing"].(bool) != false {
		t.Errorf("mismatch response: %+v", resp)
	}

	sync.etcdClient.Delete(context.Background(), path+"/existing")
	resp = <-responseChan
	if resp.Action != "delete" || resp.Key != path+"/existing" || len(resp.Data) != 0 {
		t.Errorf("mismatch response: %+v", resp)
	}
}

func newSync(t *testing.T) *Sync {
	sync, err := NewSync(endpoints, time.Millisecond*100)
	if err != nil {
		t.Errorf("unexpected error")
	}
	return sync
}
