# Vanity

A small HTTP server for [Go vanity import paths](https://go.dev/ref/mod#vcs).

It serves the `go-import` metadata required by the Go toolchain and optionally provides `go-source` metadata for source-code navigation.

The project has no runtime dependencies and is intentionally kept small. Configuration is provided as YAML.

## How it works

Suppose a project is hosted at:

```text
https://git.example.org/example/project
```

but you want users to import it as:

```go
import "go.example.org/project"
```

The vanity server responds to:

```text
https://go.example.org/project?go-get=1
```

with metadata such as:

```html
<meta name="go-import"
      content="go.example.org/project git https://git.example.org/example/project">
```

Go then uses the repository URL to obtain the module.

The server can also provide `go-source` metadata:

```html
<meta name="go-source"
      content="go.example.org/project https://git.example.org/example/project/tree/{dir} https://git.example.org/example/project/blob/{file} #L{line}">
```

## Features

* Go `go-import` discovery.
* Optional `go-source` metadata.
* YAML configuration (w/ validation).
* Import path matching with wildcard rules.
* Most-specific matching rule wins.
* Docker image and Docker Compose configuration.

## Installation

### Build from source

Requires Go.

Clone the repository and build:

```sh
go build -o vanity .
```

Install directly into `$GOBIN`:

```sh
go install .
```

### Docker

Build the image:

```sh
docker build -t vanity .
```

The image contains only the compiled application and uses the configuration mounted at runtime.

### Docker Compose

The repository includes `compose.yaml` and `config.example.yaml`.

Adjust the configuration and start the service:

```sh
docker compose up -d --build
```

The example Compose configuration exposes the server on port `80`.

## License

See the [LICENSE](LICENSE) file for license rights and limitations (MIT).