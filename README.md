# Go Cloud App

A Go cloud storage application with SQLite.

## Quick Start

```bash
git clone https://github.com/yourname/go-cloud-app.git
cd go-cloud-app
USERID=$(id -u) GROUPID=$(id -g) docker compose up --build
```

This will:
- Run the database migrations
- Create or update the SQLite database
- Start the webapp [here](http://localhost:8080)
