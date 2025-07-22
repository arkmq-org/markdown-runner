# File Writer Test

```md {"stage":"write_file", "runtime":"writer", "destination":"hello.txt", "rootdir":"$tmpdir.writer"}
Hello from the writer!
This is a test.
```

```bash {"stage":"verify_file", "runtime":"bash", "rootdir":"$tmpdir.writer"}
if grep -q "Hello from the writer!" hello.txt; then
  echo "File content is correct."
else
  echo "File content is incorrect."
  exit 1
fi
```

