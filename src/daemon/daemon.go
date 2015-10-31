package daemon

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type daemon struct {
	*net.TCPListener
	bus chan int
	wg  sync.WaitGroup
}

type daemonConn struct {
	net.Conn
	wg *sync.WaitGroup
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
	retval.bus = make(chan int)
	retval.wg.Add(1)

	return retval, nil
}

func (l *daemon) Accept() (net.Conn, error) {
	log.Println("Accept()")

	for {
		l.SetDeadline(time.Now().Add(time.Second))

		newConn, err := l.TCPListener.Accept()

		select {
		case <-l.bus:
			log.Println("Signal")
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

		log.Println("returning newConn")
		l.wg.Add(1)
		newConn = daemonConn{Conn: newConn, wg: &l.wg}
		return newConn, nil
	}
}

func (l *daemon) Close() error {
	log.Println("Listener Close()")
	err := l.TCPListener.Close()
	defer l.wg.Done()
	return err
}

func (l *daemon) Stop() {
	log.Println("Stop()")
	close(l.bus)
	l.wg.Wait()
}

func (w daemonConn) Close() error {
	log.Println("Close()")
	defer w.wg.Done()
	return w.Conn.Close()
}
