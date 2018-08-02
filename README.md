# stellar

A command line utility for accessing the StellarStation API. Let's build stellar apps together!

## Using the app

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

## Developing

The utility requires either Java or Go 1.11beta2. When using Java with Gradle, Go is automatically
downloaded and does not need to be installed.

```bash
# Java users
$ ./gradlew build

# Non-Java users
$ go build
```

The Gradle build using Java is our canonical build and is recommended to make sure results are
the same as continuous integration.
