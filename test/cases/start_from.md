# Start from test

Used by run.sh with `-s start_from@main/second` to check that the chunks preceding
the requested one are not executed.

```bash {"stage":"init"}
echo start-from marker init
```

```bash {"stage":"main", "id":"first"}
echo start-from marker first
```

```bash {"stage":"main", "id":"second"}
echo start-from marker second
```

```bash {"stage":"end"}
echo start-from marker end
```
