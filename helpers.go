package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func print_schedules(schedules [][]Class) {
	days := [5]string{"MON", "TUE", "WED", "THU", "FRI"}
	for i, schedule := range schedules {
		fmt.Printf("OPTION %d\n", i+1)
		if len(schedule) != 0 {
			for j, class := range schedule {
				fmt.Printf("%d: %s %s %s ", j, class.Name, class.Code, class.Instructor)
				for _, constr := range class.Constraints {
					fmt.Printf("%s %d-%d ", days[constr.day], constr.start_t, constr.end_t)
				}
				fmt.Println()
			}
		}
	}
}

func print_schedules_to_file(schedules [][]Class, file string) {
	days := [5]string{"MON", "TUE", "WED", "THU", "FRI"}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	for i, schedule := range schedules {
		fmt.Fprintf(f, "OPTION %d\n", i+1)
		if len(schedule) != 0 {
			for j, class := range schedule {
				fmt.Fprintf(f, "%d: %s %s %s ", j, class.Name, class.Code, class.Instructor)
				for _, constr := range class.Constraints {
					fmt.Fprintf(f, "%s %d-%d ", days[constr.day], constr.start_t, constr.end_t)
				}
				fmt.Fprintln(f)
			}
		}
	}
}

func get_safe_atoi(num string) int {
	var val, err = strconv.Atoi(num)
	if err != nil {
		panic(err)
	}
	return val
}

func parse_string_time(time string) int {
	parsed_time := strings.ReplaceAll(time, ":", "")
	int_time, err := strconv.Atoi(parsed_time)
	if err != nil {
		panic(err)
	}
	return int_time
}

func parse_string_time_slice(time_slice string) []int {
	tuple := make([]int, 2)
	parsed_time_slice := strings.ReplaceAll(time_slice, ":", "")
	times := strings.Split(parsed_time_slice, "-")
	for i := 0; i < 2; i++ {
		time_int, err := strconv.Atoi(times[i])
		if err != nil {
			panic(err)
		}
		tuple[i] = time_int
	}
	return tuple

}

func create_query_body(class string) string {
	var body string = "{\"other\":{\"srcdb\":\"2227\"},\"criteria\":[{\"field\":\"alias\",\"value\":\"" + class + "\"}]}"
	return body
}

func ping_classes(_url, body string) string {
	res, _ := http.Post(_url, "application/json", bytes.NewBufferString(body))
	defer res.Body.Close()
	res_body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var res_string = string(res_body)
	return res_string
}

func create_courses_list(infile string) []string {
	list := make([]string, 0)
	file, err := os.Open(infile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}
	return list
}
