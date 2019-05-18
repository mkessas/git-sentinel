package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/kelseyhightower/envconfig"
)

var opt struct {
	LogLevel string `default:"INFO" split_words:"true"`
	RepoList string
	DbURL    string `default:"mongodb://root:root@mongodb:27017/sentinel?authSource=admin" split_words:"true"`
}

type repo struct {
	Name string
	URL  string
}

type commit struct {
	Author string `json:"author"`
	Date   string `json:"date"`
	Title  string `json:"title"`
	Hash   string `json:"hash"`
	Ref    string `json:"ref"`
}

func init() {

	err := envconfig.Process("Sentinel", &opt)
	if err != nil {
		log.Printf(err.Error())
		return
	}
}

func (r repo) parse() {

	cmd := exec.Command("git", "log", "--all", "--pretty=format:{\"author\":\"%aE\",\"date\":\"%aI\",\"title\":\"%f\",\"hash\":\"%h\",\"ref\":\"%D\"}")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// buf := new(bytes.Buffer)
	// buf.ReadFrom(stdout)

	var c commit
	if err := json.NewDecoder(stdout).Decode(&c); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s commit %s on %s with the comment %s\n", c.Author, c.Hash, c.Date, c.Title)
	// fmt.Printf("%s\n", buf)

}

func main() {
	log.Printf("Sentinel - A Git log analyzer v1.0.%%BUILD_ID%% Starting...")

	for _, r := range []repo{repo{Name: "oh-my-zsh", URL: ""}} {
		r.parse()
	}

}
