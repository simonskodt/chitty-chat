package logic

import "sync"

type Lamport struct {
	lamportTimestamp int32
	mu sync.Mutex
}

func (l *Lamport) IncrementLamport() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lamportTimestamp++
}

func (l *Lamport) IncrementMaxLamport(number int32) {
	l.mu.Lock()
	defer l.mu.Unlock()

	highestNumber := maxLamport(l.lamportTimestamp, number)
	l.lamportTimestamp = highestNumber + 1
}

func (l *Lamport) GetTimestamp() (int32) {
	return l.lamportTimestamp
}

func maxLamport(a int32, b int32) (int32) {
	if (a > b) {
		return a
	}
	return b
}