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
import "strings"

func resetColors() {
	os.Stdout.WriteString("\033]11;0\007")
	os.Stdout.WriteString("\033]10;7\007")
}

func hostColors(name string) (color.HSL, color.HSL) {
	csum := crc32.ChecksumIEEE([]byte(name))

	log.Printf("Got checksum %d\n", csum)

	var r, g, b, levelmod uint8

	r = uint8(csum)
	g = uint8(csum >> 8)
	b = uint8(csum >> 16)
	levelmod = uint8(csum >> 24)

	foreground := color.RGB{float64(r) / 255, float64(g) / 255, float64(b) / 255}.ToHSL()

	if (csum >> 25) & 0x1 != 0 {
		// if the 25th bit is set then this is a dark color, otherwise it is light
		foreground.L = 0.0 + 0.25 * (float64(levelmod) / 255)
	} else {
		foreground.L = 0.75 + 0.25 * (float64(levelmod) / 255)
	}

	// the background should be the complement of the foreground color with the oppisite darkness
	background := color.HSL{math.Mod(foreground.H + 0.5, 1.0), foreground.S, foreground.L}
	if foreground.L > 0.5 {
		background.L = 0.0 + 0.40 * (float64(levelmod) / 255);
	} else {
		background.L = 0.60 + 0.40 * (float64(levelmod) / 255);
	}

	return  foreground, background
}

func extractHostname(args []string) string {
	for _,e := range args {
		if strings.Contains(e, "@") {
			e = strings.Split(e, "@")[1]
		}
		if strings.Contains(e, ".") {
			return e;
		}
	}
	log.Printf("failed to find hostname in args %v\n", args)
	return ""
}

func main() {
	host := extractHostname(os.Args)

	cmd := exec.Command("ssh", os.Args[1:]...)

	fg_hsl, bg_hsl := hostColors(host)

	fg := fg_hsl.ToRGB()
	bg := bg_hsl.ToRGB()

	bg_r := uint8(bg.R * 255)
	bg_g := uint8(bg.G * 255)
	bg_b := uint8(bg.B * 255)

	fg_r := uint8(fg.R * 255)
	fg_g := uint8(fg.G * 255)
	fg_b := uint8(fg.B * 255)

	log.Printf("using bg{%d, %d, %d} fg{%d, %d, %d}", bg_r, bg_b, bg_b, fg_r, fg_b, fg_b)

	// Set the background color
	fmt.Fprintf(os.Stdout, "\033]11;#%02x%02x%02x\007", bg_r, bg_g, bg_b)
	// Set the foreground color
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
