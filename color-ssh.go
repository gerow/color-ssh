package main

import "os/exec"
import "os"
import "log"
import "os/signal"
import "syscall"
import "hash/crc32"
import "fmt"
import "github.com/gerow/go-color"
import "math"

func resetColors() {
	os.Stdout.WriteString("\033]11;0\007")
	os.Stdout.WriteString("\033]10;7\007")
}

func hostColor(name string) (uint8, uint8, uint8) {
	csum := crc32.ChecksumIEEE([]byte(name))

	log.Printf("Got checksum %d\n", csum)

	var r, g, b uint8
	r = uint8(csum)
	g = uint8(csum >> 8)
	b = uint8(csum >> 16)

	return r, g, b
}

func main() {
	host := os.Args[1]

	cmd := exec.Command("ssh", host)

	r, g, b := hostColor(host)

	// Calculate the complement by converting to hsl, rotating h by 180 degrees
	// and converting back
	hsl := color.RGB{float64(r) / 255, float64(g) / 255, float64(b) / 255}.ToHSL()
	hsl.H += 0.5
	hsl.H = math.Mod(hsl.H, 1.0)
	complement := hsl.ToRGB()

	log.Printf("HSL: %d, %d, %d", hsl.H, hsl.S, hsl.L)

	fg_r := uint8(complement.R * 255)
	fg_g := uint8(complement.G * 255)
	fg_b := uint8(complement.B * 255)

	log.Printf("Setting background color to %02x%02x%02x", r, g, b)
	log.Printf("Setting foreground color to %02x%02x%02x", fg_r, fg_g, fg_b)
	// Set the background color
	fmt.Fprintf(os.Stdout, "\033]11;#%02x%02x%02x\007", r, g, b)
	fmt.Fprintf(os.Stdout, "\033]10;#%02x%02x%02x\007", fg_r, fg_g, fg_b)
	// Set up signals to restore colors when necessary
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		resetColors()
	}()

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		resetColors()
		log.Fatal(err)
	}

	// Make sure to restore it to the default color
	resetColors()
}
