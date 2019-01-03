// Copyright 2012 Google Inc. All rights reserved.
// Author: Ric Szopa (Ryszard) <ryszard.szopa@gmail.com>
// Modified By: Alpaca Markets
//   - Concurrency via locking

// Package skiplist implements skip list based maps and sets.
//
// Skip lists are a data structure that can be used in place of
// balanced trees. Skip lists use probabilistic balancing rather than
// strictly enforced balancing and as a result the algorithms for
// insertion and deletion in skip lists are much simpler and
// significantly faster than equivalent algorithms for balanced trees.
//
// Skip lists were first described in Pugh, William (June 1990). "Skip
// lists: a probabilistic alternative to balanced
// trees". Communications of the ACM 33 (6): 668–676
package skiplist

import (
	"math/rand"
	"sync"

	"github.com/shopspring/decimal"
)

// TODO(ryszard):
//   - A separately seeded source of randomness

// p is the fraction of nodes with level i pointers that also have
// level i+1 pointers. p equal to 1/4 is a good value from the point
// of view of speed and space requirements. If variability of running
// times is a concern, 1/2 is a better value for p.
const p = 0.25

const DefaultMaxLevel = 32

// A node is a container for key-value pairs that are stored in a skip
// list.
type node struct {
	forward    []*node
	backward   *node
	key, value interface{}
	sync.RWMutex
}

func (n *node) getForwardLen() int {
	n.RLock()
	defer n.RUnlock()
	return len(n.forward)
}
func (n *node) getForward(i int) *node {
	n.RLock()
	defer n.RUnlock()
	if len(n.forward) == 0 {
		return nil
	}
	return n.forward[i]
}
func (n *node) setForward(i int, nn *node) {
	n.Lock()
	defer n.Unlock()
	n.forward[i] = nn
}
func (n *node) appendForward(nn *node) {
	n.Lock()
	defer n.Unlock()
	n.forward = append(n.forward, nn)
}
func (n *node) truncateForward(left, right int) {
	n.Lock()
	defer n.Unlock()
	n.forward = n.forward[left:right]
}
func (n *node) getBackward() *node {
	n.RLock()
	defer n.RUnlock()
	return n.backward
}
func (n *node) setBackward(nn *node) {
	n.Lock()
	defer n.Unlock()
	n.backward = nn
}
func (n *node) getKey() interface{} {
	n.RLock()
	defer n.RUnlock()
	return n.key
}
func (n *node) setKey(key interface{}) {
	n.Lock()
	defer n.Unlock()
	n.key = key
}
func (n *node) getVal() interface{} {
	n.RLock()
	defer n.RUnlock()
	return n.value
}
func (n *node) setVal(val interface{}) {
	n.Lock()
	defer n.Unlock()
	n.value = val
}
func (n *node) getKeyVal() (key, val interface{}) {
	n.RLock()
	defer n.RUnlock()
	return n.key, n.value
}
func (n *node) setKeyVal(key, val interface{}) {
	n.Lock()
	defer n.Unlock()
	n.key, n.value = key, val
}

// next returns the next node in the skip list containing n.
func (n *node) next() *node {
	return n.getForward(0)
}

// previous returns the previous node in the skip list containing n.
func (n *node) previous() *node {
	return n.getBackward()
}

// hasNext returns true if n has a next node.
func (n *node) hasNext() bool {
	return n.next() != nil
}

// hasPrevious returns true if n has a previous node.
func (n *node) hasPrevious() bool {
	return n.previous() != nil
}

// A SkipList is a map-like data structure that maintains an ordered
// collection of key-value pairs. Insertion, lookup, and deletion are
// all O(log n) operations. A SkipList can efficiently store up to
// 2^MaxLevel items.
//
// To iterate over a skip list (where s is a
// *SkipList):
//
//	for i := s.Iterator(); i.Next(); {
//		// do something with i.Key() and i.Value()
//	}
type SkipList struct {
	lessThan func(l, r interface{}) bool
	header   *node
	footer   *node
	length   int
	// MaxLevel determines how many items the SkipList can store
	// efficiently (2^MaxLevel).
	//
	// It is safe to increase MaxLevel to accomodate more
	// elements. If you decrease MaxLevel and the skip list
	// already contains nodes on higer levels, the effective
	// MaxLevel will be the greater of the new MaxLevel and the
	// level of the highest node.
	//
	// A SkipList with MaxLevel equal to 0 is equivalent to a
	// standard linked list and will not have any of the nice
	// properties of skip lists (probably not what you want).
	MaxLevel int
	sync.RWMutex
}

func (s *SkipList) getHeader() *node {
	s.RLock()
	defer s.RUnlock()
	return s.header
}
func (s *SkipList) setHeader(n *node) {
	s.Lock()
	defer s.Unlock()
	s.header = n
}
func (s *SkipList) getFooter() *node {
	s.RLock()
	defer s.RUnlock()
	return s.footer
}
func (s *SkipList) setFooter(n *node) {
	s.Lock()
	defer s.Unlock()
	s.footer = n
}
func (s *SkipList) getLength() int {
	s.RLock()
	defer s.RUnlock()
	return s.length
}
func (s *SkipList) setLength(n int) {
	s.Lock()
	defer s.Unlock()
	s.length = n
}
func (s *SkipList) lengthAdd(n int) {
	s.Lock()
	defer s.Unlock()
	s.length += n
}

// Len returns the length of s.
func (s *SkipList) Len() int {
	return s.getLength()
}

// Iterator is an interface that you can use to iterate through the
// skip list (in its entirety or fragments). For an use example, see
// the documentation of SkipList.
//
// Key and Value return the key and the value of the current node.
type Iterator interface {
	// Next returns true if the iterator contains subsequent elements
	// and advances its state to the next element if that is possible.
	Next() (ok bool)
	// Previous returns true if the iterator contains previous elements
	// and rewinds its state to the previous element if that is possible.
	Previous() (ok bool)
	// Key returns the current key.
	Key() interface{}
	// Value returns the current value.
	Value() interface{}
	// Seek reduces iterative seek costs for searching forward into the Skip List
	// by remarking the range of keys over which it has scanned before.  If the
	// requested key occurs prior to the point, the Skip List will start searching
	// as a safeguard.  It returns true if the key is within the known range of
	// the list.
	Seek(key interface{}) (ok bool)
	// Close this iterator to reap resources associated with it.  While not
	// strictly required, it will provide extra hints for the garbage collector.
	Close()
}

type iter struct {
	current *node
	key     interface{}
	list    *SkipList
	value   interface{}
	sync.RWMutex
}

func (i *iter) getKey() interface{} {
	i.RLock()
	defer i.RUnlock()
	return i.key
}
func (i *iter) setKey(key interface{}) {
	i.Lock()
	defer i.Unlock()
	i.key = key
}
func (i *iter) getVal() interface{} {
	i.RLock()
	defer i.RUnlock()
	return i.value
}
func (i *iter) setVal(val interface{}) {
	i.Lock()
	defer i.Unlock()
	i.value = val
}
func (i *iter) getCurrent() *node {
	i.RLock()
	defer i.RUnlock()
	return i.current
}
func (i *iter) setCurrent(n *node) {
	i.Lock()
	defer i.Unlock()
	i.current = n
	i.key, i.value = i.current.getKeyVal()
}

func (i iter) Key() interface{} {
	return i.getKey()
}

func (i iter) Value() interface{} {
	return i.getVal()
}

func (i *iter) Next() bool {
	if !i.getCurrent().hasNext() {
		return false
	}
	i.setCurrent(i.getCurrent().next())
	return true
}

func (i *iter) Previous() bool {
	if !i.getCurrent().hasPrevious() {
		return false
	}

	i.setCurrent(i.getCurrent().previous())

	return true
}

func (i *iter) Seek(key interface{}) (ok bool) {
	current := i.getCurrent()
	list := i.list

	// If the existing iterator outside of the known key range, we should set the
	// position back to the beginning of the list.
	if current == nil {
		current = list.getHeader()
	}

	// If the target key occurs before the current key, we cannot take advantage
	// of the heretofore spent traversal cost to find it; resetting back to the
	// beginning is the safest choice.
	if current.getKey() != nil && list.lessThan(key, current.getKey()) {
		current = list.getHeader()
	}

	// We should back up so that we can seek to our present value if that
	// is requested for whatever reason.
	if current.getBackward() == nil {
		current = list.getHeader()
	} else {
		current = current.getBackward()
	}

	current = list.getPath(current, nil, key)

	if current == nil {
		return
	}

	i.setCurrent(current)

	return true
}

func (i *iter) Close() {
	i.Lock()
	defer i.Unlock()
	i.key = nil
	i.value = nil
	i.current = nil
	i.list = nil
}

type rangeIterator struct {
	iter
	upperLimit interface{}
	lowerLimit interface{}
}

func (i *rangeIterator) Next() bool {
	if !i.getCurrent().hasNext() {
		return false
	}

	next := i.getCurrent().next()

	if !i.list.lessThan(next.getKey(), i.upperLimit) {
		return false
	}

	i.setCurrent(i.getCurrent().next())
	return true
}

func (i *rangeIterator) Previous() bool {
	if !i.getCurrent().hasPrevious() {
		return false
	}

	previous := i.getCurrent().previous()

	if i.list.lessThan(previous.key, i.lowerLimit) {
		return false
	}

	i.setCurrent(i.getCurrent().previous())
	return true
}

func (i *rangeIterator) Seek(key interface{}) (ok bool) {
	if i.list.lessThan(key, i.lowerLimit) {
		return
	} else if !i.list.lessThan(key, i.upperLimit) {
		return
	}

	return i.iter.Seek(key)
}

func (i *rangeIterator) Close() {
	i.iter.Close()
	i.Lock()
	defer i.Unlock()
	i.upperLimit = nil
	i.lowerLimit = nil
}

// Iterator returns an Iterator that will go through all elements s.
func (s *SkipList) Iterator() Iterator {
	return &iter{
		current: s.getHeader(),
		list:    s,
	}
}

// Seek returns a bidirectional iterator starting with the first element whose
// key is greater or equal to key; otherwise, a nil iterator is returned.
func (s *SkipList) Seek(key interface{}) Iterator {
	current := s.getPath(s.getHeader(), nil, key)
	if current == nil {
		return nil
	}

	key, value := current.getKeyVal()
	return &iter{
		current: current,
		key:     key,
		list:    s,
		value:   value,
	}
}

// SeekToFirst returns a bidirectional iterator starting from the first element
// in the list if the list is populated; otherwise, a nil iterator is returned.
func (s *SkipList) SeekToFirst() Iterator {
	if s.getLength() == 0 {
		return nil
	}

	current := s.getHeader().next()
	key, value := current.getKeyVal()

	return &iter{
		current: current,
		key:     key,
		list:    s,
		value:   value,
	}
}

// SeekToLast returns a bidirectional iterator starting from the last element
// in the list if the list is populated; otherwise, a nil iterator is returned.
func (s *SkipList) SeekToLast() Iterator {
	current := s.getFooter()
	if current == nil {
		return nil
	}
	key, val := current.getKeyVal()

	return &iter{
		current: current,
		key:     key,
		list:    s,
		value:   val,
	}
}

// Range returns an iterator that will go through all the
// elements of the skip list that are greater or equal than from, but
// less than to.
func (s *SkipList) Range(from, to interface{}) Iterator {
	start := s.getPath(s.getHeader(), nil, from)
	return &rangeIterator{
		iter: iter{
			current: &node{
				forward:  []*node{start},
				backward: start,
			},
			list: s,
		},
		upperLimit: to,
		lowerLimit: from,
	}
}

func (s *SkipList) level() int {
	return s.getHeader().getForwardLen() - 1
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func (s *SkipList) effectiveMaxLevel() int {
	return maxInt(s.level(), s.MaxLevel)
}

// Returns a new random level.
func (s SkipList) randomLevel() (n int) {
	for n = 0; n < s.effectiveMaxLevel() && rand.Float64() < p; n++ {
	}
	return
}

// Get returns the value associated with key from s (nil if the key is
// not present in s). The second return value is true when the key is
// present.
func (s *SkipList) Get(key interface{}) (value interface{}, ok bool) {
	candidate := s.getPath(s.getHeader(), nil, key)
	if candidate == nil {
		return nil, false
	}

	cKey, cVal := candidate.getKeyVal()
	if !IsEqual(cKey, key) {
		return nil, false
	}
	return cVal, true
}

func IsEqual(l, r interface{}) bool {
	switch val := l.(type) {
	case decimal.Decimal:
		cVal := r.(decimal.Decimal)
		if !val.Equals(cVal) {
			return false
		}
	default:
		if val != r {
			return false
		}
	}
	return true
}

// GetGreaterOrEqual finds the node whose key is greater than or equal
// to min. It returns its value, its actual key, and whether such a
// node is present in the skip list.
func (s *SkipList) GetGreaterOrEqual(min interface{}) (actualKey, value interface{}, ok bool) {
	candidate := s.getPath(s.getHeader(), nil, min)

	if candidate != nil {
		key, val := candidate.getKeyVal()
		return key, val, true
	}
	return nil, nil, false
}

// getPath populates update with nodes that constitute the path to the
// node that may contain key. The candidate node will be returned. If
// update is nil, it will be left alone (the candidate node will still
// be returned). If update is not nil, but it doesn't have enough
// slots for all the nodes in the path, getPath will panic.
func (s *SkipList) getPath(current *node, update []*node, key interface{}) *node {
	depth := current.getForwardLen() - 1

	for i := depth; i >= 0; i-- {
		for current.getForward(i) != nil && s.lessThan(current.getForward(i).getKey(), key) {
			current = current.getForward(i)
		}
		if update != nil {
			update[i] = current
		}
	}
	return current.next()
}

// Sets set the value associated with key in s.
func (s *SkipList) Set(key, value interface{}) {
	if key == nil {
		panic("goskiplist: nil keys are not supported")
	}
	// s.level starts from 0, so we need to allocate one.
	update := make([]*node, s.level()+1, s.effectiveMaxLevel()+1)
	candidate := s.getPath(s.getHeader(), update, key)

	//	if candidate != nil && candidate.key == key {
	if candidate != nil && IsEqual(candidate.getKey(), key) {
		candidate.setVal(value)
		return
	}

	newLevel := s.randomLevel()

	if currentLevel := s.level(); newLevel > currentLevel {
		// there are no pointers for the higher levels in
		// update. Header should be there. Also add higher
		// level links to the header.
		for i := currentLevel + 1; i <= newLevel; i++ {
			update = append(update, s.getHeader())
			s.getHeader().appendForward(nil)
		}
	}

	newNode := &node{
		forward: make([]*node, newLevel+1, s.effectiveMaxLevel()+1),
		key:     key,
		value:   value,
	}

	if previous := update[0]; previous.getKey() != nil {
		newNode.setBackward(previous)
	}

	for i := 0; i <= newLevel; i++ {
		newNode.setForward(i, update[i].getForward(i))
		update[i].setForward(i, newNode)
	}

	s.lengthAdd(1)

	if newNode.getForward(0) != nil {
		if newNode.getForward(0).getBackward() != newNode {
			newNode.getForward(0).setBackward(newNode)
		}
	}

	if s.getFooter() == nil || s.lessThan(s.getFooter().getKey(), key) {
		s.setFooter(newNode)
	}
}

// Delete removes the node with the given key.
//
// It returns the old value and whether the node was present.
func (s *SkipList) Delete(key interface{}) (value interface{}, ok bool) {
	if key == nil {
		panic("goskiplist: nil keys are not supported")
	}
	update := make([]*node, s.level()+1, s.effectiveMaxLevel())
	candidate := s.getPath(s.getHeader(), update, key)

	//	if candidate == nil || candidate.key != key {
	if candidate == nil || !IsEqual(candidate.getKey(), key) {
		return nil, false
	}

	previous := candidate.getBackward()
	if s.getFooter() == candidate {
		s.setFooter(previous)
	}

	next := candidate.next()
	if next != nil {
		next.setBackward(previous)
	}

	for i := 0; i <= s.level() && update[i].getForward(i) == candidate; i++ {
		update[i].setForward(i, candidate.getForward(i))
	}

	for s.level() > 0 && s.getHeader().getForward(s.level()) == nil {
		s.getHeader().truncateForward(0, s.level())
	}
	s.lengthAdd(-1)

	return candidate.getVal(), true
}

// NewCustomMap returns a new SkipList that will use lessThan as the
// comparison function. lessThan should define a linear order on keys
// you intend to use with the SkipList.
func NewCustomMap(lessThan func(l, r interface{}) bool) *SkipList {
	return &SkipList{
		lessThan: lessThan,
		header: &node{
			forward: []*node{nil},
		},
		MaxLevel: DefaultMaxLevel,
	}
}

// Ordered is an interface which can be linearly ordered by the
// LessThan method, whereby this instance is deemed to be less than
// other. Additionally, Ordered instances should behave properly when
// compared using == and !=.
type Ordered interface {
	LessThan(other Ordered) bool
}

// New returns a new SkipList.
//
// Its keys must implement the Ordered interface.
func New() *SkipList {
	comparator := func(left, right interface{}) bool {
		return left.(Ordered).LessThan(right.(Ordered))
	}
	return NewCustomMap(comparator)

}

// NewIntKey returns a SkipList that accepts int keys.
func NewIntMap() *SkipList {
	return NewCustomMap(func(l, r interface{}) bool {
		return l.(int) < r.(int)
	})
}

// NewStringMap returns a SkipList that accepts string keys.
func NewStringMap() *SkipList {
	return NewCustomMap(func(l, r interface{}) bool {
		return l.(string) < r.(string)
	})
}

// Set is an ordered set data structure.
//
// Its elements must implement the Ordered interface. It uses a
// SkipList for storage, and it gives you similar performance
// guarantees.
//
// To iterate over a set (where s is a *Set):
//
//	for i := s.Iterator(); i.Next(); {
//		// do something with i.Key().
//		// i.Value() will be nil.
//	}
type Set struct {
	skiplist SkipList
}

// NewSet returns a new Set.
func NewSet() *Set {
	comparator := func(left, right interface{}) bool {
		return left.(Ordered).LessThan(right.(Ordered))
	}
	return NewCustomSet(comparator)
}

// NewCustomSet returns a new Set that will use lessThan as the
// comparison function. lessThan should define a linear order on
// elements you intend to use with the Set.
func NewCustomSet(lessThan func(l, r interface{}) bool) *Set {
	return &Set{skiplist: SkipList{
		lessThan: lessThan,
		header: &node{
			forward: []*node{nil},
		},
		MaxLevel: DefaultMaxLevel,
	}}
}

// NewIntSet returns a new Set that accepts int elements.
func NewIntSet() *Set {
	return NewCustomSet(func(l, r interface{}) bool {
		return l.(int) < r.(int)
	})
}

// NewStringSet returns a new Set that accepts string elements.
func NewStringSet() *Set {
	return NewCustomSet(func(l, r interface{}) bool {
		return l.(string) < r.(string)
	})
}

// Add adds key to s.
func (s *Set) Add(key interface{}) {
	s.skiplist.Set(key, nil)
}

// Remove tries to remove key from the set. It returns true if key was
// present.
func (s *Set) Remove(key interface{}) (ok bool) {
	_, ok = s.skiplist.Delete(key)
	return ok
}

// Len returns the length of the set.
func (s *Set) Len() int {
	return s.skiplist.Len()
}

// Contains returns true if key is present in s.
func (s *Set) Contains(key interface{}) bool {
	_, ok := s.skiplist.Get(key)
	return ok
}

func (s *Set) Iterator() Iterator {
	return s.skiplist.Iterator()
}

// Range returns an iterator that will go through all the elements of
// the set that are greater or equal than from, but less than to.
func (s *Set) Range(from, to interface{}) Iterator {
	return s.skiplist.Range(from, to)
}

// SetMaxLevel sets MaxLevel in the underlying skip list.
func (s *Set) SetMaxLevel(newMaxLevel int) {
	s.skiplist.Lock()
	defer s.skiplist.Unlock()
	s.skiplist.MaxLevel = newMaxLevel
}

// GetMaxLevel returns MaxLevel fo the underlying skip list.
func (s *Set) GetMaxLevel() int {
	return s.skiplist.MaxLevel
}
