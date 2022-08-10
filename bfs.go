/**
automatic scheduler for cu boulder classes (or honestly any other school)
**/

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// GLOBALS

// CLASSES

type Constraint struct {
	// A class will have several of these, for each day
	day     int
	start_t int
	end_t   int
}

type Class struct {
	// represents a single class instance
	// starting and ending at a certain time
	// lect and rec are separate Class instances,
	// however they will have different types set
	Code        string
	Name        string
	Instructor  string
	CreditHours int
	Type        string
	Constraints []Constraint
}

type Course struct {
	// A course, ex PHYS 1110, has a list of specific Classes
	// to choose from, each with its own instructor and times
	Code    string
	Name    string
	Type    string
	Classes []Class
}

type Globals struct {
	// i use a struct for globals and params, since we have a lot of them
	days_in_week         int
	mid_day              int
	time_between_classes int
	min_time             int
	max_time             int
	break_start          int
	break_end            int
	topk                 int

	rank_by_mid_day bool // add score by a function of the mid_day
	rank_by_teacher bool // add score if the teacher is in prioritize_teacher

	exclude_teacher    []string
	prioritize_teacher []string
}

func (g Globals) init() {
	g.days_in_week = 7
	g.mid_day = 1200
	g.time_between_classes = 0
	g.min_time = 0
	g.max_time = 2400
	g.break_start = 0
	g.break_end = 0
	g.topk = 10
	g.rank_by_mid_day = false
	g.rank_by_teacher = false
	g.exclude_teacher = nil
	g.prioritize_teacher = nil
}

// FUNCTIONS

func get_safe_atoi(num string) int {
	var val, err = strconv.Atoi(num)
	if err != nil {
		panic(err)
	}
	return val
}

func create_class(data string) Class {
	// todo handle if no results
	var cart_opts_string = gjson.Get(data, "cart_opts").Str
	var meeting_times_str = gjson.Parse(gjson.Get(data, "meetingTimes").Str)
	var instr = gjson.Get(data, "instr").Str
	if instr == "" {
		instr = "TA"
	}

	var _type = gjson.Get(data, "schd").Str
	var code = gjson.Get(data, "code").Str
	var name = gjson.Get(data, "title").Str
	var credit_hours_string = gjson.Get(cart_opts_string, "credit_hrs.options.0.label").Str
	var credit_hours = get_safe_atoi(credit_hours_string)

	var new_class Class
	new_class.Instructor = instr
	new_class.Type = _type
	new_class.Code = code
	new_class.Name = name

	new_class.CreditHours = credit_hours

	constraints := make([]Constraint, 0)

	meeting_times_str.ForEach(func(key, value gjson.Result) bool {
		var meet_day = get_safe_atoi(value.Get("meet_day").Str)
		var start_time = get_safe_atoi(value.Get("start_time").Str)
		var end_time = get_safe_atoi(value.Get("end_time").Str)
		new_constr := Constraint{meet_day, start_time, end_time}
		constraints = append(constraints, new_constr)
		return true
	})
	new_class.Constraints = constraints
	return new_class
}

func get_med_time(a, b int) int {
	return (a + b) / 2
}

func get_mid_day_score(a int) int {
	return int(math.Abs(float64(1200 - a)))
}

func create_course(data string, params Globals) []Course {

	// return a slice with either 1 or 2 courss, depending on whether the course
	// needs a recitation

	courses := make([]Course, 0)
	var lec_course Course
	var rec_course Course

	var code = gjson.Get(data, "results.0.code").Str
	var name = gjson.Get(data, "results.0.title").Str

	lec_course.Code = code
	lec_course.Name = name
	lec_course.Type = "LEC"
	rec_course.Code = code
	rec_course.Name = name
	rec_course.Type = "REC"

	lec_classes := make([]Class, 0)
	rec_classes := make([]Class, 0)

	var course_list = gjson.Get(data, "results")
	// get name and code from first class
	course_list.ForEach(func(key, value gjson.Result) bool {
		class_str := value.String()
		new_class := create_class(class_str)
		if new_class.Type == "LEC" {
			lec_classes = append(lec_classes, new_class)
		} else if new_class.Type == "REC" {
			rec_classes = append(rec_classes, new_class)
		} else {
			fmt.Println("unknown class type: ", new_class.Type)
		}
		return true
	})
	if params.rank_by_mid_day {
		sort.SliceStable(lec_classes, func(i, j int) bool {
			a := lec_classes[i]
			b := lec_classes[j]
			sum_a := 0
			sum_b := 0

			for _, constr := range a.Constraints {
				sum_a += get_mid_day_score(get_med_time(constr.start_t, constr.end_t))
			}

			for _, constr := range b.Constraints {
				sum_b += get_mid_day_score(get_med_time(constr.start_t, constr.end_t))
			}

			return sum_a < sum_b
		})
		if len(rec_classes) > 0 {
			sort.SliceStable(rec_classes, func(i, j int) bool {
				// i < j
				a := rec_classes[i]
				b := rec_classes[j]
				sum_a := 0
				sum_b := 0

				for _, constr := range a.Constraints {
					sum_a += get_mid_day_score(get_med_time(constr.start_t, constr.end_t))
				}

				for _, constr := range b.Constraints {
					sum_b += get_mid_day_score(get_med_time(constr.start_t, constr.end_t))
				}

				return sum_a < sum_b
			})
		}
	}
	if params.rank_by_teacher {
		// TODO
	}

	lec_course.Classes = lec_classes
	rec_course.Classes = rec_classes
	courses = append(courses, lec_course)
	if len(rec_course.Classes) > 0 {
		courses = append(courses, rec_course)
	}
	return courses
}

func check_constraint_overlap(c1, c2 Constraint) bool {
	var start_overlap_1 = (c2.start_t < c1.start_t && c1.start_t < c2.end_t)
	var start_overlap_2 = (c2.start_t < c2.start_t && c2.start_t < c2.end_t)
	var end_overlap_1 = (c2.start_t < c1.end_t && c1.end_t < c2.end_t)
	var end_overlap_2 = (c1.start_t < c2.end_t && c2.end_t < c1.end_t)

	return !(end_overlap_1 || end_overlap_2 || start_overlap_1 || start_overlap_2)
}

func check_class_overlap(cls1, cls2 Class) bool {
	if len(cls1.Constraints) == 0 || len(cls2.Constraints) == 0 {
		return true
	}
	one_ptr := 0
	two_ptr := 0
	var cnstr1 = cls1.Constraints
	var cnstr2 = cls2.Constraints
	// the constraints are already sorted by increasing day
	for i := 0; i < 7; i++ {
		if one_ptr == len(cnstr1) || two_ptr == len(cnstr2) {
			return true
		}
		d_1 := cnstr1[one_ptr].day
		d_2 := cnstr2[two_ptr].day
		if d_1 == i && d_2 == i {
			return check_constraint_overlap(cnstr1[one_ptr], cnstr2[two_ptr])
		} else if cnstr1[one_ptr].day == i {
			one_ptr += 1
		} else if cnstr2[two_ptr].day == i {
			two_ptr += 1
		}
	}
	return true
}

func pop(stack []interface{}) []interface{} {
	return stack[:len(stack)-1]
}

func push(stack []interface{}, new_item interface{}) []interface{} {
	return append(stack, new_item)
}

func check_class_against_list(classes_list []Class, new_class Class) bool {
	if len(classes_list) == 0 {
		return true
	}
	for _, class := range classes_list {
		if !check_class_overlap(class, new_class) {
			return false
		}
	}
	return true
}

func DFS_recursive(options []Course, solution []Class, i int) []Class {
	// search complete
	if i == len(options) {
		return solution
	}
	for _, new_class := range options[i].Classes {
		// compare the class with our constr
		check := check_class_against_list(solution, new_class)
		if check {
			next := DFS_recursive(options, append(solution, new_class), i+1)
			if next != nil {
				return next
			}
		}
	}
	return nil
}

func search(options []Course, params Globals) []Class {
	result := DFS_recursive(options, make([]Class, 0), 0)
	return result
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

func main() {

	var infile string
	var outfile string
	var topk int
	var min_buffer int
	var max_time_string string
	var min_time_string string
	var break_time_string string
	var include_teacher_string string
	var exclude_teacher_string string
	var mid_day string
	var rank_by_teacher bool
	var rank_by_mid_day bool
	var _url = "https://classes.colorado.edu/api/?page=fose&route=search"

	flag.IntVar(&min_buffer, "min_buffer", 0, "Minimum time between classes, in minutes. Default is 0")
	flag.IntVar(&topk, "topk", 10, "list of possible schedules to return, ranked by fit. Default is 10")

	flag.StringVar(&max_time_string, "max_time", "", "latest time for any class. 24-hr format. Ex. 20:00")
	flag.StringVar(&min_time_string, "min_time", "", "earliest time for any class. 24-hr format. Ex. 8:00.")
	flag.StringVar(&break_time_string, "break_time", "", "time reserved for a break, or lunch. 24-hr format. Ex. 10:00-11:00")
	flag.StringVar(&infile, "infile", "input.txt", "input file with each class on a new line.")
	flag.StringVar(&outfile, "outfile", "output.txt", "output file with a list of new schedules")
	flag.StringVar(&outfile, "outfile", "output.txt", "output file with a list of new schedules")
	flag.StringVar(&include_teacher_string, "rank_by_teacher", "", "list of teachers to include, comma-separated, spaces after commas. Ex. E.Musk, K.Kardashian")
	flag.StringVar(&exclude_teacher_string, "rank_by_teacher", "", "list of teachers to exclude, comma-separated, spaces after commas. Ex. E.Musk, K.Kardashian")
	flag.StringVar(&outfile, "outfile", "output.txt", "output file with a list of new schedules")
	flag.StringVar(&mid_day, "mid_day", "12:00", "time of the day to prioritize classes from. 24-hr format. Ex. 12:00")

	flag.BoolVar(&rank_by_teacher, "rank_by_teacher", false, "whether to rank results by included/excluded teachers. true/false. For this to work, you must specify lists of teachers to include/exclude using --include_teacher or --exclude_teacher.")
	flag.BoolVar(&rank_by_mid_day, "rank_by_mid_day", false, "whether to prioritize classes from around the time specified by --mid_day (default is 12:00)")

	flag.Parse()

	var max_time *int = nil
	var min_time *int = nil
	var break_time *[]int = nil

	if max_time_string != "" {
		*max_time = parse_string_time(max_time_string)
	}

	if min_time_string != "" {
		*min_time = parse_string_time(min_time_string)
	}

	if break_time_string != "" {
		*break_time = parse_string_time_slice(break_time_string)
	}

	var params Globals
	params.init()
	if max_time != nil {
		params.max_time = *max_time
	}
	if min_time != nil {
		params.min_time = *min_time
	}
	if break_time != nil {
		params.break_start = (*break_time)[0]

	}

	// search is applied as we create list
	courses_list := create_courses_list(infile)
	courses := make([]Course, 0)

	for _, course_name := range courses_list {
		body := create_query_body(course_name)
		res_string := ping_classes(_url, body)
		courses = append(courses, create_course(res_string, params)...)
	}

	sched := search(courses, params)

	for i, class := range sched {
		fmt.Println(i, ":")
		fmt.Println(class)
	}

}
