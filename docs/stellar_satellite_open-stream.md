## stellar satellite open-stream

Opens a proxy to stream packets to and from a satellite.

### Synopsis

Opens a proxy to stream packets to and from a satellite. Currently only
UDP is supported. Packets received by the proxy will be sent with the specified framing to
the satellite and any incoming packets will be returned as is.

```
stellar satellite open-stream [satellite-id] [flags]
```

### Options

```
      --accepted-framing strings   Framing type to receive. One of: IMAGE_PNG|IMAGE_JPEG|FREE_TEXT_UTF8|WATERFALL|BITSTREAM|AX25|IQ
      --accepted-plan-id strings   Plan ID(s) to accept data from.
      --correct-order              When set to true, packets will be sorted by time_first_byte_received. This feature is alpha quality.
      --debug                      Output debug information. (default false)
      --delay-threshold duration   The maximum amount of time that packets remain in the sorting pool. (default 500ms)
  -h, --help                       help for open-stream
      --listen-host string         Deprecated: use udp-listen-host instead.
      --listen-port uint16         Deprecated: use udp-listen-port instead.
      --proxy string               Proxy protocol. One of: udp|tcp (default "udp")
      --send-host string           Deprecated: use udp-send-host instead.
      --send-port uint16           Deprecated: use udp-send-port instead.
  -r, --stream-id string           The StreamId to resume.
      --tcp-listen-host string     The host to listen for TCP connection on. (default "127.0.0.1")
      --tcp-listen-port uint16     The port used to communicate with satellite. Clients can receive and send data through the port. (default 6001)
      --udp-listen-host string     The host to listen for packets on. (default "127.0.0.1")
      --udp-listen-port uint16     The port stellar listens for packets on. Packets on this port will be sent to the satellite. (default 6000)
      --udp-send-host string       The host to send UDP packets to. (default "127.0.0.1")
      --udp-send-port uint16       The port stellar sends UDP packets to. Packets from the satellite will be sent to this port. (default 6001)
  -v, --verbose                    Output more information. (default false)
```

### SEE ALSO

* [stellar satellite](stellar_satellite.md)	 - Commands for working with satellites

