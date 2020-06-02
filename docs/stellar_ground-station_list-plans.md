## stellar ground-station list-plans

Lists plans on a ground station.

### Synopsis

Lists plans on a ground station. Plans having AOS between the given time range are returned.
When run with default flags, plans in the next 31 days are returned.

```
stellar ground-station list-plans [Ground Station ID] [flags]
```

### Options

```
  -a, --aos-after string    The start time (UTC) of the range of plans to list (inclusive). Example: "2006-01-02 15:04:00" (default current time)
  -b, --aos-before string   The end time (UTC) of the range of plans to list (exclusive). Example: "2006-01-02 15:14:00" (default aos-after + 31 days)
  -d, --duration uint8      Duration of the range of plans to list (1-31), in days. Duration will be ignored when aos-before is specified. (default 31)
  -h, --help                help for list-plans
  -o, --output string       Output format. One of: csv|wide|json (default "wide")
  -v, --verbose             Output more information in JSON format. (default false)
```

### SEE ALSO

* [stellar ground-station](stellar_ground-station.md)	 - Commands for working with ground stations.

