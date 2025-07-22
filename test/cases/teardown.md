```bash {"stage":"main", "id":"succeeding", "runtime":"bash"}
exit 0
```

```bash {"stage":"main", "id":"failing", "runtime":"bash"}
exit 1
```

```bash {"stage":"teardown", "requires":"main/failing"}
echo "teardown does execute"
```

```bash {"stage":"teardown", "requires":"main/succeeding"}
echo teardown should execute
```
