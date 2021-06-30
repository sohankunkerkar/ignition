package main

import (
	"fmt"

	"github.com/coreos/ignition/v2/internal/exec/util"
)

func main() {
	blkDeviceList, err := util.GetUdfBlockDevices()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("length of the list: %d\n", len(blkDeviceList))

	for _, blk := range blkDeviceList {
		fmt.Printf("Device name : %s", blk)
	}
}
