package main

import (
	"context"
	"fmt"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"os"
)

func main() {
	// Create a new BLE device.
	d, err := linux.NewDevice()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create BLE device: %v\n", err)
		return
	}
	ble.SetDefaultDevice(d)

	// Scan for specified duration, or until interrupted by user.
	fmt.Println("Scanning...")
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 5*ble.Second))
	cln, err := ble.Connect(ctx, func(a ble.Advertisement) bool {
		return strings.Contains(a.LocalName(), "OxySmart") // Adjust this for your device's name
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		return
	}
	defer cln.CancelConnection()

	// Discover services and characteristics.
	fmt.Println("Discovering services and characteristics...")
	svcs, err := cln.DiscoverServices(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to discover services: %v\n", err)
		return
	}

	for _, svc := range svcs {
		chars, err := cln.DiscoverCharacteristics(nil, svc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to discover characteristics: %v\n", err)
			continue
		}

		for _, char := range chars {
			if char.UUID.Equal(ble.MustParse("6e400003-b5a3-f393-e0a9-e50e24dcca9e")) { // Nordic UART RX Characteristic UUID
				// Subscribe to notifications from the characteristic.
				err = cln.Subscribe(char, false, func(req []byte) {
					fmt.Printf("Received data: %x\n", req)
					// Here you would decode the oximeter data
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to subscribe to characteristic: %v\n", err)
					return
				}
			}
		}
	}

	// Keep the program alive to receive notifications.
	select {}
}
