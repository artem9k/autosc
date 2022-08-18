/**
automatic scheduler for cu boulder classes
**/

package main

import (
	"flag"
	"fmt"
	"sort"

	"github.com/tidwall/gjson"
)

// FUNCTIONS

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

	lec_course.Classes = lec_classes
	rec_course.Classes = rec_classes
	courses = append(courses, lec_course)
	if len(rec_course.Classes) > 0 {
		courses = append(courses, rec_course)
	}
	return courses
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
	flag.StringVar(&include_teacher_string, "include_teacher", "", "list of teachers to include, comma-separated, spaces after commas. Ex. E.Musk, K.Kardashian")
	flag.StringVar(&exclude_teacher_string, "exclude_teacher", "", "list of teachers to exclude, comma-separated, spaces after commas. Ex. E.Musk, K.Kardashian")
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

	schedules := search(courses, params)
	pprint_schedules(schedules)

	for i, schedule := range schedules {
		render(schedule, i+1)
	}
}
