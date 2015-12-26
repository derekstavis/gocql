// +build ccm

package gocql

import (
	"github.com/gocql/gocql/ccm_test"
	"testing"
	"time"
)

func TestEventDiscovery(t *testing.T) {
	if err := ccm.AllUp(); err != nil {
		t.Fatal(err)
	}

	session := createSession(t)
	defer session.Close()

	status, err := ccm.Status()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("status=%+v\n", status)

	session.pool.mu.RLock()
	poolHosts := session.pool.hostConnPools // TODO: replace with session.ring
	t.Logf("poolhosts=%+v\n", poolHosts)
	// check we discovered all the nodes in the ring
	for _, host := range status {
		if _, ok := poolHosts[host.Addr]; !ok {
			t.Errorf("did not discover %q", host.Addr)
		}
	}
	session.pool.mu.RUnlock()
	if t.Failed() {
		t.FailNow()
	}
}

func TestEventNodeDownControl(t *testing.T) {
	const targetNode = "node1"
	t.Log("marking " + targetNode + " as down")
	if err := ccm.AllUp(); err != nil {
		t.Fatal(err)
	}

	session := createSession(t)
	defer session.Close()

	if err := ccm.NodeDown(targetNode); err != nil {
		t.Fatal(err)
	}

	status, err := ccm.Status()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("status=%+v\n", status)
	t.Logf("marking node %q down: %v\n", targetNode, status[targetNode])

	time.Sleep(5 * time.Second)

	session.pool.mu.RLock()

	poolHosts := session.pool.hostConnPools
	node := status[targetNode]
	t.Logf("poolhosts=%+v\n", poolHosts)

	if _, ok := poolHosts[node.Addr]; ok {
		session.pool.mu.RUnlock()
		t.Fatal("node not removed after remove event")
	}
	session.pool.mu.RUnlock()
}

func TestEventNodeDown(t *testing.T) {
	const targetNode = "node3"
	if err := ccm.AllUp(); err != nil {
		t.Fatal(err)
	}

	session := createSession(t)
	defer session.Close()

	if err := ccm.NodeDown(targetNode); err != nil {
		t.Fatal(err)
	}

	status, err := ccm.Status()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("status=%+v\n", status)
	t.Logf("marking node %q down: %v\n", targetNode, status[targetNode])

	time.Sleep(5 * time.Second)

	session.pool.mu.RLock()
	defer session.pool.mu.RUnlock()

	poolHosts := session.pool.hostConnPools
	node := status[targetNode]
	t.Logf("poolhosts=%+v\n", poolHosts)

	if _, ok := poolHosts[node.Addr]; ok {
		t.Fatal("node not removed after remove event")
	}
}

func TestEventNodeUp(t *testing.T) {
	if err := ccm.AllUp(); err != nil {
		t.Fatal(err)
	}

	status, err := ccm.Status()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("status=%+v\n", status)

	session := createSession(t)
	defer session.Close()

	const targetNode = "node2"

	if err := ccm.NodeDown(targetNode); err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	session.pool.mu.RLock()

	poolHosts := session.pool.hostConnPools
	t.Logf("poolhosts=%+v\n", poolHosts)
	node := status[targetNode]

	if _, ok := poolHosts[node.Addr]; ok {
		session.pool.mu.RUnlock()
		t.Fatal("node not removed after remove event")
	}
	session.pool.mu.RUnlock()

	if err := ccm.NodeUp(targetNode); err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)

	session.pool.mu.RLock()
	t.Logf("poolhosts=%+v\n", poolHosts)
	if _, ok := poolHosts[node.Addr]; !ok {
		session.pool.mu.RUnlock()
		t.Fatal("node not added after node added event")
	}
	session.pool.mu.RUnlock()
}
