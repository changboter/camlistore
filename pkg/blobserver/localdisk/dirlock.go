/*
Copyright 2012 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package localdisk

import (
	"sync"
)

var (
	dirLockMu sync.Mutex // guards rest:
	locksOut  int64
	dirLocks  = map[string]*sync.RWMutex{}
)

func getDirLock(dir string) *sync.RWMutex {
	dirLockMu.Lock()
	defer dirLockMu.Unlock()
	locksOut++
	l, ok := dirLocks[dir]
	if !ok {
		l = new(sync.RWMutex)
		dirLocks[dir] = l
	}
	return l
}

func unlockDirLock() {
	dirLockMu.Lock()
        defer dirLockMu.Unlock()
	locksOut--
	if locksOut == 0 {
		dirLocks = map[string]*sync.RWMutex{}
	}
}

type unlocker interface {
	Unlock()
}

// keepDirectoryLock locks directory and returns the locked object.
// Holding the lock prevents the directory from being deleted.
// The caller must Unlock it when finished.
func keepDirectoryLock(dir string) unlocker {
	mu := getDirLock(dir)
	mu.RLock()
	return keepLock{mu}
}

type keepLock struct {
	mu *sync.RWMutex
}

func (l keepLock) Unlock() {
	l.mu.RUnlock()
	unlockDirLock()
}

// deleteDirectoryLock locks directory and returns the locked object.
// Holding the lock is necessary while deleting the directory.
// The caller must Unlock it when finished.
func deleteDirectoryLock(dir string) unlocker {
	mu := getDirLock(dir)
	mu.Lock()
	return deleteLock{mu}
}

type deleteLock struct {
	mu *sync.RWMutex
}

func (l deleteLock) Unlock() {
	l.mu.Unlock()
	unlockDirLock()
}
