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
const ledger = "LEDGER.txt"

type attempt struct {
	sha1 string
	body string
}

func main() {
	tokens := make(chan string)
	attempts := make(chan attempt)

	go func() {
		for {
			tokens <- uniuri.New()
		}
	}()

	go func() {
		for {
			attempts <- generateAttempt(tokens)
		}
	}()

	difficulty := calculateDifficulty()
	update_ledger()

	for {
		go func() {
			for {
				if solve(difficulty, attempts) {
					break
				}
			}
		}()

		if push() {
			log.Println("Successfully mined and pushed!")
			break
		} else {
			reset()
		}
	}

}

func update_ledger() {
	data, err := ioutil.ReadFile(ledger)
	if err != nil {
		log.Fatalln(err)
	}

	if !strings.Contains(string(data), username) {
		data = []byte(fmt.Sprintf("%s%s: 1", string(data), username))
	}

	ioutil.WriteFile(ledger, data, 0644)

	exec.Command("git", "add", ledger).Run()
}

func push() bool {
	output, _ := exec.Command("git", "push", "origin", "master").CombinedOutput()
	fmt.Println(string(output))

	cmd := exec.Command("git", "push", "origin", "master")

	output, _ = cmd.CombinedOutput()

	fmt.Println(string(output))

	return cmd.ProcessState.Success()
}

func reset() {
	log.Println("Resetting!")

	exec.Command("git", "fetch", "origin", "master").Run()
	exec.Command("git", "reset", "--hard", "origin/master").Run()
}

func calculateDifficulty() string {
	difficulty, err := ioutil.ReadFile("difficulty.txt")
	if err != nil {
		log.Fatalln(err)
	}

	return string(difficulty)
}

func solve(difficulty string, attempts chan attempt) bool {
	fmt.Print(".")

	attempt := <-attempts

	if bytes.Compare([]byte(attempt.sha1), []byte(difficulty)) < 0 {
		fmt.Println(string(attempt.sha1))
		cmd := exec.Command("git", "hash-object", "-t", "commit", "--stdin", "-w")
		cmd.Stdin = strings.NewReader(attempt.body)
		cmd.Run()
		exec.Command("git", "reset", "--hard", attempt.sha1).Run()
		return true
	}

	return false
}

func generateAttempt(tokens chan string) attempt {
	tree, _ := exec.Command("git", "write-tree").Output()

	parent, _ := exec.Command("git", "rev-parse", "HEAD").Output()

	timestamp := time.Now().Unix()

	hasher := sha1.New()

	random := <-tokens

	content := fmt.Sprintf("tree %sparent %sauthor %s <itsmeduncan@gmail.com> %v +0000\ncommitter %s <itsmeduncan@gmail.com> %v +0000\n\nGive me a Gitcoin\n\n%s", string(tree), string(parent), username, timestamp, username, timestamp, random)

	head := fmt.Sprintf("commit %d", len(content))
	body := fmt.Sprintf("%s%s", append([]byte(head), *new(byte)), content)

	io.WriteString(hasher, body)
	sha1 := fmt.Sprintf("%x", hasher.Sum(nil))

	fmt.Println(sha1)

	return attempt{sha1: sha1, body: content}
}
