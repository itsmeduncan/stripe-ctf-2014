package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/dchest/uniuri"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"time"
)

const username = "user-5tr2yruv"

type attempt struct {
	sha1 string
	body string
}

func main() {
	tokens := make(chan string, 500)
	attempts := make(chan attempt)
	successes := make(chan attempt)

	difficulty, _ := ioutil.ReadFile("difficulty.txt")

	reset()

	tree, _ := exec.Command("git", "write-tree").Output()
	parent, _ := exec.Command("git", "rev-parse", "HEAD").Output()
	timestamp := time.Now().Unix()

	go func() {
		for {
			tokens <- uniuri.New()
		}
	}()

	go func() {
		for {
			attempts <- newAttempt(tokens, tree, parent, timestamp)
		}
	}()

	runner := func() {
		for {
			solve(attempts, successes, difficulty)
		}
	}

	for i := 0; i < 4; i++ {
		go runner()
	}

	for {
		attempt := <-successes
		if commit(attempt) {
			log.Println("Mined a coin: ", attempt.sha1)
			break
		} else {
			log.Println("Resetting!")
			reset()
		}
	}
}

func reset() {
	exec.Command("git", "reset", "HEAD").Run()
	exec.Command("git", "fetch", "origin", "master").Run()
	exec.Command("git", "reset", "--hard", "origin/master").Run()

	ledger := "LEDGER.txt"

	data, err := ioutil.ReadFile(ledger)
	if err != nil {
		log.Fatalln(err)
	}

	if !strings.Contains(string(data), username) {
		data = []byte(fmt.Sprintf("%s%s: 1\n", string(data), username))
	}

	ioutil.WriteFile(ledger, data, 0644)

	exec.Command("git", "add", ledger).Run()
}

func commit(success attempt) bool {
	commit := exec.Command("git", "hash-object", "-t", "commit", "--stdin", "-w")
	commit.Stdin = strings.NewReader(success.body)
	sha1, _ := commit.CombinedOutput()

	exec.Command("git", "reset", "--hard", strings.TrimSpace(string(sha1))).Run()

	push := exec.Command("git", "push", "origin", "master")
	output, _ := push.CombinedOutput()

	fmt.Println("Body: ", success.body)
	fmt.Println("Push: ", string(output))

	return push.ProcessState.Success()
}

func solve(attempts chan attempt, successes chan attempt, difficulty []byte) {
	attempt := <-attempts
	if bytes.Compare([]byte(attempt.sha1), []byte(difficulty)) < 0 {
		successes <- attempt
	}
}

func newAttempt(tokens chan string, tree []byte, parent []byte, timestamp int64) attempt {
	random := <-tokens

	hasher := sha1.New()

	content := fmt.Sprintf("tree %sparent %sauthor %s <itsmeduncan@gmail.com> %v +0000\ncommitter %s <itsmeduncan@gmail.com> %v +0000\n\nGive me a Gitcoin\n\n%s", string(tree), string(parent), username, timestamp, username, timestamp, random)

	head := fmt.Sprintf("commit %d", len(content))
	body := fmt.Sprintf("%s%s", append([]byte(head), *new(byte)), content)

	io.WriteString(hasher, body)
	sha1 := fmt.Sprintf("%x", hasher.Sum(nil))

	return attempt{sha1: sha1, body: content}
}
