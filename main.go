package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const all_letters = "abcdefghijklmnopqrstuvwxyz"

func main() {
	fmt.Println("\n\nStarting...")
	time.Sleep(2 * time.Second)
	// run_count_letters()
	// run_stingy_and_spendy()
	// run_matchRecorder()
	// time.Sleep(101 * time.Second)
	// run_doWork()
	run_multiplayer()
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
