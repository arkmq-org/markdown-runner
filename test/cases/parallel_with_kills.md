# Parallel Execution Test with failures and kills

```bash {"stage":"root", "runtime":"bash", "parallel":true, "label":"Parallel task A"}
sleep .01
echo success!
```

```bash {"stage":"root", "runtime":"bash", "parallel":true, "label":"Parallel task B"}
sleep .02
echo "This task will fail now"
exit 1
```

```bash {"stage":"root", "runtime":"bash", "parallel":true, "label":"Parallel task C"}
sleep 10
echo should not write this, getting killed before
```

```bash {"stage":"root", "runtime":"bash", "parallel":true, "label":"Parallel task D"}
sleep 20
echo should not write this, getting killed before
```
