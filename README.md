# SEOnaut
[![Go Report Card](https://goreportcard.com/badge/github.com/stjudewashere/seonaut)](https://goreportcard.com/report/github.com/stjudewashere/seonaut) [![GitHub](https://img.shields.io/github/license/StJudeWasHere/seonaut)](LICENSE) [![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/StJudeWasHere/seonaut/test.yml)](https://github.com/StJudeWasHere/seonaut/actions/workflows/test.yml)

SEOnaut is an open-source SEO auditing tool designed to analyze websites for issues that may impact search engine rankings. It performs a comprehensive site scan and generates a report detailing any identified issues, organized by severity and potential impact on SEO.

SEOnaut categorizes issues into three levels of severity: critical, high, and low. The tool can detect various SEO-related problems, such as broken links (to avoid 404 errors), redirect issues (temporary, permanent, or loops), missing or duplicate meta tags, incorrectly ordered headings, and more.

A hosted version of SEOnaut is available at [seonaut.org](https://seonaut.org).

![seonaut](https://github.com/user-attachments/assets/6184b418-bd54-4456-9266-fcfd4ce5726d)

## Technology

SEOnaut is a web-based application built with the Go programming language and a MySQL database for data storage. The frontend is designed for simplicity, using custom CSS and minimal vanilla JavaScript. Apache ECharts is used to provide an interactive dashboard experience.

While it is possible to configure a custom database and compile SEOnaut manually, using the provided Docker files is recommended. These files simplify the setup process and eliminate the need for manual configuration, allowing for quicker and easier deployment.

### Quick Start Guide

#### Using docker compose

Using docker is the recommended way of running SEOnaut. As you need to provide a database, you can use `docker compose` to do so creating a `docker-compose.yml` file like this:

```yml
services:
  db:
    image: mysql:8.4.10
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

This uses the default settings, which you can overwrite using environment variables if needed.

#### Using docker from source code

To run SEOnaut from source code, follow these steps to run it using Docker:

1. **Install Docker**  
   Ensure Docker is installed on your system. You can download and install Docker from the [official website](https://www.docker.com/).

2. **Clone the Repository**  
   Clone the SEOnaut repository:

   `git clone https://github.com/stjudewashere/seonaut.git`

3. **Navigate to the Project Directory**  
   Change into the project directory:

   `cd seonaut`

4. **Build and Run Docker Containers**  
   Run the following command to build and start the Docker containers:

   `docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build`

5. **Access the Application**  
   Once the containers are running, open your browser and visit:

   `http://localhost:9000`

   SEOnaut is set up to run on port 9000 using unencrypted HTTP by default. For added security, it is recommended to configure HTTPS using a reverse proxy. This will ensure encrypted communication between the client and the server.

For more detailed installation and configuration instructions, refer to the [INSTALL.md](docs/INSTALL.md) file.

## Contributing

Please see [CONTRIBUTING](docs/CONTRIBUTING.md) for details.

## License

SEOnaut is open-source under the MIT license. See [License File](LICENSE) for more information.
