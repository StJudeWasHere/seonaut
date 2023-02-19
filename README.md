# SEOnaut
[![Go Report Card](https://goreportcard.com/badge/github.com/stjudewashere/seonaut)](https://goreportcard.com/report/github.com/stjudewashere/seonaut) [![GitHub](https://img.shields.io/github/license/StJudeWasHere/seonaut)](LICENSE) [![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/StJudeWasHere/seonaut/test.yml)](https://github.com/StJudeWasHere/seonaut/actions/workflows/test.yml)

<img src="resources/logo.jpg" alt="Logo" align="right"/>

SEOnaut is an open source SEO auditing tool that checks your website for any issues that might be affecting your search engine rankings. It will look at your entire site and give you a report with a list of any problems it finds, organized by how important they are to fix.

The issues on your website are organized into three categories based on their level of severity and potential impact on your search engine rankings. SEOnaut can identify broken links to prevent 404 not found errors, temporary or permanent redirects and redirect loops, missing or duplicated meta tags, missing or incorrectly ordered headings and more.

A hosted version of SEOnaut is available at [seonaut.org](https://seonaut.org).

## Running SEOnaut with Docker

Run SEOnaut with docker-compose:

```shell
$ docker-compose up -d --build
```

Edit the _config_ file if you need to customize the settings, then browse to ```http://localhost:9000```.

## Contributing

Please see [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

SEOnaut is licensed under the MIT license. See [License File](LICENSE) for more information.
