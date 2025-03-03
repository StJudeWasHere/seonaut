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

### Using Docker

1. **Clone the Repository**  
   Clone the SEOnaut repository from GitHub:

   `git clone https://github.com/stjudewashere/seonaut.git`

2. **Navigate to the Project Directory**  
   Move to the project folder:

   `cd seonaut`

3. **Build and Run Docker Containers**  
   Run docker-compose to build and start the containers:

   `docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build`

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

---

## Setting Up a Reverse Proxy

### Nginx Configuration

1. **Install Nginx**  
   Install Nginx using your package manager:

   `sudo apt update && sudo apt install nginx`

2. **Create a Configuration File**  
   Create `/etc/nginx/sites-available/seonaut` with the following content:

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
        }
    }

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
