//go:build windows
// +build windows

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	winio "github.com/Microsoft/go-winio"
	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

func readPayload() ([]byte, error) {
	path := os.Getenv("STREAM_PROCESSOR_PIPE")

	if path == "" {
		return nil, nil
	}

	conn, err := winio.DialPipe(path, nil)
	if err != nil {
		return nil, fmt.Errorf("could not DialPipe: %w", err)
	}
	defer conn.Close()
	return io.ReadAll(conn)
}

// createEvent creates a Windows event ACL'd to builtin administrator
// and local system. Can use docker-signal to signal the event.
func createEvent(event string) (windows.Handle, error) {
	ev, _ := windows.UTF16PtrFromString(event)
	sd, err := windows.SecurityDescriptorFromString("D:P(A;;GA;;;BA)(A;;GA;;;SY)")
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get security descriptor for event '%s'", event)
	}
	var sa windows.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 1
	sa.SecurityDescriptor = sd
	h, err := windows.CreateEvent(&sa, 0, 0, ev)
	if h == 0 || err != nil {
		return 0, errors.Wrapf(err, "failed to create event '%s'", event)
	}
	return h, nil
}

// setupDebuggerEvent listens for an event to allow a debugger such as delve
// to attach for advanced debugging. It's called when handling a ContainerCreate
func setupDebuggerEvent() {
	if os.Getenv("DEBUG_OCIWASM") == "" {
		return
	}
	event := "Global\\debugger-" + fmt.Sprint(os.Getpid())
	handle, err := createEvent(event)
	if err != nil {
		return
	}
	_, _ = windows.WaitForSingleObject(handle, windows.INFINITE)
}
