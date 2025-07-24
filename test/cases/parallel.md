# Parallel Execution Test

```bash {"stage":"parallel_write", "runtime":"bash", "parallel":true, "rootdir":"$tmpdir.parallel", "label":"Parallel task A"}
for i in $(seq 1 5); do echo "A$i" >> output; sleep .01; done
```

```bash {"stage":"parallel_write", "runtime":"bash", "parallel":true, "rootdir":"$tmpdir.parallel", "label":"Parallel task B"}
for i in $(seq 1 5); do echo "B$i" >> output; sleep .01; done
```

```bash {"stage":"sequential_write", "runtime":"bash", "rootdir":"$tmpdir.parallel", "label":"Sequential task A"}
for i in $(seq 1 5); do echo "A$i" >> output_seq; done
```

```bash {"stage":"sequential_write", "runtime":"bash", "rootdir":"$tmpdir.parallel", "label":"Sequential task B"}
for i in $(seq 1 5); do echo "B$i" >> output_seq; done
```

```bash {"stage":"verify", "runtime":"bash", "rootdir":"$tmpdir.parallel", "label":"Compare results execution"}
# The parallel output file should be different from the sequential one.
# If they are the same, something is wrong with parallel execution.
if diff output output_seq; then
  echo "Files are the same, parallelism test failed."
  exit 1
else
  echo "Files are different, parallelism test passed."
fi
```

