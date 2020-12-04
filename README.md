# vmconsole

read and parse vmconsole logs for Kata containers

## How to run


Download and build

```
$ go get github.com/liubin/vmconsole
```

### Realtime watching

```
$ vmconsole
```

### Parse log file

```
$ journalctl --no-pager -t kata --since="2020-12-04 06:07:20" --until="2020-12-04 08:08:00" > /tmp/guest.log
$ vmconsole /tmp/guest.log
```

