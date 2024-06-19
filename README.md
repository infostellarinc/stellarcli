# stellar

The `stellar` command line interface (CLI) is an __experimental__ tool. Support is provided on a "best effort" basis. It is strongly recommended that users develop their own client-side software to interface with the StellarStation API.

See the [documentation](/docs/stellar.md) for more details about commands in the latest version.

In the CLI itself, execute `stellar --help` to see more details about commands available your installed version.

## Getting Started and Basic Usage

### Download and Installation

Precompiled binaries of the app can be found on the [releases page](https://github.com/infostellarinc/stellarcli/releases).

Remember to extract the `stellar` executable, place the executable in a good location, give your user the correct file permissions to execute `stellar`, and optionally make sure that location is on the system path.

### Preparing your Environment

Decide which service endpoint you want to use. As of writing this, there are two service endpoints available to users:
1. Test Environment: `api.qa.stellarstation.com:443`
2. Production Environment: `api.stellarstation.com:443`

We will use the Test Environment for the purpose of this walkthrough.

Create an environment variable for your machine. This depends on your operating system and shell.

Linux

We will use Bash shell for the purpose of this walkthrough.

1. Open ~/.bashrc
1. Add `export STELLARSTATION_API_URL="api.qa.stellarstation.com:443"` to the end of the file.
1. Save and close the file.
1. In the terminal, execute `source ~/.bashrc`

Windows

It will look something like [this procedure](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_environment_variables?view=powershell-7.4#set-environment-variables-in-the-system-control-panel).

1. Search `Environment Variables` in the Windows search bar and open the `Environment Variables` manager.
1. Create a new environment variable with the name `STELLARSTATION_API_URL` and the value `api.qa.stellarstation.com:443`.
1. Log out and log back in.

### Authentication

You will need an API key to fully use the CLI. This walkthrough assumes you have created an account on the appropriate server and know how to generate an API key through the StellarStation Console. As of writing this, there are two possible StellarStation Consoles available for users to log in to:
1. Test Environment: `www.qa.stellarstation.com`
2. Production Environment: `www.stellarstation.com`

Generate a private API key, download it, and place it somewhere safe.

__IMPORTANT__: Do not share this private key with anyone.

Execute `stellar auth activate-api-key path/to/key.json`.

### Change the TLE

Now that you have completed authentication, we can update the TLE. This walkthrough assumes an Infostellar support engineer has configured a test ground station for your account below 51 degrees latitude.

Execute `stellar satellite add-tle [STELLARSTATION SATELLITE ID HERE] "[TLE LINE 1]" "[TLE LINE 2]"`. Try the [Space Station](https://celestrak.org/NORAD/elements/gp.php?CATNR=25544) as a good temporary TLE for now.

Now that you have provided a new TLE, wait several seconds, and check that the TLE propagated to StellarStation.

Execute `stellar satellite get-tle [STELLARSTATION SATELLITE ID HERE]`. The returned TLE should match the one you provided above.

__WARNING__: `add-tle` sets the TLE updater in StellarStation to "Manual" mode - meaning that StellarStation will not automatically update the TLE. To set up automatic TLE updates please contact an Infostellar support engineer.

### Getting a list of Available Passes

We can see a list of available passes now that the TLE is set.

Execute `stellar satellite list-passes [STELLARSTATION SATELLITE ID HERE] --verbose`.

This will print out a huge list of passes (and corresponding reservation tokens for later use).

### Reserving a Plan

We can make a reservation now that we have the list of passes. Try reserving the last available pass in the list using it's `reservationToken`.

Execute `stellar satellite reserve-pass [reservationToken]`.

### Getting a list of Reserved Plans

We can get a list of plans as well.

Execute `stellar satellite list-plans [STELLARSTATION SATELLITE ID HERE]`.

### Canceling a Plan

We can cancel the plan above now that we have finished this walkthrough. You'll need the Plan ID from the reserved plan list.

Execute `stellar satellite cancel-plan [PLAN ID]`.

## Development

### Go Version
Infostellar uses the 2nd-latest Go version to reduce security risk and release issues.

### Linting
Infostellar uses [golangci-lint](https://golangci-lint.run/welcome/install/) for linting/static analysis rules.

### Releases
Infostellar uses [Go Releaser](https://goreleaser.com/).

## Frequently Asked Questions (FAQ)

### Can anyone use this?
The `stellar` CLI tool is offered with an Open Source license. Anyone is welcome to use this tool for testing, experimentation, etc.

### I need help. What should I do?
If you experience trouble with the tool, please open a new Issue or contact an Infostellar support engineer.

### I want to help. What can I do?
If you want to contribute, feel free to comment on Issues or open a Pull Request.

### Why are things misaligning and borders at the wrong widths when I use `interactive-plan`?
This is most likely due to your locale and encoding, particularly with regard to Chinese, Japanese, and Korean (for example, `zh_CN.UTF-8` or `ja_JP.UTF-8`). The most direct way to fix this is to set `RUNEWIDTH_EASTASIAN=0` in your environment. For more details see https://github.com/charmbracelet/lipgloss/issues/40.