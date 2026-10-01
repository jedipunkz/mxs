# mxs

A CLI to look up Minecraft player IDs: Bedrock XUID, Java UUID, and Floodgate UUID.

## Install

```sh
go install github.com/jedipunkz/mxs@latest
```

Or build from source:

```sh
make build   # produces ./mxs
```

## Usage

```
mxs bedrock <gamertag>              Bedrock XUID
mxs bedrock -r <xuid>               Bedrock gamertag
mxs java <account name>             Java UUID
mxs java -r <uuid>                  Java account name
mxs floodgate <gamertag>            Floodgate UUID
mxs floodgate -r <floodgate uuid>   Bedrock gamertag
```

`-r` (reverse) must come before the argument.

```console
$ mxs java Notch
069a79f4-44e9-4726-a5be-fca90e38aaf5
$ mxs java -r 069a79f4-44e9-4726-a5be-fca90e38aaf5
Notch
$ mxs bedrock Dream
2535414915229641
$ mxs floodgate Dream
00000000-0000-0000-0009-01f2496167c9
$ mxs floodgate -r 00000000-0000-0000-0009-01f2496167c9
Dream
```

## Data sources

| Edition | API |
|---|---|
| Java | [Mojang API](https://api.mojang.com) / session server |
| Bedrock | [GeyserMC Global API](https://api.geysermc.org/v2/docs) |

## Notes

- Bedrock lookups only work for players cached by GeyserMC, i.e. players who have joined a Geyser server at least once. Others return `HTTP 503 Unable to find user in our cache`.
- A Floodgate UUID is derived from the XUID as `new UUID(0, xuid)`, so it always starts with `00000000-0000-0000-`. Players with linked Java accounts appear on Floodgate servers with their Java UUID instead; use `mxs java -r` for those.

## Development

```sh
make test    # go vet + go test
make clean
```

## License

[MIT](LICENSE)
