package common

import (
	"sync"
)

type RWLockDict struct{
	 lock *sync.RWMutex
	 set map[interface{}] interface{}
}

func NewRWLockDict() *RWLockDict{
	return &RWLockDict{
	lock : new(sync.RWMutex),
	set  : make(map[interface{}]interface{}),
	}
}

func (self *RWLockDict)HasKey(key interface{}) interface{}{
	self.lock.RLock()
	defer self.lock.RUnlock()
	if _, ok := self.set[key]; ok{
		return true
	}
	return false
}

func (self *RWLockDict)Get(key interface{}) interface{}{
	self.lock.RLock()
	defer self.lock.RUnlock()
	if val, ok := self.set[key]; ok{
		return val
	}
	return nil
}

func (self *RWLockDict)Set(key interface{}, val interface{}) {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.set[key] = val
}

func (self *RWLockDict)Del(key interface{}){
	self.lock.Lock()
	defer self.lock.Unlock()
	delete(self.set, key)
}

func (self *RWLockDict)Len() int {
	self.lock.RLock()
	defer self.lock.RUnlock()
	return len(self.set)
}

func (self *RWLockDict) Items() map[interface{}]interface{} {
	set := make(map[interface{}]interface{})
	self.lock.RLock()
	defer self.lock.RUnlock()
	for key, value := range self.set {
		set[key] = value
	}
	return set
}