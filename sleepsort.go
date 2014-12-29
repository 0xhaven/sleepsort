package sleepsort

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type GetSetIntIterator interface {
	Next() int
	SetNext(int)
	NumLeft() int
	Reset()
}

type arrayIterator struct {
	array []int
	index int
}

func NewArrayIterator(array []int) GetSetIntIterator {
	return &arrayIterator{array: array}
}
func (it *arrayIterator) Next() (val int) {
	if it.NumLeft() <= 0 {
		log.Panicln("Improper call to Next() with no values remaining")
	}
	val = it.array[it.index]
	it.index++
	return
}
func (it *arrayIterator) SetNext(val int) {
	it.array[it.index] = val
	it.index++
}
func (it *arrayIterator) NumLeft() int { return len(it.array) - it.index }
func (it *arrayIterator) Reset()       { it.index = 0 }

type boundedRandIterator struct {
	rand   *rand.Rand
	size   int
	maxInt int
	index  int
}

func NewBoundedRandIterator(Size int, MaxInt int) GetSetIntIterator {
	return &boundedRandIterator{
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		size:   Size,
		maxInt: MaxInt}
}
func (r *boundedRandIterator) Next() int {
	if r.NumLeft() <= 0 {
		log.Panicln("Improper call to Next() with no values remaining")
	}
	r.index++
	return r.rand.Intn(r.maxInt)
}
func (r *boundedRandIterator) NumLeft() int  { return r.size - r.index }
func (r *boundedRandIterator) SetNext(n int) { r.index++ }
func (r *boundedRandIterator) Reset()        { r.index = 0 }

type SleepSorter struct {
	Iterator   GetSetIntIterator
	TimeStep   time.Duration
	dataLen    int
	output     chan int
	startGroup sync.WaitGroup
	killed     chan bool
}

func NewSleepSorter(it GetSetIntIterator, timeStep time.Duration) *SleepSorter {
	return &SleepSorter{Iterator: it, TimeStep: timeStep, killed: make(chan bool)}
}

func (sorter *SleepSorter) Kill() {
	defer func() { recover() }()
	close(sorter.killed)
}

func (sorter *SleepSorter) Run() error {
	sorter.Iterator.Reset()
	sorter.dataLen = sorter.Iterator.NumLeft()
	if sorter.TimeStep != 0 {
		sorter.TimeStep = time.Millisecond * 1 << 8
	}
	sorter.output = make(chan int, sorter.dataLen)
	sorter.startGroup.Add(sorter.dataLen)
	for sorter.Iterator.NumLeft() > 0 {
		go sorter.sleepAndOutput(sorter.Iterator.Next())
	}
	sorter.Iterator.Reset()
	return sorter.process()
}

func (sorter *SleepSorter) sleepAndOutput(input int) {
	sleepDuration := time.Duration(input) * sorter.TimeStep
	sorter.startGroup.Done()
	sorter.startGroup.Wait()
	select {
	case <-time.After(sleepDuration):
		sorter.output <- input
	case <-sorter.killed:
	}
}

func (sorter *SleepSorter) process() error {
	var max int
	for sorter.Iterator.NumLeft() > 0 {
		select {
		case num := <-sorter.output:
			sorter.Iterator.SetNext(num)
			switch {
			case num > max:
				max = num
			case num < max:
				sorter.Kill()
				return fmt.Errorf("output channel not sorted. %d is less than previously seen %d", num, max)
			}
		case <-sorter.killed:
			return errors.New("Proccessing killed")
		}
	}
	return nil
}
