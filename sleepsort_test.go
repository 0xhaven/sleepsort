package sleepsort

import (
	"fmt"
	"runtime"
	"sort"
	"testing"
	"time"
)

func TestKill(t *testing.T) {
	sorter := NewSleepSorter(
		NewBoundedRandIterator(1<<8, 1<<32),
		time.Nanosecond*1<<15,
	)
	running := make(chan bool)
	output := make(chan error)
	go func() {
		running <- true
		output <- sorter.Run()
	}()
	<-running
	sorter.Kill()
	if nil == <-output {
		t.Fail()
	}
}

func (r *boundedRandIterator) MakeArrayAndIterator() (array []int, it GetSetIntIterator) {
	array = make([]int, r.NumLeft())
	it = NewArrayIterator(array)
	for r.NumLeft() > 0 && it.NumLeft() > 0 {
		it.SetNext(r.Next())
	}
	it.Reset()
	r.Reset()
	return
}

func TestSortArray(t *testing.T) {
	array, it := NewBoundedRandIterator(1<<10, 1).(*boundedRandIterator).MakeArrayAndIterator()
	err := NewSleepSorter(it, time.Nanosecond<<0).Run()
	if err != nil {
		t.Error(err)
	}
	if !sort.IntsAreSorted(array) {
		t.Error("Array not sorted")
	}
}

func between(start uint8, end uint8) []uint8 {
	arr := make([]uint8, end-start+1)
	for i, _ := range arr {
		arr[i] = start
		start++
	}
	return arr
}

var (
	dataLens    = between(8, 16)
	bitSizes    = between(1, 16)
	minTimeStep = time.Nanosecond
	maxTimeStep = 100 * time.Millisecond
	numSorters  = 5
)

func TestAllSizes(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	for _, bitSize := range bitSizes {
		for _, dataLen := range dataLens {
			err := RunTimesteps(t, dataLen, bitSize, numSorters)
			if err != nil {
				t.Error(err)
			}
		}
	}
}
func BenchmarkMedium(b *testing.B) {
	err := NewSleepSorter(NewBoundedRandIterator(1<<20-1, 1<<1-1), time.Nanosecond).Run()
	if err != nil {
		b.Error(err)
	}
}

func RunTimesteps(t *testing.T, dataLen, bitSize uint8, numSorters int) error {
	defer runtime.GC()
	sortErrors := make(chan error, numSorters)
	fmt.Print("Trying ")
	for timeStep := minTimeStep; timeStep < maxTimeStep; timeStep <<= 1 {
		fmt.Printf("%s... ", timeStep)
		sorters := make([]*SleepSorter, numSorters)
		for i, _ := range sorters {
			sorters[i] = NewSleepSorter(
				NewBoundedRandIterator(1<<dataLen, 1<<bitSize), timeStep)
			go func(sorter *SleepSorter) {
				sortErrors <- sorter.Run()
			}(sorters[i])
		}
		for i := 0; i < numSorters; i++ {
			if nil == <-sortErrors {
				fmt.Printf("Sorting %d %d-bit integers is possible with %s timesteps\n",
					1<<dataLen, bitSize, timeStep)
				for _, sorter := range sorters {
					sorter.Kill()
				}
				return nil
			}
		}
	}
	return fmt.Errorf("No timesteps between %s and %s succeeded\n",
		minTimeStep, maxTimeStep)
}
