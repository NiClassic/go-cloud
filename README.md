# Go Cloud App

A Go cloud storage application with SQLite.

## Quick Start

```bash
git clone https://github.com/NiClassic/go-cloud.git
cd go-cloud
docker compose up --build
```

### Available Environment variables in docker-compose.yml

| Variable  | Example Value    | Effect                                      |
|-----------|------------------|---------------------------------------------|
| DATA_ROOT | /data            | Root storage path of all the uploaded files |
| DB_FILE   | /data/storage.db | Path of the SQLite database                 |
| DEBUG     | true             | Enable debug logs                           |
| TZ        | Europe/Berlin    | Set the local timezone for date formatting  |

This will:
- Run the database migrations
- Create or update the SQLite database
- Start the webapp [here](http://localhost:8080)
