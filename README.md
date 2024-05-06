# SEOnaut
[![Go Report Card](https://goreportcard.com/badge/github.com/stjudewashere/seonaut)](https://goreportcard.com/report/github.com/stjudewashere/seonaut) [![GitHub](https://img.shields.io/github/license/StJudeWasHere/seonaut)](LICENSE) [![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/StJudeWasHere/seonaut/test.yml)](https://github.com/StJudeWasHere/seonaut/actions/workflows/test.yml)

SEOnaut is an open source SEO auditing tool that checks your website for any issues that might be affecting your search engine rankings. It will look at your entire site and give you a report with a list of any problems it finds, organized by how important they are to fix.

The issues on your website are organized into three categories based on their level of severity and potential impact on your search engine rankings. SEOnaut can identify broken links to prevent 404 not found errors, temporary or permanent redirects and redirect loops, missing or duplicated meta tags, missing or incorrectly ordered headings and more.

A hosted version of SEOnaut is available at [seonaut.org](https://seonaut.org).

## Technology

SEOnaut is a web based application built with the Go programming language and a MySQL database for its data storage. On the frontend side, the user interface is designed with simplicity in mind, using custom CSS and minimal vanilla Javascript. To make the dashboard interactive, the application utilizes Apache ECharts.

While it is possible to configure your own database and compile SEOnaut by yourself, it's generally more convenient to use the provided Docker files. These files streamline the setup process and eliminate the need for manual configuration, allowing you to get started with SEOnaut more quickly and easily.

### Docker

Make sure you have Docker installed on your system. You can download and install Docker from their official website.

Run the following command and wait for the Docker images to download and build.

```shell
$ docker-compose up -d --build
```

Once the process is complete, you can access SEOnaut in your web browser by visiting ```http://localhost:9000```.

SEOnaut is set up to run on port 9000 using unencrypted HTTP by default. However, for security reasons, it is often advisable to run it on HTTPS behind a reverse proxy. This adds an extra layer of protection to the application and ensures that any sensitive data transmitted between the server and the client is encrypted.

## Contributing

Please see [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

SEOnaut is open-source under the MIT license. See [License File](LICENSE) for more information.
