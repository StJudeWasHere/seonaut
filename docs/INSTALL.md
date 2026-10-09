# INSTALLATION GUIDE

This document provides instructions for installing and configuring SEOnaut. Follow the steps below to set up the application using Docker or by compiling it from source. Instructions for configuring a reverse proxy with HTTPS and WebSocket support are also included.

## Prerequisites

Before installing SEOnaut, ensure the following are installed on your system:

- **Docker**: Install Docker from [docker.com](https://www.docker.com/).
- **Git**: Required to clone the repository.
- **Go Programming Language** (optional, if not using Docker): Install Go from [golang.org](https://golang.org/).
- **Make** (optional): To use the Makefile commands provided in the project.
- **MySQL Database**: Install and configure a MySQL server (if not using Docker).
- **Nginx or Apache**: Required to set up a reverse proxy.

---

## Installation Options

### Using docker compose

Using docker is the recommended way of running SEOnaut. As you need to provide a database, you can use `docker compose` to do so creating a `docker-compose.yml` file like this:

```yml
services:
  db:
    image: mysql:8.4
    container_name: "SEOnaut-db"
    environment:
      - MYSQL_ROOT_PASSWORD=root
      - MYSQL_DATABASE=seonaut
      - MYSQL_USER=seonaut
      - MYSQL_PASSWORD=seonaut
    networks:
      - seonaut_network

  app:
    image: ghcr.io/stjudewashere/seonaut:latest
    container_name: "SEOnaut-app"
    ports:
      - "${SEONAUT_PORT:-9000}:9000"
    depends_on:
      - db
    command: sh -c "/bin/wait && /app/seonaut"
    environment:
      - WAIT_HOSTS=db:3306
      - WAIT_TIMEOUT=300
      - WAIT_SLEEP_INTERVAL=30
      - WAIT_HOST_CONNECT_TIMEOUT=30
      # Seonaut config overrides
      # - SEONAUT_SERVER_HOST=${SEONAUT_SERVER_HOST:-0.0.0.0}
      # - SEONAUT_SERVER_PORT=${SEONAUT_INTERNAL_PORT:-9000}
      # - SEONAUT_SERVER_URL=${SEONAUT_SERVER_URL:-http://localhost:${SEONAUT_PORT:-9000}}
      # - SEONAUT_DATABASE_SERVER=${SEONAUT_DB_SERVER:-db}
      # - SEONAUT_DATABASE_PORT=${SEONAUT_DB_PORT:-3306}
      # - SEONAUT_DATABASE_USER=${SEONAUT_DB_USER:-seonaut}
      # - SEONAUT_DATABASE_PASSWORD=${SEONAUT_DB_PASSWORD:-seonaut}
      # - SEONAUT_DATABASE_DATABASE=${SEONAUT_DB_NAME:-seonaut}
    networks:
      - seonaut_network

networks:
  seonaut_network:
    driver: bridge
```

This will pull the latest SEOnaut image and run it with the default settings. You can overwrite the default settings using environment variables if needed.

### Getting latest docker image

In case you want to pull the latest image to use it in a different setting, you can pull it from the registry:

```sh
$ docker pull ghcr.io/stjudewashere/seonaut:latest
```

### Using docker from source code

1. **Clone the Repository**  
   Clone the SEOnaut repository from GitHub:

   `git clone https://github.com/stjudewashere/seonaut.git`

2. **Navigate to the Project Directory**  
   Move to the project folder:

   `cd seonaut`

3. **Build and Run Docker Containers**  
   Run docker-compose to build and start the containers:

   `docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build`

   Or use the provided Makefile to build and start the Docker containers:

   `make docker`

4. **Access the Application**  
   Open your browser and navigate to:

   `http://localhost:9000`

   To secure SEOnaut with HTTPS, set up a reverse proxy as explained below.

---

### Compiling from Source

1. **Clone the Repository**  
   Clone the SEOnaut repository:

   `git clone https://github.com/stjudewashere/seonaut.git`

2. **Navigate to the Project Directory**  
   Change into the project directory:

   `cd seonaut`

3. **Run the Application**  
   Use the Makefile to run the application with the default configuration:

   `make run`

   To use a custom configuration file, specify it with the `-c` option:

   `go run -race cmd/server/main.go -c path/to/your/config`

4. **Compile the CSS**  
   To compile the CSS you'll need to install esbuild. Then you can compile the CSS:

```
   esbuild ./web/css/style.css \
      --bundle \
      --minify \
      --outdir=./web/static \
      --public-path=/resources \
      --loader:.woff=file \
      --loader:.woff2=file \
      --loader:.png=file
```

   Alternatively you can use `make front` to compile the CSS or `make watch` to compile and monitor CSS file changes while developing.

5. **Access the Application**  
   Navigate to:

   `http://localhost:9000`

---

## Configuration File

The application uses a configuration file named `config`, located in the root directory. Customize this file to match your environment or specify a custom file using the `-c` option.

### Default Configuration

    [server]
    host = "0.0.0.0"
    port = 9000
    url = "http://localhost:9000"

    [database]
    server = "db"
    port = 3306
    user = "seonaut"
    password = "seonaut"
    database = "seonaut"

    [crawler]
    agent = "Mozilla/5.0 (compatible; SEOnautBot/1.0; +https://seonaut.org/bot)"

    [scheduler]
    enabled = true

### Key Configuration Options

- **[server]**
  - `host`: IP address to bind the server (default: `0.0.0.0`).
  - `port`: Port for the server (default: `9000`).
  - `url`: Base URL of the application, e.g., `https://example.com`.

- **[database]**
  - `server`: Hostname or IP of the database server.
  - `port`: Port of the database (default: `3306`).
  - `user`: Database username.
  - `password`: Database password.
  - `database`: Name of the database.

- **[crawler]**
  - `agent`: User agent string for the crawler.

- **[scheduler]**
  - `enabled`: Enable the automatic crawl scheduler (default: `true`). Can be disabled with the `SEONAUT_SCHEDULER_ENABLED` environment variable set to `false`.

### Scheduled crawls

Each project can have a crawl schedule (hourly, daily or weekly), selected when the project is created or edited. While the server is running, a scheduler checks every minute for scheduled projects whose next run time has passed and starts a new crawl for each of them. The next run time is recalculated from the current time after every scheduled crawl starts, including when a crawl fails to start, so a failing project never blocks the schedule. Projects using HTTP Basic auth are skipped, as their credentials are only provided interactively and are not stored.

Restarting the server never triggers a burst of catch-up crawls: a project whose scheduled time passed while the server was down runs at most once on the next scheduler tick, then resumes its normal interval.

### Crawl notifications

Each project can have a notification webhook URL, entered when the project is created or edited. After every crawl finishes — manual or scheduled — a JSON summary is POSTed to that URL:

```json
{
  "project_id": 1,
  "url": "https://example.com",
  "crawl_id": 42,
  "scheduled": true,
  "start": "2026-10-09T12:00:00Z",
  "end": "2026-10-09T12:05:00Z",
  "total_urls": 120,
  "total_issues": 15,
  "critical_issues": 1,
  "alert_issues": 4,
  "warning_issues": 10
}
```

The `scheduled` flag tells whether the crawl was started by the scheduler or manually. Delivery failures are logged by the server and never fail the crawl. The URL must be an absolute `http` or `https` URL; leaving it empty disables notifications. The payload is plain JSON, so any receiver that accepts JSON POSTs works; services expecting their own payload shape (Slack `text`, Discord `content`, ...) need a small transform, e.g. a Zapier/n8n webhook step.

---

## Setting Up a Reverse Proxy

### Nginx Configuration

1. **Install Nginx**  
   Install Nginx using your package manager:

   `sudo apt update && sudo apt install nginx`

2. **Create a Configuration File**  
   Create `/etc/nginx/sites-available/seonaut` with the following content:

```nginx
    server {
        listen 80;
        server_name example.com;

        location / {
            proxy_pass http://localhost:9000;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_cache_bypass $http_upgrade;
        }

        location /crawl/ws {
            proxy_pass http://localhost:9000/crawl/ws;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_cache_bypass $http_upgrade;
            chunked_transfer_encoding off;
            proxy_buffering off;
            proxy_cache off;
        }
    }
```

3. **Enable HTTPS with Certbot**  
   Install Certbot and configure HTTPS:

   `sudo apt install certbot python3-certbot-nginx`

   `sudo certbot --nginx -d example.com`

4. **Restart Nginx**  
   Reload the configuration:

   `sudo systemctl reload nginx`

---

### Apache Configuration

1. **Install Apache**  
   Install Apache using your package manager:

   `sudo apt update && sudo apt install apache2`

2. **Enable Required Modules**  
   Enable the necessary Apache modules:

   `sudo a2enmod proxy proxy_http proxy_wstunnel ssl`

3. **Create a Virtual Host File**  
   Add `/etc/apache2/sites-available/seonaut.conf` with the following content:

    <VirtualHost *:80>
        ServerName example.com

        ProxyPreserveHost On
        ProxyRequests Off

        <Location />
            ProxyPass http://localhost:9000/
            ProxyPassReverse http://localhost:9000/
        </Location>

        <Location /crawl/ws>
            ProxyPass ws://localhost:9000/crawl/ws
            ProxyPassReverse ws://localhost:9000/crawl/ws
        </Location>
    </VirtualHost>

4. **Enable HTTPS with Certbot**  
   Configure HTTPS:

   `sudo apt install certbot python3-certbot-apache`

   `sudo certbot --apache -d example.com`

5. **Restart Apache**  
   Reload the configuration:

   `sudo systemctl reload apache2`
