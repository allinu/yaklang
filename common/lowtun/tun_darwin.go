/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2017-2023 WireGuard LLC. All Rights Reserved.
 */

package lowtun

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const utunControlName = "com.apple.net.utun_control"

type NativeTun struct {
	name        string
	tunFile     *os.File
	events      chan Event
	errors      chan error
	routeSocket int
	closeOnce   sync.Once
}

func retryInterfaceByIndex(index int) (iface *net.Interface, err error) {
	for i := 0; i < 20; i++ {
		iface, err = net.InterfaceByIndex(index)
		if err != nil && errors.Is(err, unix.ENOMEM) {
			time.Sleep(time.Duration(i) * time.Second / 3)
			continue
		}
		return iface, err
	}
	return nil, err
}

func (tun *NativeTun) routineRouteListener(tunIfindex int) {
	var (
		statusUp  bool
		statusMTU int
	)

	defer close(tun.events)

	data := make([]byte, os.Getpagesize())
	for {
	retry:
		n, err := unix.Read(tun.routeSocket, data)
		if err != nil {
			if errno, ok := err.(unix.Errno); ok && errno == unix.EINTR {
				goto retry
			}
			tun.errors <- err
			return
		}

		if n < 14 {
			continue
		}

		if data[3 /* type */] != unix.RTM_IFINFO {
			continue
		}
		ifindex := int(*(*uint16)(unsafe.Pointer(&data[12 /* ifindex */])))
		if ifindex != tunIfindex {
			continue
		}

		iface, err := retryInterfaceByIndex(ifindex)
		if err != nil {
			tun.errors <- err
			return
		}

		// Up / Down event
		up := (iface.Flags & net.FlagUp) != 0
		if up != statusUp && up {
			tun.events <- EventUp
		}
		if up != statusUp && !up {
			tun.events <- EventDown
		}
		statusUp = up

		// MTU changes
		if iface.MTU != statusMTU {
			tun.events <- EventMTUUpdate
		}
		statusMTU = iface.MTU
	}
}

func CreateTUN(name string, mtu int) (Device, error) {
	ifIndex := -1
	if name != "utun" {
		_, err := fmt.Sscanf(name, "utun%d", &ifIndex)
		if err != nil || ifIndex < 0 {
			return nil, fmt.Errorf("Interface name must be utun[0-9]*")
		}
	}

	fd, err := socketCloexec(unix.AF_SYSTEM, unix.SOCK_DGRAM, 2)
	if err != nil {
		return nil, err
	}

	ctlInfo := &unix.CtlInfo{}
	copy(ctlInfo.Name[:], []byte(utunControlName))
	err = unix.IoctlCtlInfo(fd, ctlInfo)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("IoctlGetCtlInfo: %w", err)
	}

	sc := &unix.SockaddrCtl{
		ID:   ctlInfo.Id,
		Unit: uint32(ifIndex) + 1,
	}

	err = unix.Connect(fd, sc)
	if err != nil {
		unix.Close(fd)
		return nil, err
	}

	err = unix.SetNonblock(fd, true)
	if err != nil {
		unix.Close(fd)
		return nil, err
	}
	tun, err := CreateTUNFromFile(os.NewFile(uintptr(fd), ""), mtu)

	if err == nil && name == "utun" {
		fname := os.Getenv("WG_TUN_NAME_FILE")
		if fname != "" {
			os.WriteFile(fname, []byte(tun.(*NativeTun).name+"\n"), 0o400)
		}
	}

	return tun, err
}

func CreateTUNFromFile(file *os.File, mtu int) (Device, error) {
	tun := &NativeTun{
		tunFile: file,
		events:  make(chan Event, 10),
		errors:  make(chan error, 5),
	}

	name, err := tun.Name()
	if err != nil {
		tun.tunFile.Close()
		return nil, err
	}

	tunIfindex, err := func() (int, error) {
		iface, err := net.InterfaceByName(name)
		if err != nil {
			return -1, err
		}
		return iface.Index, nil
	}()
	if err != nil {
		tun.tunFile.Close()
		return nil, err
	}

	tun.routeSocket, err = socketCloexec(unix.AF_ROUTE, unix.SOCK_RAW, unix.AF_UNSPEC)
	if err != nil {
		tun.tunFile.Close()
		return nil, err
	}

	go tun.routineRouteListener(tunIfindex)

	if mtu > 0 {
		err = tun.setMTU(mtu)
		if err != nil {
			tun.Close()
			return nil, err
		}
	}

	return tun, nil
}

func (tun *NativeTun) Name() (string, error) {
	var err error
	tun.operateOnFd(func(fd uintptr) {
		tun.name, err = unix.GetsockoptString(
			int(fd),
			2, /* #define SYSPROTO_CONTROL 2 */
			2, /* #define UTUN_OPT_IFNAME 2 */
		)
	})

	if err != nil {
		return "", fmt.Errorf("GetSockoptString: %w", err)
	}

	return tun.name, nil
}

func (tun *NativeTun) File() *os.File {
	return tun.tunFile
}

func (tun *NativeTun) Events() <-chan Event {
	return tun.events
}

func (tun *NativeTun) Read(bufs [][]byte, sizes []int, offset int) (int, error) {
	// TODO: the BSDs look very similar in Read() and Write(). They should be
	// collapsed, with platform-specific files containing the varying parts of
	// their implementations.
	select {
	case err := <-tun.errors:
		return 0, err
	default:
		buf := bufs[0][offset-4:]
		n, err := tun.tunFile.Read(buf[:])
		if n < 4 {
			return 0, err
		}
		sizes[0] = n - 4
		return 1, err
	}
}

func (tun *NativeTun) Write(bufs [][]byte, offset int) (int, error) {
	if offset < 4 {
		return 0, io.ErrShortBuffer
	}
	for i, buf := range bufs {
		buf = buf[offset-4:]
		buf[0] = 0x00
		buf[1] = 0x00
		buf[2] = 0x00
		switch buf[4] >> 4 {
		case 4:
			buf[3] = unix.AF_INET
		case 6:
			buf[3] = unix.AF_INET6
		default:
			return i, unix.EAFNOSUPPORT
		}
		if _, err := tun.tunFile.Write(buf); err != nil {
			return i, err
		}
	}
	return len(bufs), nil
}

func (tun *NativeTun) Close() error {
	var err1, err2 error
	tun.closeOnce.Do(func() {
		err1 = tun.tunFile.Close()
		if tun.routeSocket != -1 {
			unix.Shutdown(tun.routeSocket, unix.SHUT_RDWR)
			err2 = unix.Close(tun.routeSocket)
		} else if tun.events != nil {
			close(tun.events)
		}
	})
	if err1 != nil {
		return err1
	}
	return err2
}

func (tun *NativeTun) setMTU(n int) error {
	fd, err := socketCloexec(
		unix.AF_INET,
		unix.SOCK_DGRAM,
		0,
	)
	if err != nil {
		return err
	}

	defer unix.Close(fd)

	var ifr unix.IfreqMTU
	copy(ifr.Name[:], tun.name)
	ifr.MTU = int32(n)
	err = unix.IoctlSetIfreqMTU(fd, &ifr)
	if err != nil {
		return fmt.Errorf("failed to set MTU on %s: %w", tun.name, err)
	}

	return nil
}

func (tun *NativeTun) MTU() (int, error) {
	fd, err := socketCloexec(
		unix.AF_INET,
		unix.SOCK_DGRAM,
		0,
	)
	if err != nil {
		return 0, err
	}

	defer unix.Close(fd)

	ifr, err := unix.IoctlGetIfreqMTU(fd, tun.name)
	if err != nil {
		return 0, fmt.Errorf("failed to get MTU on %s: %w", tun.name, err)
	}

	return int(ifr.MTU), nil
}

func (tun *NativeTun) BatchSize() int {
	return 1
}

func socketCloexec(family, sotype, proto int) (fd int, err error) {
	// See go/src/net/sys_cloexec.go for background.
	syscall.ForkLock.RLock()
	defer syscall.ForkLock.RUnlock()

	fd, err = unix.Socket(family, sotype, proto)
	if err == nil {
		unix.CloseOnExec(fd)
	}
	return
}
