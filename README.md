# Neural-Theft


![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white&style=for-the-badge) ![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite&logoColor=white&style=for-the-badge) ![Concurrency](https://img.shields.io/badge/Concurrency-Enabled-brightgreen?style=for-the-badge) ![Scraper](https://img.shields.io/badge/Webscraper-Tracker.gg-purple?style=for-the-badge) ![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)

---

Neural-Theft is a **Go-powered**, **concurrent web scraper** that digs through [Tracker.gg](https://tracker.gg) for game stats and stashes them in a **SQLite3 database**.  

Concurrent. Elegant.  

---

## 🚀 Features
- **Blazing fast scraping** with Go’s goroutines & channels.
- **Lightweight persistence** via SQLite3 – no complex DB setup.
- **Modular design** – extend or repurpose the scraping engine easily.
- **CLI-based** – no nonsense.

---

## 🛠 Installation
Make sure you have Go installed.

```bash
git clone https://github.com/washiil/neural-theft.git
cd neural-theft
go build -o neural-theft
