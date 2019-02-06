package stream

import (
	"golang.org/x/net/netutil"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const (
	MAX_CONNECTION   = 1
	LISTEN_TIMEOUT   = 5 * time.Second
	RECV_BUFFER_SIZE = 1024 * 1024
	RECV_TIMEOUT     = 500 * time.Millisecond
)

type TCPProxy interface {
	io.Closer
}

type tcpProxy struct {
	listener        *net.TCPListener
	listenCloseChan chan struct{}
	conn            net.Conn
	stream          SatelliteStream

	recvBuf       []byte
	recvCloseChan chan struct{}

	sendChan      chan []byte
	sendCloseChan chan struct{}

	closeWg sync.WaitGroup
}

func StartTCPProxy(addr string, satelliteId string) (TCPProxy, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	l = netutil.LimitListener(l, MAX_CONNECTION)
	tcpListener := l.(*net.TCPListener)

	sendChan := make(chan []byte)

	stream, err := OpenSatelliteStream(satelliteId, sendChan)
	if err != nil {
		l.Close()
		return nil, err
	}

	t := &tcpProxy{
		listener: tcpListener,
		stream:   stream,
		sendChan: sendChan,
		recvBuf:  make([]byte, RECV_BUFFER_SIZE),

		listenCloseChan: make(chan struct{}),
		sendCloseChan:   make(chan struct{}),
		recvCloseChan:   make(chan struct{}),

		closeWg: sync.WaitGroup{},
	}

	t.start()

	return t, nil
}

func (t *tcpProxy) Close() error {
	t.stream.Close()

	close(t.recvCloseChan)
	close(t.sendCloseChan)
	close(t.listenCloseChan)

	t.closeWg.Wait()

	t.conn.Close()
	t.listener.Close()

	return nil
}

func (t *tcpProxy) listen() {
	defer t.closeWg.Done()
	for {
		select {
		case <-t.listenCloseChan:
			return
		default:
			t.listener.SetDeadline(time.Now().Add(LISTEN_TIMEOUT))
			conn, err := t.listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					// timeout
					continue
				}
				log.Printf("Could not accept incoming connection: %v\n", err)
				continue
			}
			t.conn = conn
			t.closeWg.Add(2)
			go t.recvLoop()
			go t.sendLoop()
		}
	}
}

func (t *tcpProxy) recvLoop() {
	defer t.closeWg.Done()
	for {
		select {
		case <-t.recvCloseChan:
			return
		default:
			t.conn.SetReadDeadline(time.Now().Add(RECV_TIMEOUT))
			n, err := t.conn.Read(t.recvBuf)
			if err != nil {
				if !err.(net.Error).Timeout() {
					log.Fatalf("Error receiving on TCP port: %v\n", err)
				}
			} else {
				t.stream.Send(t.recvBuf[:n])
			}
		}
	}

}

func (t *tcpProxy) sendLoop() {
	defer t.closeWg.Done()
	for {
		select {
		case payload := <-t.sendChan:
			t.conn.Write(payload)
		case <-t.sendCloseChan:
			return
		}
	}
}

func (t *tcpProxy) start() {
	t.closeWg.Add(1)
	go t.listen()
}
