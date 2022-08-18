package main

import (
	"fmt"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

var lines = []string{
	"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod",
	"tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam,",
	"quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo",
	"consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse",
	"cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat",
	"non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
}

// given a schedule, render a calendar.
func render(classes []Class) {

	// width := 0
	// height := 0
	// num_days := 7
	// 90 pixels for each rectangle width
	// 480. 30 pixels for each hour, since 6am-10pm

	const W = 1920
	const H = 1920
	const exp_w = W / 7
	const exp_h = H / 16
	const global_offset_x = 200
	const global_offset_y = 10
	const middle_of_screen = W / 2

	dc := gg.NewContext(W, H)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}

	big_face := truetype.NewFace(font, &truetype.Options{
		Size: 40,
	})

	small_face := truetype.NewFace(font, &truetype.Options{
		Size: 25,
	})

	tiny_face := truetype.NewFace(font, &truetype.Options{
		Size: 15,
	})

	days := [7]string{"MON", "TUE", "WED", "THU", "FRI"}
	text := "Schedule 1"

	// draw days of week
	dc.SetFontFace(small_face)

	for i := 0; i < 5; i++ {
		var x_l float64 = float64(i*exp_w + (exp_w / 2))
		var x float64 = float64(i*exp_w + exp_w)
		y := 60.
		dc.DrawStringAnchored(days[i], x+global_offset_x, y+global_offset_y, 0.5, 0.5)
		dc.DrawLine(x_l+global_offset_x, 110, x_l+global_offset_x, 1810)
	}

	// draw lines

	for time := 600; time <= 2200; time += 100 {
		y := float64(100. + (time - 600))
		atoi_time := strconv.Itoa(time)
		l := len(atoi_time)
		atoi_time = atoi_time[:l-2] + ":" + atoi_time[l-2:]
		x := 50.
		dc.DrawStringAnchored(atoi_time, x+global_offset_x, y+global_offset_y, 0.5, 0.5)
	}

	for i := 0; i < 5; i++ {
		for time := 600; time <= 2200; time += 100 {
			var x0 float64 = float64(i*exp_w + (exp_w / 2))
			x1 := x0 + exp_w
			y := float64(100. + (time - 600))
			dc.DrawLine(x0+global_offset_x, y+global_offset_y, x1+global_offset_x, y+global_offset_y)
		}
	}
	dc.SetRGB(0, 0, 0)
	dc.Stroke()

	dc.SetFontFace(big_face)
	dc.DrawStringAnchored(text, middle_of_screen, 30, 0.5, 0.5)

	color_lec := []float64{200, 79, 30} // solid blue
	color_rec := []float64{79, 79, 79}
	color_other := []float64{70, 200, 30}

	for _, cls := range classes {

		instr := cls.Instructor
		code := cls.Code
		_type := cls.Type
		color := color_lec

		switch _type {
		case "LEC":
			color = color_lec
		case "REC":
			color = color_rec
		default:
			color = color_other

		}

		for _, constr := range cls.Constraints {
			day := constr.day
			start_t := constr.start_t
			end_t := constr.end_t

			x := float64(day*exp_w + (exp_w / 2))
			y_1 := float64(100 + (start_t - 600))
			y_2 := float64(100 + (end_t - 600))
			y := y_1
			l := y_2 - y_1
			fmt.Println(int(y_2 / 100))
			fmt.Println(int(y_1 / 100))
			if int(y_2/100) == int(y_1/100) {
				l = l * 100 / 60
			}

			dc.SetRGB(color[0], color[1], color[2])

			dc.DrawRoundedRectangle(x+global_offset_x, y+global_offset_y, float64(exp_w), l, 10)
			dc.Fill()
			dc.SetFontFace(tiny_face)
			dc.SetRGB(0, 0, 0)
			dc.DrawStringAnchored(code, x+global_offset_x+10, y+global_offset_y+20, 0, 0)
			dc.DrawStringAnchored(_type, x+global_offset_x+10, y+global_offset_y+40, 0, 0)
			dc.DrawStringAnchored(instr, x+global_offset_x+10, y+global_offset_y+60, 0, 0)

		}

	}

	dc.SavePNG("out.png")

}

// alternatively, pretty-print to a file.
func print_to_file(cls []Class) {

}
