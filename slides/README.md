# Presentation Slides

This directory contains the source for a Reveal.js presentation about Markdown Runner.

## Building the Slides

To build the slides, you need to have the following tools installed:

- **Pandoc**: A universal document converter.
- **mermaid-filter**: A Pandoc filter to render Mermaid diagrams.

You can install `mermaid-filter` via npm:

```bash
npm install -g mermaid-filter
```

Once the dependencies are installed, you can build the slides by running `make` in this directory:

```bash
make
```

This will generate an `index.html` file, which contains the presentation.
