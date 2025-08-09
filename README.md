# Neural-Theft


![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white&style=for-the-badge) ![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite&logoColor=white&style=for-the-badge) ![Concurrency](https://img.shields.io/badge/Concurrency-Enabled-brightgreen?style=for-the-badge) ![Scraper](https://img.shields.io/badge/Webscraper-Tracker.gg-purple?style=for-the-badge) ![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)

---

**Neural-Theft** is a high-performance, concurrent data pipeline written in Go that scrapes player stats from [Tracker.gg](https://tracker.gg) and stores them in a lightweight SQLite database.

- **Concurrency & Parallelism**: Leveraging Go’s goroutines, channels, and rate limiting for efficient, safe, and scalable web scraping.
- **Modular Architecture**: Clear separation of concerns between configuration, scraping, data processing, and persistence.
- **Robust Data Pipeline**: From HTML parsing to API integration and database transactions, with error handling and context-based cancellation.
- **Testability & Maintenance**: Easily extendable components with unit-friendly interfaces and minimal external dependencies.

---

## 🚀 Key Features

- **Concurrent Scraping Engine**: Utilizes the `colly` library alongside Go’s concurrency primitives to fetch and parse hundreds of pages per minute without blocking or data races.
- **Rate-Limited API Integration**: Safe, token-authenticated calls to external APIs for UUID resolution, governed by `golang.org/x/time/rate` to respect service limits.
- **Transactional Persistence**: Batch commits to SQLite3 via `modernc.org/sqlite` with `INSERT OR IGNORE` semantics for idempotent inserts and efficient rollback.
- **Context-Aware Cancellation**: Graceful shutdown of all goroutines on user interrupt or error, ensuring no partial writes or resource leaks.
- **Verbose Logging**: Optional debug mode for real-time insights into scraping progress, API failures, and database operations.

---


## 🛠 Installation
Make sure you have Go installed.

```bash
git clone https://github.com/washiil/neural-theft.git
cd neural-theft
go build -o neural-theft
```


```bash
cp .env.example .env
# Add your API_KEY in .env
```

```bash
./neural-theft \
  -region na \
  -output players.db \
  -speed 100
```