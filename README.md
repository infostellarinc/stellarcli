# stellar

The `stellar` command line interface (CLI) is an __experimental__ tool primarily used by the StellarStation development team
to perform internal end-to-end tests when developing the [StellarStation API](https://github.com/infostellarinc/stellarstation-api).

For best results, it is strongly recommended that users develop their own client interfaces which communicate directly
with the defined StellarStation API, rather than through the `stellar` CLI tool.

The `stellar` CLI tool is offered with an Open Source license with the purpose of sharing knowledge about the StellarStation platform.
Anyone is welcome to use this tool for their own testing purposes, or to use it as an example of how a client can communicate
with the StellarStation API.

Since this tool is in the Alpha stages of development, support is only provided on a "best effort" basis.

## Download

Precompiled binaries of the app can be found on the [releases page](https://github.com/infostellarinc/stellarcli/releases).

We attempt to give detailed information on using the app in the tool's help message.

```bash
$ stellar -h
```

If you see anything is unclear, feel free to file an issue or send a pull request.

## Authentication

The utility can be authenticated using a StellarStation API key. If you don't have one yet,
issue a key at your account page on StellarStation Console. The key can be activated using

```bash
$ stellar auth activate-api-key path/to/stellarstation-private-key.json
```

All commands after that will be authenticated using that key.

### interactive-plan: Why are things misaligning? Why are borders at the wrong widths?

This is most likely due to your locale and encoding, particularly with regard to Chinese, Japanese, and Korean (for example, `zh_CN.UTF-8` or `ja_JP.UTF-8`). The most direct way to fix this is to set `RUNEWIDTH_EASTASIAN=0` in your environment.

For details see https://github.com/charmbracelet/lipgloss/issues/40.

## Documentation

[Documentation](/docs/stellar.md) page describes more detail of stellar commands.

## Testing

When testing the CLI against the [StellarStation API Fakeserver](https://github.com/infostellarinc/stellarstation-api/tree/master/examples/fakeserver),
you may change the server endpoint by setting the `STELLARSTATION_API_URL` environment variable, eg:

```bash
$ export STELLARSTATION_API_URL=localhost:8080
```

## Developing

We stay on the 2nd latest Go version to ensure security patches and to avoid any issues on any releases.

```bash
# Go users
$ go build
```
