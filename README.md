# Git Sentinel

## Overview

A simple tool that pulls detailed `git log` statistics from a list of Git repositories and generates a report.

```sh
SENTINEL_DATA_DIR=./data SENTINEL_DB_URL="postgres://postgres:docker@localhost/sentinel?sslmode=disable" ./git-sentinel
2019/05/19 20:45:02 Sentinel - A Git log analyzer v1.0.%BUILD_ID% Starting...
2019/05/19 20:45:02 Loading repository definitions file from 'sentinel.yaml'...
2019/05/19 20:45:02 Preparing scratch directory './data'
2019/05/19 20:45:02 Connecting to database...
2019/05/19 20:45:04 [etl.googleanalytics] Processing repository...
2019/05/19 20:45:04 [etl.googleanalytics] Working directory is data/etl.googleanalytics.git
2019/05/19 20:45:04 [etl.googleanalytics] Repository does not exist, cloning...
2019/05/19 20:45:05 [etl.googleanalytics] Determining last updated date...
2019/05/19 20:45:05 [etl.googleanalytics] No records found, grabbing the full history
2019/05/19 20:45:05 [etl.googleanalytics] Scanning repository history...
2019/05/19 20:45:05 [etl.googleanalytics] Scan complete, 143 new entries will be saved
2019/05/19 20:45:39 [etl.googleanalytics] Finished processing repository
```

## Configuration

The format of the input `sentinel.yaml` file should follow this scheme:

```yaml
- name: My Repo
  dir: my-repo
  url: https://github.com/me/my-repo
```

## Database

The application will automatically create the relevant database tables, but the database `sentinel` must be pre-created:

```postgres
postgres=# create database sentinel;
```

## Get a List of Repos from Azure DevOps

A convenience script `get_repos.sh` is included which retrives all the Git repos from Azure DevOps and produces a `sentinel.yaml` file. It requires inputing the `ORG`, `PROJECT` and `PAT`.

1. Edit `get_repos.sh` and input your project details including your Personal Access Token (PAT)
1. Run the script: `$ ./get_repos.sh`

## Requirements

- **PostgreSQL**: The tool has been refactored to use a PostgreSQL backend in order to facilitate integration with various BI tools.
- **Storage**: Sufficient capacity to store all the repositories
