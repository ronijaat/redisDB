//go:build linux

package server

import (
	"log"
	"net"
	"time"

	"golang.org/x/sys/unix"

	"github.com/ronijaat/redisDB/config"
	"github.com/ronijaat/redisDB/core"
)

var con_clients int = 0
var cronFrequency time.Duration = time.Second * 1
var lastCronExecTime time.Time = time.Now()

func RunAsyncTCPServer() error {
	log.Println("starting an asynchronous TCP server on", config.HOST, config.PORT)

	max_clients := 20000

	// Create EPOLL Event Objects to hold events
	var events []unix.EpollEvent = make([]unix.EpollEvent, max_clients)

	// Create a socket
	serverFD, err := unix.Socket(unix.AF_INET, unix.O_NONBLOCK|unix.SOCK_STREAM, 0)
	if err != nil {
		return err
	}

	defer unix.Close(serverFD)

	// Set the Socket operate in a non-blocking mode
	if err = unix.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// Bind the IP and the port
	ip4 := net.ParseIP(config.HOST)
	if err := unix.Bind(serverFD, &unix.SockaddrInet4{
		Port: config.PORT,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	// Start listening
	if err := unix.Listen(serverFD, max_clients); err != nil {
		return err
	}

	// AsyncIO starts here!!

	// creating EPOLL instance
	epollFD, err := unix.EpollCreate1(0)
	if err != nil {
		return err
	}

	defer unix.Close(epollFD)

	// Specify the events we want to get hints about
	// and set the socket on which

	var socketServerEvent unix.EpollEvent = unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// Listen to read events on the Server itself
	if err := unix.EpollCtl(epollFD, unix.EPOLL_CTL_ADD, serverFD, &socketServerEvent); err != nil {
		return err
	}

	for {
		if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys()
			lastCronExecTime = time.Now()
		}

		// see if any FD is ready for an IO
		nevents, err := unix.EpollWait(epollFD, events[:], -1)
		if err != nil {
			return err
		}

		for i := 0; i < nevents; i++ {
			// if the socket server itself is ready for an IO
			if int(events[i].Fd) == serverFD {
				// accept the incoming connection from a client
				fd, _, err := unix.Accept(serverFD)
				if err != nil {
					log.Println("err", err)
					continue
				}

				// increase the number of concurrent clients count
				con_clients++
				unix.SetNonblock(serverFD, true)

				// add this new TCP connection to be monitored
				var socketClientEvent unix.EpollEvent = unix.EpollEvent{
					Events: unix.EPOLLIN,
					Fd:     int32(fd),
				}
				if err := unix.EpollCtl(epollFD, unix.EPOLL_CTL_ADD, fd, &socketClientEvent); err != nil {
					log.Fatal(err)
				}
			} else {
				comm := core.FDComm{Fd: int(events[i].Fd)}
				cmd, err := readCommands(comm)
				if err != nil {
					unix.Close(int(events[i].Fd))
					con_clients -= 1
					continue
				}
				respond(cmd, comm)
			}
		}
	}

}
