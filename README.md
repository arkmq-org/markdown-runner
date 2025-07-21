# Markdown Runner

Markdown Runner is a simple Go application designed to execute Markdown files as
pipelines.

You can use it to turn your Markdown documentation into executable tutorials, or
into CI pipelines.

## How to use

To run all the tutorials in the `docs` directory, simply execute the following command:

```bash
go run main.go --markdown-dir docs
```

You can also specify a single tutorial to run:

```bash
go run main.go --markdown-dir docs --file <tutorial-name>.md
```
