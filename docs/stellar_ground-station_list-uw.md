## stellar ground-station list-uw

Lists unavailability windows on a ground station.

### Synopsis

Lists unavailability windows on a ground station. Unavailability windows between the given time range
are returned.

```
stellar ground-station list-uw [Ground Station ID] [flags]
```

### Options

```
  -d, --duration uint16     Duration of the range of plans to list (1-365), in days. Duration will be ignored when end-time is specified. (default 31)
  -e, --end-time string     The end time (UTC) of the range of unavailability windows to list (exclusive).
                            			Example: "2006-01-02 15:14:00" (default start-time + 31 days
  -h, --help                help for list-uw
  -o, --output string       Output format. One of: csv|wide|json (default "wide")
  -s, --start-time string   The start time (UTC) of the range of unavailability windows to list (inclusive).
                            			Example: "2006-01-02 15:04:00 (default current time"
```

### SEE ALSO

* [stellar ground-station](stellar_ground-station.md)	 - Commands for working with ground stations.

