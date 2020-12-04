# vmconsole

Read and parse VM log logs for Kata Containers. VM log includes guest's `dmesg` and log of agent.

## How to run


Download and build

```
$ go get github.com/liubin/vmconsole
```

And make sure that the `$GOPATH/bin` is in your `PATH`.

### Realtime watching

In this mode you can run `vmconsole` without arguments, and it will run `journalctl` under background, and parse the content.

```
$ vmconsole
```

### Parse log file

In this mode you should pass the path of Kata Containers' log file. This is suitable for parsing history log or sent from others.

```
$ journalctl --no-pager -t kata \
    --since="2020-12-04 06:07:20" \
    --until="2020-12-04 08:08:00" > /tmp/guest.log
$ vmconsole /tmp/guest.log
```

