# Markdown Runner

Markdown Runner is a simple Go application designed to execute Markdown files as
pipelines.

One of the main use cases of the Markdown runner tool is to have your
documentation testable in a CI environment. This way, you can always detect if
your documentation is drifting away from the reality of your codebase.

## Usage

With installing on your system:

```bash
go install -x github.com/arkmq-org/markdown-runner@latest
markdown_runner
```

Without installing, you can run the tool directly from the source code:

```bash
go run main.go
```

By default, the tool will look for markdown files in the current directory.
To point to a specific file or directory, specify the path as an argument.

> [!NOTE]
> To avoid infinite loops, the above chunks are not configured to be executable.

### Available options

````bash {"stage":"help", "label": "display help"}
markdown-runner --help
````
````shell markdown_runner
Usage: markdown-runner [options] [path]

Executes markdown files as scripts.
The default path is the current directory.

Modes:
  -d, --dry-run              Just list what would be executed without doing it
  -l, --list                 Just list the files found

Execution Control:
  -i, --interactive          Prompt to press enter between each chunk
  -s, --start-from string    Start from a specific stage name
  -t, --timeout int          The timeout in minutes for every executed command (default 10)
  -u, --update-files         Update the chunk output section in the markdown files
      --ignore-breakpoints   Ignore the breakpoints

File Selection:
  -f, --filter string        Run only the files matching the regex
  -r, --recursive            Search for markdown files recursively

Output & Logging:
      --view string          UI to be used, can be 'default' or 'ci'
  -v, --verbose              Print more logs
  -q, --quiet                Disable output
      --no-styling           Disable spinners in CLI

Help:
  -h, --help                 Show this help message
````

## Development setup

### Prerequisites

* Go 1.23 or later

### Testing

Unit tests can be run with:

```bash {"stage":"test", "rootdir":"$initial_dir", "label": "Run unit tests"}
go test ./...
```
```shell markdown_runner
?   	github.com/arkmq-org/markdown-runner/pterm	[no test files]
?   	github.com/arkmq-org/markdown-runner/runnercontext	[no test files]
?   	github.com/arkmq-org/markdown-runner/view	[no test files]
ok  	github.com/arkmq-org/markdown-runner	0.005s
ok  	github.com/arkmq-org/markdown-runner/chunk	(cached)
ok  	github.com/arkmq-org/markdown-runner/config	(cached)
ok  	github.com/arkmq-org/markdown-runner/parser	(cached)
ok  	github.com/arkmq-org/markdown-runner/runner	(cached)
ok  	github.com/arkmq-org/markdown-runner/stage	(cached)
```

Integration tests can be run with:

```bash {"stage":"integration-test", "rootdir":"$initial_dir", "label": "Run integration tests"}
./test/run.sh
```
```shell markdown_runner
[33m--- Building markdown-runner binary ---[0m
[36mRunning test #1: Dry run test for (happy) should show DRY-RUN...[0m[32m ✅[0m
[36mRunning test #2: Happy path test (happy) should succeed...[0m[32m ✅[0m
[36mRunning test #3: Test case parallel...[0m[32m ✅[0m
[36mRunning test #4: Schema error test (schema_error) should fail as expected...[0m[32m ✅[0m
[36mRunning test #5: Teardown test (teardown) should execute teardown...[0m[32m ✅[0m
[36mRunning test #6: Test case writer...[0m[32m ✅[0m
[36mRunning test #7: Recursive test...[0m[32m ✅[0m
[36mRunning test #8: Recursive test with file filter...[0m[32m ✅[0m
[33m\n--- Test Summary ---[0m
[32mAll 8 tests passed![0m
[33m--- Cleaning up ---[0m
[32mCleanup complete.[0m
```

You can also run the integration tests with `-v` to show the details of every
tests

### Architecture

```mermaid
classDiagram
    direction BT

    class Runner {
        <<Package>>
        +RunMD(file string) error
    }

    class Parser {
        <<Package>>
        +ExtractStages(file string) []*stage.Stage
        +UpdateChunkOutput(file string, stages []*stage.Stage) error
    }

    class Config {
        <<Package>>
        +Rootdir string
        +Verbose bool
        +UpdateFile bool
        +DryRun bool
    }

    class Stage {
        +Name string
        +Chunks []*chunk.ExecutableChunk
        +Execute(stages []*Stage, tmpDirs map) error
        -isParallelismConsistent() bool
    }

    class ExecutableChunk {
        +Stage string
        +Id string
        +Runtime string
        +IsParallel bool
        +Content []string
        +Commands []*RunningCommand
        +PrepareForExecution(tmpDirs map) error
        +Execute() error
        +Wait(shouldKill bool) error
    }

    class RunningCommand {
        +Cmd *exec.Cmd
        +Stdout string
        +Stderr string
        +CancelFunc context.CancelFunc
        +Execute() error
        +Start() error
        +Wait() error
    }

    Runner ..> Parser : Uses
    Runner ..> Stage : Executes
    Runner ..> Config : Reads

    Parser ..> Stage : Creates
    Parser ..> ExecutableChunk : Creates

    Stage "1" *-- "1..*" ExecutableChunk : Contains
    ExecutableChunk "1" *-- "0..*" RunningCommand : Contains

    Stage ..> ExecutableChunk : Manages
    ExecutableChunk ..> RunningCommand : Manages

    Stage ..> Config : Reads
    ExecutableChunk ..> Config : Reads
```

### Contributing

When contributing to this project, please make sure to update the documentation
as you add in new features or fix bugs.

Use the `-update-files` option when running the tool on this README.

## How to write an executable markdown file

### The terminology: Stages, chunks and commands

Executable markdown files work a bit like a CI workflow. The command is the
smallest element of them. It corresponds to an executable with its arguments to
be called onto the system. A chunk is a collection of commands and is used to
group them with an environment of execution, chunks can be executed in parallel
of each other within the same stage. And finally, there are the stages, which
are a collection of chunks. Stages gets executed one after the other.

```mermaid
graph TD
    subgraph "Execution Flow"
        direction LR
        Stage1["Stage 1"] --> Stage2["Stage 2"] --> StageN["..."]
    end

    subgraph "Inside a Stage"
        direction TB
        subgraph ParallelGroup [Parallel Chunks]
            direction LR
            ChunkA["Chunk A<br/>(parallel: true)"]
            ChunkB["Chunk B<br/>(parallel: true)"]
        end
        ChunkC["Chunk C"]

        ParallelGroup --> ChunkC
    end

    subgraph "Inside a Chunk"
        direction TB
        Command1["Command 1"] --> Command2["Command 2"]
    end

    classDef stage fill:#e1d5e7,stroke:#9673a6,stroke-width:2px;
    classDef chunk fill:#d5e8d4,stroke:#82b366,stroke-width:2px;
    classDef command fill:#f8cecc,stroke:#b85450,stroke-width:2px;

    class Stage1,Stage2,StageN stage;
    class ChunkA,ChunkB,ChunkC chunk;
    class Command1,Command2 command;
```

#### Code chunks

In an executable markdown file chunk, every line gets executed. In order to be
What makes a chunk executable is a pice of json metadata that is attached to the
code fence of the chunk.

The code fence to create an executable chunk needs to have at least 3 back
quotes and a corresponding closing fence later on in the document.

For example:

````
``` something {"stage":"init"}
echo "Hello, World!"
```
```shell markdown_runner
Hello, World!
```
````

The only mandatory field being the stage name. The other possible fields are:

##### `"runtime":"bash"`

Specifies that the content of the chunk represents a bash script. By default the
script will be executed in a newly created temporary directory. (See
`"rootdir"` below), and we recommend leaving it this way unless you don't mind
leaving temporary files on your disk.

The script will be executed with the environment variables of the parent process
that started the markdown runner like any other commands of the pipeline.

Some changes are made to the script before it gets executed, it'll get:

* prefixed with `set -euo pipefail`, meaning that it will
  error at the first failing commands
* suffixed with `printenv` to extract all the environment variables the
  user added in.

We recommend leaving these options in place.

##### `"runtime":"writer"`

Writes the content of the chunk to disk. The metadata needs to contain
`"destination":"some_place"` to know where to write the content.

##### `"label":"some label"`

Gives a pretty printable name to a chunk. It's good for bash runtimes, as
the command to execute is always `./script.sh` so if you want to have a
differentiable name in the logs, use this field.

##### `"rootdir"

###### `"rootdir":"$initial_dir"`

The commands in the chunk will the executed from directory where the markdown
runner was started.

By default, the chunk will be executed in a unique temporary directory that is
created at runtime. The temporary directory is created under `/tmp`, and is
named with `$tmpdir.$uuid`. The temporary directory is deleted after the chunk
is executed.

###### `"rootdir":"$tmpdir.x"`

Create a new temporary directory to execute the chunk. If another chunk reuses
the same suffix (`.x` in the example) it'll share the same directory. Different
suffixes will result in different directories. Useful when you need to chain
chunks or to reuse some cached files between chunks

The following diagram shows how `rootdir` affects the execution environment:

```mermaid
graph TD
    subgraph "Chunk Execution Environments"
        direction TB

        subgraph "Default Behavior (Unique Directories)"
            direction LR
            Chunk1["Chunk 1<br/>(rootdir: not set)"] --> Dir1["/tmp/xyz123"]
        end

        subgraph "Shared Directory Behavior"
            direction LR
            Chunk2["Chunk 2<br/>(rootdir: $tmpdir.shared)"] --> SharedDir["/tmp/abc456"]
            Chunk3["Chunk 3<br/>(rootdir: $tmpdir.shared)"] --> SharedDir
        end

        subgraph "Initial Directory"
             direction LR
             Chunk4["Chunk 4<br/>(rootdir: $initial_dir)"] --> InitialDir["/path/to/project"]
        end
    end
```

##### `"parallel":true`

Makes the chunk executed in parallel of the other chunks within the same stage.

```mermaid
graph TD
    subgraph "Stage Execution Flow"
        direction TB

        subgraph "Parallel Execution"
            direction TB
            StartP[Start Stage]

            subgraph " "
                direction LR

                subgraph "Chunk A (parallel: true)"
                    A1["Cmd 1"]
                end

                subgraph "Chunk B (parallel: true)"
                    B1["Cmd 1"]
                end
            end

            StartP --> A1
            StartP --> B1
            A1 --> EndP[End Stage]
            B1 --> EndP[End Stage]
        end

        subgraph "Sequential Execution (Default)"
            direction TB
            StartS[Start Stage] --> C["Chunk A"] --> D["Chunk B"] --> EndS[End Stage]
        end
    end
```

This works only if:

* there's only a single command in the chunk (or if the chunk is a bash chunk)
* if all the commands in the stage are made to run in parallel.

> [!CAUTION]
> All chunks within the same stage must have the same parallelism setting.
> Mixing parallel and sequential chunks in a single stage is not allowed and
> will result in an error. This ensures a consistent and predictable execution
> flow.

##### `"breakpoint":"true"`

Enters interactive mode when the chunk is started. Useful for debugging
purposes. Better to use alongside `--verbose`

##### `"id":"someID"`

Give an ID to a chunk so that it can get referenced later on

##### `"requires":"stageName/id"`

Makes the execution of the given stage dependant of the correct execution of the
pointed stage. To be used in teardown chunks.

This is particularly useful for running cleanup tasks only when their corresponding setup tasks have completed, as shown below:

```mermaid
graph TD
    subgraph "Stage: main"
      direction LR
      ChunkSuccess("Chunk 'setup' (succeeds)")
      ChunkFail("Chunk 'main-work' (fails)")
    end

    subgraph "Stage: teardown"
      direction LR
      TeardownSuccess("Teardown for 'setup'<br/>requires: main/setup")
      TeardownFail("Teardown for 'main-work'<br/>requires: main/main-work")
    end

    ChunkSuccess -- "dependency met" --> TeardownSuccess
    TeardownSuccess --> TeardownSuccessRun["Result: Executed"]

    ChunkFail -- "dependency NOT met" --> TeardownFail
    TeardownFail --> TeardownFailSkip["Result: Skipped"]
```

### Updating the markdown file with the output of the chunks

When running the markdown runner tool, you can use the `--update-files` option to
make the tool insert the output of every chunk (if any) inside the source
markdown

The results of the execution will get stored in a new code fence right after
the chunk that was executed, `shell markdown_runner` set indicating to the
tool that this a disposable code fence that can be overridden in the future.

### Execution Environment of the chunks

Every chunk is started with the environment of the parent process that started
it. If as chunk whom runtime is bash is executed, all the variables it adds to
its own env via `export` will get added to the environment of subsequent
chunks.

The following diagram illustrates this flow:

```mermaid
sequenceDiagram
    participant Runner as "Markdown Runner Process"
    participant BashChunk as "Bash Chunk"
    participant NextChunk as "Subsequent Chunk"

    autonumber

    Note over Runner: Initial Env: [VAR1, VAR2]

    Runner->>+BashChunk: Executes with current environment
    Note over BashChunk: Inherits [VAR1, VAR2]

    BashChunk->>BashChunk: Runs 'export NEW_VAR=value'
    Note over BashChunk: Internal env is now<br/>[VAR1, VAR2, NEW_VAR]

    BashChunk-->>-Runner: Finishes execution
    Note over Runner: Runner updates its environment<br/>New Env: [VAR1, VAR2, NEW_VAR]

    Runner->>+NextChunk: Executes with updated environment
    Note over NextChunk: Inherits [VAR1, VAR2, NEW_VAR]<br/>Can now access NEW_VAR

    NextChunk-->>-Runner: Finishes execution
```

The runner also injects a `WORKING_DIR` variable, which contains the path to the
directory where the `markdown-runner` was started.

### Examples

#### Creating and running our first executable markdown file

Let's create a markdown file containing an executable chunk, and save it as
`simple_tutorial.md`

````md {"stage":"test1", "runtime":"writer", "destination":"simple_tutorial.md", "rootdir":"$tmpdir.1"}
# Demo 1

This is a source markdown file it has a simple chunk that gets executed:

## the chunk
```bash {"stage":"inner_test1"}
echo this line will get printed in the output chunk below.
```
When this document is going to get executed, it'll have a new block above this
line containing the actual output of the executed chunk.
````

We will need to pass the folder containing this markdown file to the markdown
runner.

```bash {"stage":"test1", "rootdir":"$tmpdir.1", "runtime":"bash", "label": "Set MARKDOWN_DIR env var"}
export MARKDOWN_DIR=$(pwd)
```

Now run the script.

```bash {"stage":"test1", "runtime":"bash", "label": "Run markdown-runner on tutorial"}
cd ${WORKING_DIR}
go run main.go \
   --update-files \
   --quiet \
   ${MARKDOWN_DIR}
```

As we've executed the markdown runner tool with the `--update-files` option.
It'll make the tool insert the output of every chunks (if any) inside the
source markdown file. Let's have a look at the updated markdown file:

````bash {"stage":"test1", "rootdir":"$tmpdir.1", "label": "Display updated tutorial file"}
cat simple_tutorial.md
````
````shell markdown_runner
# Demo 1

This is a source markdown file it has a simple chunk that gets executed:

## the chunk
```bash {"stage":"inner_test1"}
echo this line will get printed in the output chunk below.
```
```shell markdown_runner
this line will get printed in the output chunk below.
```
When this document is going to get executed, it'll have a new block above this
line containing the actual output of the executed chunk.
````

#### Sharing variables between chunks

Any export in a bash environment makes the content available for subsequent
chunks.

````markdown
```bash {"stage":"test2", "runtime":"bash", "label": "Export environment variables"}
export SOME_VARIABLE=$(sleep .1s && echo "This is some content")
export TEST="some value"
```
````

The chunk below is able to perform some comparisons with value of
`SOME_VARIABLE`

````markdown
```bash {"stage":"test2", "runtime":"bash", "label": "Verify environment variable"}
if [ "$SOME_VARIABLE" == "This is some content" ]; then
  echo "same string"
fi
```
```shell markdown_runner
same string
```
````

#### Executing in parallel

Parallelism can be important for some workflows, when for instance two processes
need to chat with each other.

To demonstrate the capability, we're executing two pairs of chunks.
The first pair runs in parallel, and does concurrent writing to a file on disk.
The second pair acts as the control group and performs the writing sequentially,
one after the other.

On a parallel execution, we expect the content of file to contain interleaved
content from the two chunks, whereas on the sequential execution, we expect
the content of the file to be written in a single block, without interleaving.

##### The first pair is writing to `output` in parallel

````markdown
```bash {"stage":"test3", "runtime":"bash", "parallel":true, "rootdir":"$tmpdir.2", "label": "Parallel write: first writer"}
for i in $(seq 1 10);
do
    echo FIRST$i >> output
    sleep .1
done
```
````

````markdown
```bash {"stage":"test3", "runtime":"bash", "parallel":true, "rootdir":"$tmpdir.2", "label": "Parallel write: second writer"}
for i in $(seq 1 10);
do
    sleep .1
    echo SECOND$i >> output
done
```
````

##### The second pair is writing to `output2` sequentially this time

````markdown
```bash {"stage":"test4", "runtime":"bash", "rootdir":"$tmpdir.2", "label": "Sequential write: first writer"}
for i in $(seq 1 10);
do
    echo FIRST$i >> output2
done
```
````

````markdown
```bash {"stage":"test4", "runtime":"bash", "rootdir":"$tmpdir.2", "label": "Sequential write: second writer"}
for i in $(seq 1 10);
do
    echo SECOND$i >> output2
done
```
````

##### Then let's compare the two files and asses that the two files are different

````markdown
```{"stage":"test4", "rootdir":"$tmpdir.2", "runtime":"bash", "label": "Compare parallel and sequential outputs"}
DIFF=$(diff output output2 || echo "")
if [[ -z ${DIFF} ]]; then
    echo "the files are the same, that's a problem."
    exit 1
else
    echo "the two files are different, we're running in parallel"
fi
```
```shell markdown_runner
the two files are different, we're running in parallel
```
````

#### Teardown & dependencies

Teardown chunks are executed even if something went wrong in the middle of the
execution.

Adding a dependency to the teardown chunk makes it execute only if its dependant
chunk did execute correctly. That allows to make sure that a teardown stage only
tears down things that have been built before.

````markdown {"stage":"test5", "runtime":"writer", "destination":"simple_tutorial.md", "rootdir":"$tmpdir.3"}
# Demo teardown

## Two classical chunks

```bash {"stage":"inner_test1", "id":"someID", "label":"working command"}
echo this executes correctly
```

```bash {"stage":"inner_test1", "id":"someID2", "runtime":"bash", "label":"failing command"}
echo this has failed
exit 1
```

## Two teardown chunks

This one has an output, because it's dependency did run correctly

```bash {"stage":"teardown", "requires":"inner_test1/someID", "label": "Successful teardown"}
echo executed because inner_test1/someID got executed
```

This one won't

```bash {"stage":"teardown", "requires":"inner_test1/someID2", "label": "Skipped teardown"}
echo not executed because someID2 is missing in inner_test1
```
````
```bash {"stage":"test5", "rootdir":"$tmpdir.3", "runtime":"bash", "label": "store the folder of the teardown test in the MD_DIR variable"}
export MD_DIR=$(pwd)
```

Let's run and see the result

```bash {"stage":"test5", "runtime":"bash", "label": "Run teardown test"}
cd ${WORKING_DIR}
go run main.go \
   --no-styling \
   ${MD_DIR} | grep "SUCCESS.*someID" || exit 0
```
```shell markdown_runner

                                                                                working command
                                                                                SUCCESS: working command
                                                                                failing command
                                                                                ERROR: stdout:
this has failed

stderr:

exit code:1
                                                                                Successful teardown
                                                                                SUCCESS: Successful teardown
exit status 1
```

As we can see only the first teardown command did run.

## Presentation Slides

This repository includes a presentation about Markdown Runner.

- [View the slides](./slides/slides.md) (github will render the source somewhat)
- [Building the slides](./slides/README.md)
