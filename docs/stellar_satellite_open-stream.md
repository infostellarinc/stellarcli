## stellar satellite open-stream

Opens a proxy to stream packets to and from a satellite.

### Synopsis

Opens a proxy to stream packets to and from a satellite. Both TCP and UDP are supported but
TCP is preferred. Packets received by the proxy will be sent with the specified framing to
the satellite and any incoming packets will be returned as is.

```
stellar satellite open-stream [satellite-id] [flags]
```

### Options

```
      --accepted-framing strings    Framing type to receive. One of: WATERFALL|BITSTREAM|AX25|IQ|IMAGE_PNG|IMAGE_JPEG|FREE_TEXT_UTF8
      --accepted-plan-id strings    Plan ID(s) to accept data from.
      --auto-close-delay duration   The duration to wait before ending the stream with no more data incoming. Valid time units are "s", "m". Ex 1m30s. Range 1s to 10m (default 5s)
      --auto-close-time string      The datetime (UTC) after which auto-closing will be enabled. Format 2006-01-02 15:04:05
      --correct-order               When set to true, packets will be sorted by time_first_byte_received. This feature is alpha quality.
      --debug                       Output debug information. (default false)
      --delay-threshold duration    The maximum amount of time that packets remain in the sorting pool. (default 500ms)
      --enable-auto-close           When set to true, the stream will close after a specified auto close time.
  -h, --help                        help for open-stream
      --listen-host string          Deprecated: use udp-listen-host instead.
      --listen-port uint16          Deprecated: use udp-listen-port instead.
      --output-file string          [Alpha feature] The file to write packets to. Creates file if it does not exist; appends to file if it already exists. (default none)
      --proxy string                Proxy protocol. One of: udp|tcp|disabled (default "udp")
      --send-host string            Deprecated: use udp-send-host instead.
      --send-port uint16            Deprecated: use udp-send-port instead.
      --stats                       [Alpha feature] Output telemetry stats information and generate pass summaries (default false)
  -r, --stream-id string            The StreamId to resume.
      --tcp-listen-host string      The host to listen for TCP connection on. (default "127.0.0.1")
      --tcp-listen-port uint16      The port used to communicate with satellite. Clients can receive and send data through the port. (default 6001)
      --udp-listen-host string      The host to listen for packets on. (default "127.0.0.1")
      --udp-listen-port uint16      The port stellar listens for packets on. Packets on this port will be sent to the satellite. (default 6000)
      --udp-send-host string        The host to send UDP packets to. (default "127.0.0.1")
      --udp-send-port uint16        The port stellar sends UDP packets to. Packets from the satellite will be sent to this port. (default 6001)
  -v, --verbose                     Output more information. (default false)
```

### SEE ALSO

* [stellar satellite](stellar_satellite.md)	 - Commands for working with satellites

