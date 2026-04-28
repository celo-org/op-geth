// Copyright 2026 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package discover

import (
	"runtime"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

// TestTableMutexNotHeldDuringFeedSend is a regression test for a deadlock
// fixed by upstream PR #33665 (https://github.com/ethereum/go-ethereum/pull/33665).
//
// # The bug
//
// nodeFeed.Send was called from nodeAdded() while tab.mutex was held. If a
// subscriber to nodeFeed was not currently reading from its channel - e.g.,
// because it was blocked trying to re-acquire tab.mutex (the case in
// waitForNodes' between-iterations state) - Send would block indefinitely
// while still holding tab.mutex.
//
// In production this manifested as a 24-hour silent stall in peer discovery
// on a Celo node, which became visible only on graceful shutdown:
// FairMix.Close -> bufferIter.Close -> "for range b.buffer" could not
// complete because the buffer producer goroutine was permanently stuck in
// Table.waitForNodes -> tab.mutex.Lock(), which was held by Table.doRefresh
// -> loadSeedNodes -> handleAddNode -> nodeAdded -> nodeFeed.Send.
//
// # What this test verifies
//
// Adding a node via the public API (which routes through Table.loop's
// addNodeCh handler and triggers nodeFeed.Send) must not hold tab.mutex
// while Send is in progress. The test sets up a subscriber whose channel
// is never read, so Send is guaranteed to block. It then verifies that
// tab.mutex is acquirable by an unrelated goroutine while Send is blocked.
func TestTableMutexNotHeldDuringFeedSend(t *testing.T) {
	transport := newPingRecorder()
	tab, db := newTestTable(transport, Config{})
	defer db.Close()
	defer tab.close()

	// Subscribe to nodeFeed but never read from ch. This simulates a
	// waitForNodes goroutine that has subscribed and is currently between
	// iterations - i.e., its channel is not being read because it's blocked
	// trying to re-acquire tab.mutex.
	ch := make(chan *enode.Node) // unbuffered, like waitForNodes
	sub := tab.nodeFeed.Subscribe(ch)
	defer sub.Unsubscribe()

	// Trigger a node addition through the public API. This routes through
	// Table.loop's addNodeCh handler:
	//   - Lock(); handleAddNode(...); Unlock();
	//   - nodeFeed.Send(addedNode)   <- with the fix, OUTSIDE the lock
	//   - addNodeHandled <- ok
	//
	// With the bug, Send happens inside handleAddNode under the mutex, so
	// the mutex is held while Send blocks.
	//
	// addFoundNode itself blocks waiting for addNodeHandled, so we run it
	// in a goroutine.
	added := make(chan bool, 1)
	go func() {
		added <- tab.addFoundNode(nodeAtDistance(tab.self().ID(), 250, intIP(1)), false)
	}()

	// Give the loop time to pick up the addNodeCh op and reach Send.
	// 50ms is generous; on a healthy machine the loop reaches Send in <1ms.
	time.Sleep(50 * time.Millisecond)

	// Verify tab.mutex is acquirable by another goroutine while Send is
	// blocked. With the fix, this succeeds because Send happens after Unlock.
	// With the bug, this would block forever because the loop holds tab.mutex
	// while Send is blocked waiting for our subscriber to receive.
	mutexAcquired := make(chan struct{})
	go func() {
		tab.mutex.Lock()
		tab.mutex.Unlock()
		close(mutexAcquired)
	}()

	select {
	case <-mutexAcquired:
		// Good - mutex is free; the fix is working.
	case <-time.After(3 * time.Second):
		// Drain ch in the background so cleanup (defer tab.close()) can run,
		// then fail with a full goroutine dump to make diagnosis easy.
		go func() {
			for range ch {
			}
		}()
		var buf [1 << 16]byte
		n := runtime.Stack(buf[:], true)
		t.Fatalf("DEADLOCK: tab.mutex held while nodeFeed.Send is blocked - "+
			"PR #33665 regression.\n--- Goroutine dump:\n%s", buf[:n])
	}

	// Drain ch so the loop's pending Send can complete and tab.close() can
	// shut down cleanly via the deferred call.
	go func() {
		for range ch {
		}
	}()

	// Wait for addFoundNode to return.
	select {
	case <-added:
	case <-time.After(3 * time.Second):
		t.Fatal("addFoundNode did not return after Send was unblocked")
	}
}
