package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/kelseyhightower/envconfig"
)

var opt struct {
	LogLevel string `default:"INFO" split_words:"true"`
	RepoList string
	DbURL    string `default:"mongodb://root:root@mongodb:27017/sentinel?authSource=admin" split_words:"true"`
}

// Repo represents a Git repository object composed of a name and a URL
type Repo struct {
	Name    string
	URL     string
	Commits []Commit
}

// Commit is a Git commit amended with lines added & deleted
type Commit struct {
	Repo       string `json:"repo"`
	Author     string `json:"author"`
	Date       string `json:"date"`
	Title      string `json:"title"`
	Hash       string `json:"hash"`
	Ref        string `json:"ref"`
	Insertions int
	Deletions  int
}

func init() {

	err := envconfig.Process("Sentinel", &opt)
	if err != nil {
		log.Printf(err.Error())
		return
	}
}

func (r *Repo) parse() error {

	cmd := exec.Command("git", "log", "--all", "--shortstat", "--pretty=format:{\"author\":\"%aE\",\"date\":\"%aI\",\"title\":\"%f\",\"hash\":\"%h\",\"ref\":\"%D\"}")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	ins := regexp.MustCompile(`(?P<Insertions>\d+) insertions\(\+\)`)
	del := regexp.MustCompile(`(?P<Deletions>\d+) deletions\(\-\)`)

	scanner := bufio.NewScanner(stdout)

	var c Commit

	for scanner.Scan() {
		line := scanner.Text()

		err := json.Unmarshal([]byte(line), &c)
		if err == nil {
			r.Commits = append(r.Commits, Commit{Repo: r.Name, Author: c.Author, Date: c.Date, Title: c.Title, Hash: c.Hash, Ref: c.Ref})
			continue
		}

		if match := ins.FindStringSubmatch(line); len(match) != 0 {
			i, _ := strconv.Atoi(match[1])
			r.Commits[len(r.Commits)-1].Insertions = i
		}

		if match := del.FindStringSubmatch(line); len(match) != 0 {
			i, _ := strconv.Atoi(match[1])
			r.Commits[len(r.Commits)-1].Deletions = i
		}

	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func main() {
	log.Printf("Sentinel - A Git log analyzer v1.0.%%BUILD_ID%% Starting...")

	for _, r := range []Repo{Repo{Name: "oh-my-zsh", URL: "", Commits: make([]Commit, 0)}} {
		err := r.parse()
		if err == nil {
			for _, c := range r.Commits {
				j, _ := json.Marshal(c)
				fmt.Printf("%s\n", j)
			}
		}
	}

}
