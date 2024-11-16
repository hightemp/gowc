package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

type DataRange struct {
	prevRange    *DataRange
	endsWithWord bool
	data         []byte
}

type WordCounter struct {
	mtx           sync.Mutex
	wg            *sync.WaitGroup
	workersNumber int
	wordCount     int64
	dataChan      chan *DataRange
}

func (w *WordCounter) incCounter(count int64) {
	atomic.AddInt64(&w.wordCount, count)
}

func isSpace(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	}
	return false
}

func (w *WordCounter) worker() {
	defer w.wg.Done()
	for {
		dataRange, ok := <-w.dataChan
		if !ok {
			return
		}

		if dataRange.prevRange != nil && dataRange.prevRange.endsWithWord {
			w.incCounter(-1)
		}

		isWord := false

		var wordCount int64
		for _, char := range dataRange.data {
			if isSpace(char) {
				if isWord {
					wordCount++
					isWord = false
				}
			} else {
				if !isWord {
					isWord = true
				}
			}
		}

		if isWord {
			wordCount++
		}

		w.incCounter(wordCount)
	}
}

func (w *WordCounter) run() {
	for i := 0; i < w.workersNumber; i++ {
		w.wg.Add(1)
		go w.worker()
	}
}

func main() {
	stdinReader := bufio.NewReader(os.Stdin)

	workersNumber := runtime.NumCPU()

	wordCounter := WordCounter{
		workersNumber: workersNumber,
		wg:            &sync.WaitGroup{},
		wordCount:     0,
		dataChan:      make(chan *DataRange),
	}

	chunkSize := stdinReader.Size() / workersNumber
	buffer := make([]byte, chunkSize)

	var prevRange *DataRange = nil

	wordCounter.run()

	for {
		numBytesRead, err := stdinReader.Read(buffer)

		if err != nil {
			if err == io.EOF {
				close(wordCounter.dataChan)
				break
			}
			fmt.Println("Error reading from stdin: ", err)
			os.Exit(10)
		}

		currentRange := DataRange{
			prevRange:    prevRange,
			endsWithWord: numBytesRead > 0 && !isSpace(buffer[numBytesRead-1]),
			data:         make([]byte, numBytesRead),
		}
		copy(currentRange.data, buffer[:numBytesRead])
		wordCounter.dataChan <- &currentRange
		prevRange = &currentRange
	}

	wordCounter.wg.Wait()

	fmt.Println(wordCounter.wordCount)
}
