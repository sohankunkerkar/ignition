package main

import (
	"fmt"

	"github.com/coreos/ignition/v2/internal/exec/util"
)

func main() {
	blkDeviceList, err := util.GetUdfBlockDevices()
	if err == nil {
		fmt.Println("Error")
	}

	fmt.Printf("length of a list: %d\n", len(blkDeviceList))

	for _, blk := range blkDeviceList {
		fmt.Printf("Device list : %s", blk)
	}
	fmt.Println("Sohan")
}
