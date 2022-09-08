# How to deploy using Docker Run

> Prereqs: a file called `./config/melp.yaml` with your settings,<br/>
> otherwise replace the `$(pwd)/config`-location as needed (absolute path is required)


## Interactive mode
_(this is mostly to get human-friendly logs, and full support of Ctrl-C to exit the program)_
```sh
# linux / macos
docker run --rm -it \
  -v "$(pwd)/config:/config" \
  -e CONFIG=/config/melp.yaml \
  -p 10000:10000 \
  lindex/melp
```

```powershell
# powershell
docker run --rm -it `
  -v "$(pwd)/config:/config" `
  -e CONFIG=/config/melp.yaml `
  -p 10000:10000 `
  lindex/melp
```

## Background service
```sh
docker run --rm -d --restart unless-stopped \
  -v "$(pwd)/config:/config" \
  -e CONFIG=/config/melp.yaml \
  -p 10000:10000 \
  lindex/melp
```

```powershell
docker run --rm -d --restart unless-stopped `
  -v "$(pwd)/config:/config" `
  -e CONFIG=/config/melp.yaml `
  -p 10000:10000 `
  lindex/melp
```

#### Getting the 'help'
```sh
docker run --rm lindex/melp melp -h
```

## Powershell considerations
You need to change the command to a one-liner, or replace ending the '\' to an \` character