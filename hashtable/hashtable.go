package hashtable

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"

	"algorithms/iterator"
)

type Node[K, V any] struct {
	entry Entry[K, V]
	hash  uint64
	next  *Node[K, V]
}

type Entry[K, V any] struct {
	Key   K
	Value V
}

type HashTable[K, V any] struct {
	actualBucketLength uint32
	actualBucketSize   uint32
	sizeItems          uint32
	buckets            []*Node[K, V]
	hasher             hash.Hash64
}

func NewHashTable[K, V any]() *HashTable[K, V] {
	hashTable := HashTable[K, V]{
		actualBucketLength: 2,
		actualBucketSize:   0,
		sizeItems:          0,
		hasher:             fnv.New64(),
	}

	hashTable.buckets = make([]*Node[K, V], hashTable.actualBucketLength)

	return &hashTable
}

func (h *HashTable[K, V]) isFull() bool {
	return h.actualBucketSize > (h.actualBucketLength >> 1)
}

func (h *HashTable[K, V]) resetBucket(newLength uint32) {
	h.actualBucketLength = newLength
	h.buckets = make([]*Node[K, V], h.actualBucketLength)
	h.actualBucketSize = 0
	h.sizeItems = 0
}

func (h *HashTable[K, V]) Resize() {
	// Copy all nodes to the tempBucket
	tempBucket := make([]*Node[K, V], h.actualBucketLength)
	copy(tempBucket, h.buckets)

	h.resetBucket(h.actualBucketLength << 1)

	// Insert all nodes from the tempBucket to the new bucket
	for _, node := range tempBucket {
		if node == nil {
			continue
		}

		for {
			h.Insert(node.entry.Key, node.entry.Value)

			if node.next == nil {
				break
			}

			node = node.next
		}
	}
}

func (h HashTable[K, V]) Hash(key K) (hash uint64, index uint32) {

	hash = h.generateHash(key)
	index = h.generateIndex(hash)

	return
}

func (h HashTable[K, V]) generateHash(key K) (hash uint64) {
	defer h.hasher.Reset()

	keyBuffer := bytes.Buffer{}
	gob.NewEncoder(&keyBuffer).Encode(key)

	h.hasher.Write(keyBuffer.Bytes())
	hash = uint64(h.hasher.Sum64())

	return
}

func (h *HashTable[K, V]) generateIndex(hash uint64) uint32 {
	return uint32(hash % uint64(h.actualBucketLength))
}

func (h *HashTable[K, V]) Insert(key K, value V) {
	hash, index := h.Hash(key)

	newNode := &Node[K, V]{
		hash: hash,
		entry: Entry[K, V]{
			Key:   key,
			Value: value,
		},
	}

	h.insertNode(newNode, index)
}

func (h *HashTable[K, V]) insertNode(newNode *Node[K, V], index uint32) {
	if h.buckets[index] == nil {
		h.buckets[index] = newNode
		h.actualBucketSize++
		h.sizeItems++
	} else {
		h.HandleColision(newNode, h.buckets[index], index)
	}

	if h.isFull() {
		h.Resize()
	}
}

func (h *HashTable[K, V]) HandleColision(newNode *Node[K, V], colidedNode *Node[K, V], index uint32) {
	for {
		if colidedNode.hash == newNode.hash {
			colidedNode.entry = newNode.entry
			return
		}

		if colidedNode.next == nil {
			break
		}

		colidedNode = colidedNode.next
	}

	colidedNode.next = newNode
	h.sizeItems++
}

func (h *HashTable[K, V]) Get(key K) (value V) {

	hash, index := h.Hash(key)
	var node = h.buckets[index]

	if node == nil {
		msg := fmt.Sprintf("key not found: %v", key)
		panic(errors.New(msg))
	}

	for {
		if node.hash == hash {
			return node.entry.Value
		}

		if node.next == nil {
			break
		}

		node = node.next
	}

	msg := fmt.Sprintf("key not found: %v", key)
	panic(errors.New(msg))
}

func (h *HashTable[K, V]) Delete(key K) {
	_, index := h.Hash(key)

	h.buckets[index] = nil
	h.actualBucketSize--
}

func (h *HashTable[K, V]) Size() uint32 {
	return h.sizeItems
}

func (h *HashTable[K, V]) Iter() <-chan Entry[K, V] {
	iterator := make(chan Entry[K, V])

	go func() {
		for _, node := range h.buckets {
			if node == nil {
				continue
			}

			iterator <- Entry[K, V]{
				Key:   node.entry.Key,
				Value: node.entry.Value,
			}

			for node.next != nil {
				node = node.next

				iterator <- Entry[K, V]{
					Key:   node.entry.Key,
					Value: node.entry.Value,
				}
			}
		}

		close(iterator)
	}()

	return iterator
}

func (h *HashTable[K, V]) Map(f func(Entry[K, V]) interface{}) iterator.Collection[interface{}] {
	collection := iterator.NewList[interface{}]()

	for entry := range h.Iter() {
		collection.Append(f(entry))
	}

	return collection
}

func (h *HashTable[K, V]) Filter(f func(Entry[K, V]) bool) iterator.Collection[Entry[K, V]] {
	collection := iterator.NewList[Entry[K, V]]()

	for entry := range h.Iter() {
		if f(entry) {
			collection.Append(entry)
		}
	}

	return collection
}

func (h *HashTable[K, V]) ForEach(f func(Entry[K, V])) {
	for entry := range h.Iter() {
		f(entry)
	}
}
