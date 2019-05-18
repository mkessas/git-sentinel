package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/9spokes/go/db"
	"github.com/globalsign/mgo/bson"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

var mongo db.MongoDB
var repos []Repo

var opt struct {
	LogLevel string `default:"INFO" split_words:"true"`
	RepoList string `default:"sentinel.yaml" slit_words:"true"`
	DataDir  string `default:"/data" split_words:"true"`
	DbURL    string `default:"mongodb://root:root@mongodb:27017/sentinel?authSource=admin" split_words:"true"`
}

// Repo represents a Git repository object composed of a name and a URL
type Repo struct {
	Name    string
	Dir     string
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

func (r *Repo) save() error {

	database := mongo.Session.DB("")

	col := database.C("commits")

	for _, c := range r.Commits {
		r := bson.M{
			"repo":   c.Repo,
			"author": c.Author,
			//"date":       time.d (c.Date)
			"title":      c.Title,
			"hash":       c.Hash,
			"ref":        c.Ref,
			"insertions": c.Insertions,
			"deletions":  c.Deletions,
		}
		_, err := col.Upsert(r, r)
		if err != nil {
			log.Printf("error inserting row: %s", err.Error())
		}
	}

	return nil
}

func (r *Repo) sync() error {

	fullPath := path.Join(opt.DataDir, r.Dir)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "git", "clone", r.URL, "--bare", "-c", r.Dir)
		cmd.Dir = opt.DataDir
		_, err := cmd.Output()
		if err != nil {
			log.Panicf("error executing command: %s", err.Error())
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "git", "fetch", "--all")
		cmd.Dir = fullPath
		_, err := cmd.Output()
		if err != nil {
			log.Panicf("error executing command: %s", err.Error())
		}
	}
	return nil
}

func (r *Repo) parse() error {

	cmd := exec.Command("git", "log", "--all", "--shortstat", "--pretty=format:{\"author\":\"%aE\",\"date\":\"%aI\",\"title\":\"%f\",\"hash\":\"%h\",\"ref\":\"%D\"}")
	cmd.Dir = path.Join(opt.DataDir, r.Dir)
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

func dbConnect() {

	var err error

	mongo, err = db.Connect(opt.DbURL)
	if err != nil {
		log.Panicf("Failed to connect to database: " + err.Error())
	}
}

func loadRepos() {

	dat, err := ioutil.ReadFile(opt.RepoList)
	if err != nil {
		log.Printf("Failed to read repository definition: " + err.Error())
		return
	}
	err = yaml.Unmarshal(dat, &repos)
	if err != nil {
		log.Printf("Failed to parse configuration file: %s", err.Error())
	}
}

func prepDataDir() {
	if _, err := os.Stat(opt.DataDir); os.IsNotExist(err) {
		err := os.MkdirAll(opt.DataDir, 0700)
		if err != nil {
			log.Panicf("Failed to create data directory: %s", err.Error())
		}
	}
}

func main() {

	log.Printf("Sentinel - A Git log analyzer v1.0.%%BUILD_ID%% Starting...")

	log.Printf("Loading repository definitions file from '%s'...", opt.RepoList)
	loadRepos()

	log.Printf("Preparing scratch directory '%s'", opt.DataDir)
	prepDataDir()

	log.Printf("Connecting to database on '%s'...", opt.DbURL)
	dbConnect()

	for _, r := range repos {

		log.Printf("Processing repository '%s'...", r.Name)
		r.Dir = path.Base(r.URL) + ".git"
		log.Printf("Working directory is %s", path.Join(opt.DataDir, r.Dir))
		err := r.sync()
		if err != nil {
			log.Printf("Failed to process repository '%s': %s", r.Name, err.Error())
			continue
		}

		err = r.parse()
		if err != nil {
			log.Printf("Failed to process repository '%s': %s", r.Name, err.Error())
			continue
		}

		err = r.save()
		if err != nil {
			log.Printf("Failed to save stats to database: %s", err.Error())
		}

	}
}
