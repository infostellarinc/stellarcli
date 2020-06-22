# stellar

A command line utility for accessing the [StellarStation API](https://github.com/infostellarinc/stellarstation-api).
Let's build stellar apps together!

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

## Documentation

[Documentation](/docs/stellar.md) page describes more detail of stellar commands.

## Testing

When testing the CLI against the [StellarStation API Fakeserver](https://github.com/infostellarinc/stellarstation-api/tree/master/examples/fakeserver), you may change the server endpoint by setting the `STELLARSTATION_API_URL` environment variable, eg:

```bash
$ export STELLARSTATION_API_URL=localhost:8080
```

## Developing

The utility requires either Java or Go 1.13. When using Java with Gradle, Go is automatically
downloaded and does not need to be installed.

```bash
# Java users
$ ./gradlew build

# Go users
$ go build
```

The Gradle build using Java is our canonical build and is recommended to make sure results are
the same as continuous integration.
