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

const (
	minChunkSize = 4096
)

type Chunk struct {
	data             []byte
	isFirst          bool
	lastChunk        bool
	prevEndsWithWord bool
}

type WordCounter struct {
	wordCount     int64
	workersNumber int
	wg            sync.WaitGroup
	chunks        chan Chunk
}

func isSpace(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	}
	return false
}

func (wc *WordCounter) processChunk(chunk Chunk) int64 {
	var count int64
	inWord := false

	// Проверяем, нужно ли считать слово на стыке чанков
	if !chunk.isFirst && !isSpace(chunk.data[0]) && chunk.prevEndsWithWord {
		count--
	}

	for i := 0; i < len(chunk.data); i++ {
		if isSpace(chunk.data[i]) {
			if inWord {
				count++
				inWord = false
			}
		} else {
			if !inWord {
				inWord = true
			}
		}
	}

	// Если чанк заканчивается словом, увеличиваем счетчик
	if inWord {
		count++
	}

	return count
}

func (wc *WordCounter) worker() {
	defer wc.wg.Done()

	for chunk := range wc.chunks {
		count := wc.processChunk(chunk)
		atomic.AddInt64(&wc.wordCount, count)
	}
}

func main() {
	workersNumber := runtime.NumCPU()
	wordCounter := &WordCounter{
		workersNumber: workersNumber,
		chunks:        make(chan Chunk, workersNumber),
	}

	reader := bufio.NewReaderSize(os.Stdin, minChunkSize*workersNumber)

	// Запускаем воркеров
	wordCounter.wg.Add(workersNumber)
	for i := 0; i < workersNumber; i++ {
		go wordCounter.worker()
	}

	buffer := make([]byte, minChunkSize)
	var prevEndsWithWord bool
	firstChunk := true

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		// Определяем, заканчивается ли текущий чанк словом
		currentEndsWithWord := n > 0 && !isSpace(buffer[n-1])

		chunk := Chunk{
			data:             make([]byte, n),
			isFirst:          firstChunk,
			prevEndsWithWord: prevEndsWithWord,
		}
		copy(chunk.data, buffer[:n])

		wordCounter.chunks <- chunk

		prevEndsWithWord = currentEndsWithWord
		firstChunk = false
	}

	close(wordCounter.chunks)
	wordCounter.wg.Wait()

	fmt.Println(wordCounter.wordCount)
}
