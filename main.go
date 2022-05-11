package main

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
	"go4.org/media/heif/bmff"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) <= 1 {
		return nil
	}
	a := args[1]
	fmt.Fprintln(os.Stderr, a)

	f, err := os.Open(a)
	if err != nil {
		return errors.WithStack(err)
	}

	r := bmff.NewReader(f)

	for {
		box, err := r.ReadBox()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return errors.WithStack(err)
		}

		if box.Type().EqualString("gps0") {
			var gps gps0
			w := csv.NewWriter(os.Stdout)
			for {
				if err := binary.Read(box.Body(), binary.LittleEndian, &gps); err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return errors.WithStack(err)
				}
				if gps.Invalid > 0 {
					continue
				}
				w.Write([]string{
					gps.Latitude(),
					gps.Longitude(),
					fmt.Sprintf("%dm", gps.Altitude),
					fmt.Sprintf("%dkm/h", gps.Speed),
					gps.Date().In(time.Local).Format(time.RFC3339),
					fmt.Sprintf("%d°", gps.Track*2),
				})
			}
			w.Flush()
		}
	}

	return run(args[1:])
}

type gps0 struct {
	LatitudeNMEA  float64
	LongitudeNMEA float64
	Altitude      int32
	Speed         uint16
	Year          uint8
	Month         uint8
	Day           uint8
	Hour          uint8
	Minute        uint8
	Secound       uint8
	Track         uint8
	NS            uint8
	EW            uint8
	Invalid       uint8
}

func (g gps0) Latitude() string {
	deg := int(g.LatitudeNMEA / 100)
	lat := float64(deg) + (g.LatitudeNMEA-float64(deg*100))/60
	ns := g.NS
	if ns > 2 {
		ns = 0
	}
	return fmt.Sprintf(`%f%c`, lat, " NS"[ns])
}

func (g gps0) Longitude() string {
	deg := int(g.LongitudeNMEA / 100)
	lon := float64(deg) + (g.LongitudeNMEA-float64(deg*100))/60
	ew := g.EW
	if ew > 2 {
		ew = 0
	}
	return fmt.Sprintf(`%f%c`, lon, " EW"[ew])
}

func (g gps0) Date() time.Time {
	return time.Date(int(g.Year)+2000, time.Month(g.Month), int(g.Day), int(g.Hour), int(g.Minute), int(g.Secound), 0, time.UTC)
}
