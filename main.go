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
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/containerd/typeurl"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/urfave/cli"
)

var (
	Usage = "wasm-oci processor is used to process wasm-oci images."
)

func main() {
	app := cli.NewApp()
	app.Name = "wasm-oci-processor"
	app.Usage = Usage
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "parse-annotations",
			Usage: "parse annotations and do stuff. (optional)",
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx *cli.Context) error {
	setupDebuggerEvent()
	if err := process(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
		return err
	}
	return nil
}

func process(ctx *cli.Context) error {
	_, err := getPayload()
	if err != nil {
		return err
	}

	var newTarBuffer bytes.Buffer
	newTarWriter := tar.NewWriter(&newTarBuffer)

	// for testing locally
	// flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	// file, err := os.OpenFile("archive.tar", flags, 0644)
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()
	// newTarWriter := tar.NewWriter(file)

	tr := tar.NewReader(os.Stdin)
	// for testing locally.
	// f, err := os.OpenFile("layer.tar", os.O_RDWR, os.ModePerm)
	// defer f.Close()
	// if err != nil {
	// 	os.Exit(1)
	// }
	// tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s test\n", err)
			os.Exit(1)
		}

		if strings.HasSuffix(hdr.Name, ".wasm") {
			//fmt.Printf("Found Wasm File, moving to correct location %s:\n", hdr.Name)
			createFolderHeader(newTarWriter, "Files")
			createFolderHeader(newTarWriter, "Files/Windows")
			createFolderHeader(newTarWriter, "Files/Windows/System32")
			createFolderHeader(newTarWriter, "Files/Windows/System32/config")
			createFile(newTarWriter, "Files/Windows/System32/config/DEFAULT")
			createFile(newTarWriter, "Files/Windows/System32/config/SAM")
			createFile(newTarWriter, "Files/Windows/System32/config/SECURITY")
			createFile(newTarWriter, "Files/Windows/System32/config/SOFTWARE")
			createFile(newTarWriter, "Files/Windows/System32/config/SYSTEM")

			hdr.Name = "Files/" + hdr.Name
			if err := newTarWriter.WriteHeader(hdr); err != nil {
				os.Exit(1)
			}

			if _, err := io.Copy(newTarWriter, tr); err != nil {
				os.Exit(1)
			}
			continue
		}

		//fmt.Printf("writing file with name %s:\n", hdr.Name)
		if err := newTarWriter.WriteHeader(hdr); err != nil {
			os.Exit(1)
		}

		if _, err := io.Copy(newTarWriter, tr); err != nil {
			os.Exit(1)
		}
	}
	if err := newTarWriter.Close(); err != nil {
		os.Exit(1)
	}

	_, err = io.Copy(os.Stdout, &newTarBuffer)
	if err != nil {
		return fmt.Errorf("could not copy data: %w", err)
	}

	return nil
}

func createFolderHeader(tw *tar.Writer, name string) {
	system32 := &tar.Header{
		Name:     name,
		Typeflag: tar.TypeDir,
	}

	if err := tw.WriteHeader(system32); err != nil {

		os.Exit(1)
	}
}

func createFile(tw *tar.Writer, name string) {
	system32 := &tar.Header{
		Name: name,
		Mode: 0600,
		Size: int64(len("")),
	}

	if err := tw.WriteHeader(system32); err != nil {
		os.Exit(1)
	}

	if _, err := tw.Write([]byte("")); err != nil {
		os.Exit(1)
	}
}

func getPayload() (*Payload, error) {
	data, err := readPayload()
	if err != nil {
		return nil, fmt.Errorf("read payload: %w", err)
	}

	if data == nil {
		// no payload passed.
		return nil, nil
	}

	var anything types.Any
	if err := proto.Unmarshal(data, &anything); err != nil {
		return nil, fmt.Errorf("could not proto.Unmarshal() decrypt data: %w", err)
	}
	v, err := typeurl.UnmarshalAny(&anything)
	if err != nil {
		return nil, fmt.Errorf("could not UnmarshalAny() the decrypt data: %w", err)
	}
	l, ok := v.(*Payload)
	if !ok {
		return nil, fmt.Errorf("unknown payload type %s", anything.TypeUrl)
	}
	return l, nil
}

const (
	PayloadURI = "io.containerd.ociwasm.v1.Payload"
)

func init() {
	typeurl.Register(&Payload{}, PayloadURI)
}

// Payload holds data
type Payload struct {
	Descriptor ocispec.Descriptor
}
