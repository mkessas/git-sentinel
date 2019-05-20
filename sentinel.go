package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"

	_ "github.com/lib/pq"
)

var repos []Repo
var db *sql.DB

var opt struct {
	LogLevel string `default:"INFO" split_words:"true"`
	RepoList string `default:"sentinel.yaml" split_words:"true"`
	DataDir  string `default:"/data" split_words:"true"`
	DbURL    string `default:"postgres://postgres:docker@localhost/sentinel?sslmode=disable" split_words:"true"`
}

// Repo represents a Git repository object composed of a name and a URL
type Repo struct {
	Name        string
	Dir         string
	URL         string
	LastUpdated int64
	Commits     []Commit
}

// Commit is a Git commit amended with lines added & deleted
type Commit struct {
	Repo       string `json:"repo"`
	Author     string `json:"author"`
	Date       int64  `json:"date"`
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

	for _, c := range r.Commits {
		_, err := db.Exec("INSERT INTO commits(hash, repo, author, date, title, ref, additions, deletions) VALUES($1,$2,$3,$4,$5,$6,$7,$8)",
			c.Hash, c.Repo, c.Author, c.Date, c.Title, c.Ref, c.Insertions, c.Deletions)

		if err != nil {
			log.Printf("[%s] error inserting row: %s", r.Name, err.Error())
		}
	}

	return nil
}

func (r *Repo) load() {

	var date int64

	err := db.QueryRow("SELECT date FROM commits WHERE repo = $1 ORDER BY date DESC LIMIT 1", r.Name).Scan(&date)
	switch {
	case err == sql.ErrNoRows:
		r.LastUpdated = 0
	case err != nil:
		fmt.Printf("[%s] Failed to execute query: %s", r.Name, err.Error())
	default:
		r.LastUpdated = date
	}
}

func (r *Repo) sync() error {

	fullPath := path.Join(opt.DataDir, r.Dir)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {

		log.Printf("[%s] Repository does not exist, cloning...", r.Name)

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "git", "clone", r.URL, "--bare", "-c", r.Dir)
		cmd.Dir = opt.DataDir
		_, err := cmd.Output()
		if err != nil {
			return err
		}
	} else {

		log.Printf("[%s] Repository exists, fetching...", r.Name)

		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "git", "fetch", "origin", "+refs/*:refs/*")
		cmd.Dir = fullPath
		_, err := cmd.Output()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) parse() error {

	last := "--since=\"5 years ago\""
	if r.LastUpdated > 0 {
		last = fmt.Sprintf("--since=%d", r.LastUpdated+1)
	}

	cmd := exec.Command("git", "log", "--all", "--shortstat", last, "--pretty=format:{\"author\":\"%aE\",\"date\":%ct,\"title\":\"%f\",\"hash\":\"%h\",\"ref\":\"%D\"}")
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

func dbConnect() error {

	var err error
	db, err = sql.Open("postgres", opt.DbURL)
	if err != nil {
		return fmt.Errorf("Failed to connect to database '%s': %s", opt.DbURL, err.Error())
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS commits (hash VARCHAR(12) NOT NULL PRIMARY KEY, repo VARCHAR(128), author VARCHAR(256), date BIGINT, title VARCHAR(256), ref VARCHAR(256),additions BIGINT, deletions BIGINT)")
	if err != nil {
		log.Printf("failed to create table: %v\n", err.Error())
	}
	for _, i := range []string{"date", "author", "repo"} {
		db.Exec(fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", i, "commits", i))
	}
	return nil
}

func loadRepos() error {

	dat, err := ioutil.ReadFile(opt.RepoList)
	if err != nil {
		return fmt.Errorf("Failed to read repository definition: %s", err.Error())
	}
	err = yaml.Unmarshal(dat, &repos)
	if err != nil {
		return fmt.Errorf("Failed to parse configuration file: %s", err.Error())
	}
	return nil
}

func prepDataDir() error {
	if _, err := os.Stat(opt.DataDir); os.IsNotExist(err) {
		err := os.MkdirAll(opt.DataDir, 0700)
		if err != nil {
			return fmt.Errorf("Failed to create data directory: %s", err.Error())
		}
	}
	return nil
}

func main() {

	log.Printf("Sentinel - A Git log analyzer v1.0.%%BUILD_ID%% Starting...")

	log.Printf("Loading repository definitions file from '%s'...", opt.RepoList)
	if err := loadRepos(); err != nil {
		log.Printf("%s", err.Error())
		return
	}

	log.Printf("Preparing scratch directory '%s'", opt.DataDir)
	if err := prepDataDir(); err != nil {
		log.Printf("%s", err.Error())
		return
	}

	log.Printf("Connecting to database...")
	if err := dbConnect(); err != nil {
		log.Printf("%s", err.Error())
		return
	}

	for _, r := range repos {

		log.Printf("[%s] Processing repository...", r.Name)
		r.Dir = path.Base(r.URL) + ".git"
		log.Printf("[%s] Working directory is %s", r.Name, path.Join(opt.DataDir, r.Dir))
		err := r.sync()
		if err != nil {
			log.Printf("[%s] Failed to process repository: %s", r.Name, err.Error())
			continue
		}

		log.Printf("[%s] Determining last updated date...", r.Name)
		r.load()
		if r.LastUpdated == 0 {
			log.Printf("[%s] No records found, grabbing the full history", r.Name)
		} else {
			log.Printf("[%s] Last updated on '%s'", r.Name, time.Unix(r.LastUpdated, 0))
		}

		log.Printf("[%s] Scanning repository history...", r.Name)
		err = r.parse()
		if err != nil {
			log.Printf("[%s] Failed to parse repository logs: %s", r.Name, err.Error())
			continue
		}

		log.Printf("[%s] Scan complete, %d new entries will be saved", r.Name, len(r.Commits))
		err = r.save()
		if err != nil {
			log.Printf("[%s] Failed to save stats to database: %s", r.Name, err.Error())
		}

		log.Printf("[%s] Finished processing repository", r.Name)
	}
}
