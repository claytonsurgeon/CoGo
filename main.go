package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const all_letters = "abcdefghijklmnopqrstuvwxyz"

func main() {
	fmt.Println("\n\nStarting...")
	time.Sleep(1 * time.Second)
	// run_count_letters()
	// run_stingy_and_spendy()
	// run_matchRecorder()
	// time.Sleep(101 * time.Second)
	// run_doWork()
	// run_multiplayer()
	// run_read_preffered_locking()
	// run_write_preffered_locking()
	// run_semaphore()
	// run_waitGroup()
	// run_countLetters_waitGroup()
	// run_WaitGroupie()
	// run_fileSearch_WaitGroupie2()
	// run_workAndWait()
	// run_rowMul()
	// run_channels()
	// run_slowReceiver()
	// run_receiver_after_close()
	// run_receiver_with_done()
	// run_findFactors()
	// run_RingTest()

	// for val := range ring.Range() {
	// 	fmt.Println(val)
	// }

	fmt.Println("\n\nTerminating...")
	os.Exit(0)
}

type Channel[M any] struct {
	capaSema *Semaphore
	sizeSema *Semaphore
	mutex    sync.Mutex
	buffer   *Ring[M]
}

func NewChannel[M any](capacity int) *Channel[M] {
	return &Channel[M]{
		capaSema: NewSemaphore(capacity),
		sizeSema: NewSemaphore(0),
		buffer:   NewRing[M](capacity),
	}
}

func (c *Channel[M]) Send(message M) {
	c.capaSema.Acquire()

	c.mutex.Lock()
	c.buffer.Push(message)
	c.mutex.Unlock()

	c.sizeSema.Release()
}

func (c *Channel[M]) Receive() (M, error) {
	c.capaSema.Release()

	c.sizeSema.Acquire()

	c.mutex.Lock()
	val, err := c.buffer.Pop()
	if err != nil {
		var zero M
		return zero, err
	}
	c.mutex.Unlock()

	return val, nil
}

func run_RingTest() {

	ring := NewRing[int](5)

	for i := range 25 {
		ok := ring.Push(i)
		if !ok {
			for range 3 {
				head := ring.head
				val, err := ring.Pop()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Printf("at index: %d, val: %d\n", head, val)
			}
			ring.Push(i)
		}
	}
}

type Ring[M any] struct {
	buffer   []M
	capacity int

	head int
	tail int
	size int
}

func NewRing[M any](capacity int) *Ring[M] {
	if capacity <= 0 {
		panic("Ring buffer capacity must be greater than 0")
	}

	return &Ring[M]{
		buffer:   make([]M, capacity),
		capacity: capacity,
	}
}

func (ring *Ring[M]) Push(item M) (ok bool) {

	if ring.size < ring.capacity {
		ring.size++
	} else {
		return false
	}

	ring.buffer[ring.tail] = item
	ring.tail = (ring.tail + 1) % ring.capacity

	return true
}

func (ring *Ring[M]) Pop() (M, error) {
	if ring.size == 0 {
		var zero M
		return zero, errors.New("ring buffer is empty")
	}

	item := ring.buffer[ring.head]
	// Optional: Zero out the popped element to avoid memory leaks if storing pointers
	// rb.buffer[rb.head] = 0 (not strictly necessary for int, but good practice for complex types)
	ring.head = (ring.head + 1) % ring.capacity
	ring.size--
	return item, nil
}
func (ring *Ring[M]) Range() <-chan M {
	// rb.mu.Lock() // Uncomment for concurrency if reading state needs to be atomic with modifications
	// currentSize := rb.size
	// currentHead := rb.head
	// elements := make([]int, currentSize)
	// for i := 0; i < currentSize; i++ {
	// 	elements[i] = rb.buffer[(currentHead + i) % rb.capacity]
	// }
	// rb.mu.Unlock() // Uncomment for concurrency

	// Create a buffered channel with the current size of the buffer.
	// This allows the goroutine to send all elements and exit quickly.
	// If size is 0, it becomes an unbuffered channel, which is fine.
	ch := make(chan M, ring.size) // Using rb.size directly here for simplicity

	go func() {
		defer close(ch) // Ensure the channel is closed when the goroutine finishes

		// If using the locked approach above, iterate over `elements` slice.
		// For this simpler version, we iterate based on current head and size.
		// This is a snapshot-in-time based on when Range() is called.
		// For true concurrent safety, copy elements under a lock.
		if ring.size == 0 { // Check IsEmpty which might be locked or not depending on choice
			return
		}

		current := ring.head
		count := ring.size
		for range count {
			ch <- ring.buffer[current]
			current = (current + 1) % ring.capacity
		}
	}()
	return ch
}

func run_findFactors() {
	resultCh := make(chan []int)
	go func() {
		resultCh <- findFactors(3419110721)
	}()

	// fmt.Println(findFactors(65536))
	fmt.Println(findFactors(4033836233))
	fmt.Println(<-resultCh)
	// fmt.Println(findFactors(3419110721))
}

func findFactors(num int) []int {
	result := []int{}
	for i := range num {
		if num%(i+1) == 0 {
			result = append(result, i+1)
		}
	}
	return result
}

func run_receiver_with_done() {
	msgChannel := make(chan int)
	// wg := &sync.WaitGroup{}
	go receiver_with_done(msgChannel)
	for i := range 3 {
		fmt.Println(time.Now().Format("15:04:05"), "Sending:", i)
		msgChannel <- i
		time.Sleep(1 * time.Second)
	}
	// wg.Wait()
	// time.Sleep(3 * time.Second)
	close(msgChannel)
	println("done")
}

func receiver_with_done(messages <-chan int) {
	for msg := range messages {
		fmt.Println(time.Now().Format("15:04:05"), "Received:", msg)
		time.Sleep(1 * time.Second)
	}
	// for {
	// 	msg, active := <-messages
	// 	if !active {
	// 		wg.Done()
	// 		return
	// 	}

	// 	fmt.Println(time.Now().Format("15:04:05"), "Received:", msg)
	// 	time.Sleep(1 * time.Second)

	// }
}

func run_receiver_after_close() {
	msgChannel := make(chan int)
	go receiver_after_close(msgChannel)
	for i := 1; i <= 3; i++ {
		fmt.Println(time.Now().Format("15:04:05"), "Sending:", i)
		msgChannel <- i
		time.Sleep(1 * time.Second)
	}
	close(msgChannel)
	time.Sleep(3 * time.Second)
}

func receiver_after_close(messages <-chan int) {
	for {
		msg := <-messages
		fmt.Println(time.Now().Format("15:04:05"), "Received:", msg)
		time.Sleep(1 * time.Second)
	}
}

func run_slowReceiver() {
	msgChan := make(chan int, 6)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go slowReceiver(msgChan, wg)
	for i := range 20 {
		time.Sleep(333 * time.Millisecond)
		size := len(msgChan)
		fmt.Printf("%s Sending: %d. Buffer Size: %d\n", time.Now().Format("15:04:05"), i, size)
		msgChan <- i
	}
	msgChan <- -1
	wg.Wait()
}

func slowReceiver(msgChan chan int, wg *sync.WaitGroup) {
	msg := 0
	for msg != -1 {
		time.Sleep(600 * time.Millisecond)
		msg = <-msgChan
		fmt.Println("Received:", msg)
	}
	wg.Done()
}

func run_channels() {
	msgChan := make(chan string)
	go receiver(msgChan)
	fmt.Println("Sending HELLO...")
	msgChan <- "HELLO"
	fmt.Println("Sending THERE...")
	msgChan <- "THERE"
	fmt.Println("Sending STOP...")
	msgChan <- "STOP"
}

func receiver(msgChan chan string) {
	msg := ""
	for msg != "STOP" {
		msg = <-msgChan
		fmt.Println("Recieved:", msg)
	}
}

const matrixSize = 3

func run_rowMul() {
	var matA, matB, result [matrixSize][matrixSize]int
	barrier := NewBarrier(matrixSize + 1)
	for row := range matrixSize {
		go rowMul(&matA, &matB, &result, row, barrier)
	}

	for range 4 {
		genRandMatrix(&matA)
		genRandMatrix(&matB)

		barrier.Wait()
		barrier.Wait()

		for i := range matrixSize {
			fmt.Println(matA[i], matB[i], result[i])
		}
		fmt.Println()
	}
}

func rowMul(matA, matB, result *[matrixSize][matrixSize]int, row int, barrier *Barrier) {
	for {
		barrier.Wait()
		for col := range matrixSize {
			sum := 0
			for i := range matrixSize {
				sum += matA[row][i] * matB[i][col]
			}
			result[row][col] = sum
		}
		barrier.Wait()
	}
}

func matMul(matA, matb, result *[matrixSize][matrixSize]int) {
	for row := range matrixSize {
		for col := range matrixSize {
			sum := 0
			for i := range matrixSize {
				sum += matA[row][i] * matb[i][col]
			}
			result[row][col] = sum
		}
	}
}

func genRandMatrix(mat *[matrixSize][matrixSize]int) {
	for row := range matrixSize {
		for col := range matrixSize {
			mat[row][col] = rand.IntN(11) - 5
		}
	}
}

func run_workAndWait() {
	barrier := NewBarrier(6)

	go workAndWait("Red", 4, barrier)
	go workAndWait("Orange", 6, barrier)
	go workAndWait("Yellow", 2, barrier)
	go workAndWait("Green", 7, barrier)
	go workAndWait("Blue", 10, barrier)
	go workAndWait("Violet", 5, barrier)

	time.Sleep(100 * time.Second)
}

func workAndWait(name string, timeToWork int, barrier *Barrier) {
	start := time.Now()
	for {
		fmt.Println(time.Since(start), name, "is running")
		time.Sleep(time.Duration(timeToWork) * time.Second)
		fmt.Println(time.Since(start), name, "is waiting on barrier")
		barrier.Wait()
	}
}

type Barrier struct {
	size      int
	waitCount int
	cond      *sync.Cond
}

func NewBarrier(size int) *Barrier {
	return &Barrier{size, 0, sync.NewCond(&sync.Mutex{})}
}

func (b *Barrier) Wait() {
	b.cond.L.Lock()
	b.waitCount += 1

	if b.waitCount == b.size {
		b.waitCount = 0
		b.cond.Broadcast()
	} else {
		b.cond.Wait()
	}
	b.cond.L.Unlock()
}

func run_fileSearch_WaitGroupie2() {
	wg := NewWaitGroupie2()
	wg.Add(1)
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	projDir := filepath.Dir(filepath.Dir(exePath))
	fmt.Printf("Project's directory: %s\n", projDir)
	go fileSearch_WaitGroupie2(projDir, "main", wg)
	wg.Wait()
}

func fileSearch_WaitGroupie2(dir, filename string, wg *WaitGroupie2) {
	files, _ := os.ReadDir(dir)
	for _, file := range files {

		fpath := filepath.Join(dir, file.Name())

		if strings.Contains(file.Name(), filename) {
			fmt.Println(fpath)
		}
		if file.IsDir() {
			wg.Add(1)
			go fileSearch_WaitGroupie2(fpath, filename, wg)
		}
	}
	wg.Done()
}

type WaitGroupie2 struct {
	groupSize int
	cond      *sync.Cond
}

func NewWaitGroupie2() *WaitGroupie2 {
	return &WaitGroupie2{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (wg *WaitGroupie2) Add(delta int) {
	wg.cond.L.Lock()
	wg.groupSize += delta
	wg.cond.L.Unlock()
}

func (wg *WaitGroupie2) Wait() {
	wg.cond.L.Lock()
	for wg.groupSize > 0 {
		wg.cond.Wait()
	}
	wg.cond.L.Unlock()
}

func (wg *WaitGroupie2) Done() {
	wg.cond.L.Lock()
	wg.groupSize--
	if wg.groupSize == 0 {
		wg.cond.Broadcast()
	}
	wg.cond.L.Unlock()
}

func run_fileSearch() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	projDir := filepath.Dir(filepath.Dir(exePath))
	fmt.Printf("Project's directory: %s\n", projDir)
	go fileSearch(projDir, "main", &wg)
	wg.Wait()
}

func fileSearch(dir, filename string, wg *sync.WaitGroup) {
	files, _ := os.ReadDir(dir)
	for _, file := range files {

		fpath := filepath.Join(dir, file.Name())

		if strings.Contains(file.Name(), filename) {
			fmt.Println(fpath)
		}
		if file.IsDir() {
			wg.Add(1)
			go fileSearch(fpath, filename, wg)
		}
	}
	wg.Done()
}

func run_WaitGroupie() {
	count := 40
	wg := NewWaitGroupie(count)
	for i := range count {
		go doWork_WaitGroupie(i, wg)
	}
	wg.Wait()
	fmt.Println("All complete")
}

func doWork_WaitGroupie(id int, wg *WaitGroupie) {
	fmt.Println(id, "Done working")
	wg.Done()
}

type WaitGroupie struct {
	sema *Semaphore
}

func NewWaitGroupie(size int) *WaitGroupie {
	return &WaitGroupie{sema: NewSemaphore(1 - size)}
}

func (wg *WaitGroupie) Wait() {
	wg.sema.Acquire()
}

func (wg *WaitGroupie) Done() {
	wg.sema.Release()
}

func run_countLetters_waitGroup() {
	wg := sync.WaitGroup{}
	count := 31
	wg.Add(count)
	mutex := sync.Mutex{}
	freq := make([]int, 26)

	for i := 1000; i < 1000+31; i++ {
		url := fmt.Sprintf("https://rfc-editor.org/rfc/rfc%d.txt", i)

		go func() {
			countLetters(url, freq, &mutex)
			wg.Done()
		}()
	}

	wg.Wait()

	mutex.Lock()
	for i, c := range all_letters {
		fmt.Printf("%c: %d\n", c, freq[i])
	}
	mutex.Unlock()
}

func run_waitGroup() {
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := range 10 {
		go doWork_waitGroup(i, &wg)
	}
	wg.Wait()
	fmt.Println("All complete")
}

func doWork_waitGroup(id int, wg *sync.WaitGroup) {
	i := rand.IntN(10000)
	time.Sleep(time.Duration(i) * time.Millisecond)

	fmt.Println(id, "Done working after", float64(i)/1000, "seconds")
	wg.Done()
}

func run_semaphore() {
	semaphore := NewSemaphore(0)
	for range 50000 {
		go doWork_semaphore(semaphore)
		fmt.Println("Waiting for child goroutine")
		semaphore.Acquire()
		fmt.Printf("Child goroutine finished\n\n")
	}
}

func doWork_semaphore(semaphore *Semaphore) {
	fmt.Println("Work started")
	fmt.Println("Work finished")
	semaphore.Release()
}

type Semaphore struct {
	permits int
	cond    *sync.Cond
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{
		permits: n,
		cond:    sync.NewCond(&sync.Mutex{}),
	}
}

func (rw *Semaphore) Acquire() {
	rw.cond.L.Lock()
	for rw.permits <= 0 {
		rw.cond.Wait()
	}
	rw.permits--
	rw.cond.L.Unlock()
}

func (rw *Semaphore) Release() {
	rw.cond.L.Lock()

	rw.permits++
	rw.cond.Signal()

	rw.cond.L.Unlock()
}

func run_write_preffered_locking() {
	rwMutex := NewReadWriteMutex()
	for range 2 {
		go func() {
			for {
				rwMutex.ReadLock()
				time.Sleep(1 * time.Second)
				fmt.Println("Read done")
				rwMutex.ReadUnlock()
			}
		}()
	}

	time.Sleep(5 * time.Second)
	rwMutex.WriteLock()
	fmt.Println("Write finished")
}

func run_read_preffered_locking() {
	rwMutex := ReadWriteMutex_reader_preferred{}
	for range 2 {
		go func() {
			for {
				rwMutex.ReadLock()
				time.Sleep(1 * time.Second)
				fmt.Println("Read done")
				rwMutex.ReadUnlock()
			}
		}()
	}

	time.Sleep(1 * time.Second)
	rwMutex.WriteLock()
	fmt.Println("Write finished")
}

// write preferring mutex using conditions
type ReadWriteMutex struct {
	readersCounter int
	writersWaiting int
	writerActive   bool
	cond           *sync.Cond
}

func NewReadWriteMutex() *ReadWriteMutex {
	return &ReadWriteMutex{cond: sync.NewCond(&sync.Mutex{})}
}

func (rw *ReadWriteMutex) ReadLock() {
	rw.cond.L.Lock()
	for rw.writersWaiting > 0 || rw.writerActive {
		rw.cond.Wait()
	}
	rw.readersCounter++
	rw.cond.L.Unlock()
}

func (rw *ReadWriteMutex) WriteLock() {
	rw.cond.L.Lock()

	rw.writersWaiting++
	for rw.readersCounter > 0 || rw.writerActive {
		rw.cond.Wait()
	}
	rw.writersWaiting--
	rw.writerActive = true

	rw.cond.L.Unlock()
}

func (rw *ReadWriteMutex) ReadUnlock() {
	rw.cond.L.Lock()
	rw.readersCounter--
	if rw.readersCounter == 0 {
		rw.cond.Broadcast()
	}
	rw.cond.L.Unlock()
}

func (rw *ReadWriteMutex) WriteUnlock() {
	rw.cond.L.Lock()
	rw.writerActive = false
	rw.cond.Broadcast()
	rw.cond.L.Unlock()
}

type ReadWriteMutex_reader_preferred struct {
	readersCounter int
	readersLock    sync.Mutex
	globalLock     sync.Mutex
}

func (rw *ReadWriteMutex_reader_preferred) ReadLock() {
	rw.readersLock.Lock()
	rw.readersCounter++
	if rw.readersCounter == 1 {
		rw.globalLock.Lock()
	}
	rw.readersLock.Unlock()
}

func (rw *ReadWriteMutex_reader_preferred) WriteLock() {
	rw.globalLock.Lock()
}
func (rw *ReadWriteMutex_reader_preferred) ReadUnlock() {
	rw.readersLock.Lock()
	rw.readersCounter--
	if rw.readersCounter == 0 {
		rw.globalLock.Unlock()
	}
	rw.readersLock.Unlock()
}

func (rw *ReadWriteMutex_reader_preferred) WriteUnlock() {
	rw.globalLock.Unlock()
}

func run_multiplayer() {
	cond := sync.NewCond(&sync.Mutex{})
	playersInGame := 4
	for playerID := range playersInGame {
		go playerHandler(cond, &playersInGame, playerID)
		time.Sleep(1 * time.Second)
	}
}

func playerHandler(cond *sync.Cond, playersRemaining *int, playerId int) {
	cond.L.Lock()
	fmt.Println(playerId, ": Connected")
	*playersRemaining--
	if *playersRemaining == 0 {
		cond.Broadcast()
	}
	for *playersRemaining > 0 {
		fmt.Println(playerId, ": Waiting for more players")
		cond.Wait()
	}
	cond.L.Unlock()
	fmt.Println("All players connected. Ready player", playerId)
	//Game started
}

func run_doWork() {
	cond := sync.NewCond(&sync.Mutex{})
	cond.L.Lock()
	for range 5000 {
		go doWork(cond)
		fmt.Println("Waiting for child goroutine")
		cond.Wait()
		fmt.Println("Child goroutine finished")
	}
	cond.L.Unlock()
}
func doWork(cond *sync.Cond) {
	fmt.Println("Work started")
	fmt.Println("Work finished")

	cond.L.Lock()
	cond.Signal()
	cond.L.Unlock()
}

func run_matchRecorder() {
	mutex := sync.RWMutex{}
	matchEvents := make([]string, 0, 10000)
	for range 10000 {
		matchEvents = append(matchEvents, "Match event")
	}
	go matchRecorder(&matchEvents, &mutex)
	start := time.Now()
	for range 5000 {
		go matchClient(&matchEvents, &mutex, start)
	}
	time.Sleep(100 * time.Second)
}

func matchRecorder(matchEvents *[]string, mutex *sync.RWMutex) {
	for i := 0; ; i++ {
		mutex.Lock()
		*matchEvents = append(*matchEvents, "Match Event "+strconv.Itoa(i))
		mutex.Unlock()
		time.Sleep(200 * time.Millisecond)
		// fmt.Println("Appended match event")
	}
}

func matchClient(mEvents *[]string, mutex *sync.RWMutex, st time.Time) {
	for range 100 {
		mutex.RLock()
		allEvents := copyAllEvents(mEvents)
		mutex.RUnlock()

		timeTaken := time.Since(st)
		fmt.Println(len(allEvents), "events copied in", timeTaken)
	}
}

func copyAllEvents(matchEvents *[]string) []string {
	allEvents := make([]string, 0, len(*matchEvents))
	for _, e := range *matchEvents {
		allEvents = append(allEvents, e)
	}
	return allEvents
}

func run_stingy_and_spendy() {
	money := 100
	mutex := sync.Mutex{}
	cond := sync.NewCond(&mutex)
	go stingy(&money, cond)
	go spendy(&money, cond)

	time.Sleep(2 * time.Second)

	mutex.Lock()
	fmt.Println("Money in bank account: ", money)
	mutex.Unlock()

}

func stingy(money *int, cond *sync.Cond) {
	for range 1000000 {
		cond.L.Lock()
		*money += 10
		cond.Signal()
		cond.L.Unlock()
	}
	fmt.Println("Stingy Done")
}
func spendy(money *int, cond *sync.Cond) {
	for range 200000 {
		cond.L.Lock()
		for *money < 50 {
			cond.Wait()
		}
		*money -= 50
		if *money < 0 {
			fmt.Println("Money is negative!")
			os.Exit(1)
		}
		cond.L.Unlock()
	}
	fmt.Println("Spendy Done")
}

func run_count_letters() {
	freq := make([]int, 26)

	mutex := sync.Mutex{}

	for i := 1000; i <= 1030; i++ {
		url := fmt.Sprintf("https://rfc-editor.org/rfc/rfc%d.txt", i)
		go countLetters(url, freq, &mutex)
	}

	time.Sleep(1 * time.Second)

	mutex.Lock()
	for i, c := range all_letters {
		fmt.Printf("%c:%d\n", c, freq[i])
	}
	mutex.Unlock()

}
func run_count_letters_2() {
	freq := make([]int, 26)

	mutex := sync.Mutex{}

	for i := 1000; i <= 2200; i++ {
		url := fmt.Sprintf("https://rfc-editor.org/rfc/rfc%d.txt", i)
		go countLetters(url, freq, &mutex)
	}

	for range 100 {
		time.Sleep(100 * time.Millisecond)
		if mutex.TryLock() {

			for i, c := range all_letters {
				fmt.Printf("%c:%d\n", c, freq[i])
			}
			fmt.Printf("\n\n\n\n")
			mutex.Unlock()
		} else {
			fmt.Println("Mutex already being used")
		}
	}

}

func countLetters(url string, freq []int, mutex *sync.Mutex) {
	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Println("Server return error status code: " + res.Status)
		return
	}

	body, _ := io.ReadAll(res.Body)
	mutex.Lock()
	for _, b := range body {
		c := strings.ToLower(string(b))
		cIndex := strings.Index(all_letters, c)
		if cIndex >= 0 {
			freq[cIndex] += 1
		}
	}
	mutex.Unlock()

	fmt.Println("Completed:", url)

}
