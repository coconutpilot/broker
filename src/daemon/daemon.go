package daemon

import (
	"errors"
	"log"
	"net"
	"time"
)

type daemon struct {
	*net.TCPListener
	stop chan int
}

func New(address string) (*daemon, error) {
	log.Println("New()")
	l_tmp, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	l, ok := l_tmp.(*net.TCPListener)
	if !ok {
		panic("Type conversion failed Listener -> TCPListener")
	}

	retval := &daemon{}
	retval.TCPListener = l
	retval.stop = make(chan int)

	return retval, nil
}

func (l *daemon) Accept() (net.Conn, error) {
	log.Println("Accept()")
	for {
		l.SetDeadline(time.Now().Add(time.Second))

		newConn, err := l.TCPListener.Accept()

		select {
		case <-l.stop:
			_ = l.TCPListener.Close()
			return nil, errors.New("Shutting down, listening socket closed")
			// we may have abandoned a freshly accepted connection in newConn here
		default:
		}

		if err != nil {
			netErr, ok := err.(net.Error)

			if ok && netErr.Timeout() {
				// timeout in SetDeadline hit, loop
				continue
			} else {
				return nil, err
			}
		}

		return newConn, nil
	}
}

func (l *daemon) Stop() {
	log.Println("Stop()")
	close(l.stop)
}
