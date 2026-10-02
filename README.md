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
mxs bedrock <gamertag>                                       Bedrock player info
mxs java <account name>                                      Java player info
mxs reverse <xuid | hex xuid | java uuid | floodgate uuid>   Bedrock or Java player info
```

`reverse` detects the ID type: a UUID starting with `00000000-0000-0000-` is a Floodgate UUID, any other UUID is a Java UUID, and anything else is a decimal or hex XUID (a digits-only hex XUID needs the `0x` prefix).

```console
$ mxs bedrock Dream
Gamertag: Dream
XUID(DEC): 2535414915229641
XUID(HEX): 901f2496167c9
Floodgate UUID: 00000000-0000-0000-0009-01f2496167c9
$ mxs reverse 00000000-0000-0000-0009-01f2496167c9
Gamertag: Dream
XUID(DEC): 2535414915229641
XUID(HEX): 901f2496167c9
Floodgate UUID: 00000000-0000-0000-0009-01f2496167c9
$ mxs java Notch
Name: Notch
UUID: 069a79f4-44e9-4726-a5be-fca90e38aaf5
$ mxs reverse 069a79f4-44e9-4726-a5be-fca90e38aaf5
Name: Notch
UUID: 069a79f4-44e9-4726-a5be-fca90e38aaf5
```

## Data sources

| Edition | API |
|---|---|
| Java | [Mojang API](https://api.mojang.com) / session server |
| Bedrock | [GeyserMC Global API](https://api.geysermc.org/v2/docs) |

## Notes

- Bedrock lookups only work for players cached by GeyserMC, i.e. players who have joined a Geyser server at least once. Others return `HTTP 503 Unable to find user in our cache`.
- A Floodgate UUID is derived from the XUID as `new UUID(0, xuid)`, so it always starts with `00000000-0000-0000-`. Players with linked Java accounts appear on Floodgate servers with their Java UUID instead; use `mxs reverse` for those.
- Output is colored with the Tokyo Night palette (each field in its own color, a darker shade for the label and bold for the value, red `mxs:` error prefix; requires a truecolor terminal) only when writing to a terminal. Set `NO_COLOR=1` or `TERM=dumb` to disable it.

## Development

```sh
make test    # go vet + go test
make clean
```

## License

[MIT](LICENSE)
