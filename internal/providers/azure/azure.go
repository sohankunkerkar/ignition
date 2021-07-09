// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The azure provider fetches a configuration from the Azure OVF DVD.

package azure

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/ignition/v2/config/v3_4_experimental/types"
	execUtil "github.com/coreos/ignition/v2/internal/exec/util"
	"github.com/coreos/ignition/v2/internal/providers/util"
	"github.com/coreos/ignition/v2/internal/resource"

	"github.com/coreos/vcontext/report"
	"golang.org/x/sys/unix"
)

const (
	configPath = "/CustomData.bin"
)

// These constants come from <cdrom.h>.
const (
	CDROM_DRIVE_STATUS = 0x5326
)

// These constants come from <cdrom.h>.
const (
	CDS_NO_INFO = iota
	CDS_NO_DISC
	CDS_TRAY_OPEN
	CDS_DRIVE_NOT_READY
	CDS_DISC_OK
)

// Azure uses a UDF volume for the OVF configuration.
const (
	CDS_FSTYPE_UDF = "udf"
)

// FetchConfig wraps FetchOvfDevice to implement the platform.NewFetcher interface.
func FetchConfig(f *resource.Fetcher) (types.Config, report.Report, error) {
	return FetchFromOvfDevice(f, []string{CDS_FSTYPE_UDF})
}

// FetchFromOvfDevice has the return signature of platform.NewFetcher. It is
// wrapped by this and AzureStack packages.
func FetchFromOvfDevice(f *resource.Fetcher, ovfFsTypes []string) (types.Config, report.Report, error) {
	var rawConfig []byte
	logger := f.Logger

	device, err := execUtil.GetBlockDevices(CDS_FSTYPE_UDF)
	if err != nil {
		return types.Config{}, report.Report{}, fmt.Errorf("failed to retrieve block devices with FSTYPE=UDF: %v", err)
	} else if len(device) > 0 {
		for i, dev := range device {
			rawConfig, err = getRawConfig(f, dev)
			if len(rawConfig) > 0 {
				break
			} else if err != nil && i == len(device) {
				return types.Config{}, report.Report{}, fmt.Errorf("failed to retrieve config: %v", err)
			}
		}
	}
	return util.ParseConfig(logger, rawConfig)
}

// getRawConfig returns the config by mounting the given block device
func getRawConfig(f *resource.Fetcher, devicePath string) ([]byte, error) {
	logger := f.Logger
	mnt, err := ioutil.TempDir("", "ignition-azure")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.Remove(mnt)

	logger.Debug("mounting config device")
	if err := logger.LogOp(
		func() error { return unix.Mount(devicePath, mnt, CDS_FSTYPE_UDF, unix.MS_RDONLY, "") },
		"mounting %q at %q", devicePath, mnt,
	); err != nil {
		return nil, fmt.Errorf("failed to mount device %q at %q: %v", devicePath, mnt, err)
	}
	defer func() {
		_ = logger.LogOp(
			func() error { return unix.Unmount(mnt, 0) },
			"unmounting %q at %q", devicePath, mnt,
		)
	}()

	logger.Debug("reading config")
	rawConfig, err := ioutil.ReadFile(filepath.Join(mnt, configPath))
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}
	return rawConfig, nil
}
