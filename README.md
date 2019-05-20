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

## Requirements

- **PostgreSQL**:  The tool has been refactored to use a PostgreSQL backend in order to facilitate integration with various BI tools.
- **Storage**: Sufficient capacity to store all the repositories
