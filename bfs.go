/**
automatic scheduler for cu boulder classes
**/

package main

import (
	"flag"
	"fmt"
	"reflect"
	"sort"
	"strings"

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

func check_class_against_solutions(curr_solution []Class, solutions [][]Class) bool {
	for _, s := range solutions {
		if reflect.DeepEqual(s, curr_solution) {
			return false
		}
		if len(s) == 0 {
			return true
		}
	}
	return true
}

func check_class_exclude_list(cls Class, params Globals) bool {
	for _, teacher := range params.exclude_list {
		if teacher == cls.Instructor {
			return false
		}
	}
	return true
}

func check_min_time(cls Class, params Globals) bool {
	min_time := params.min_time
	for _, constr := range cls.Constraints {
		if constr.start_t < min_time {
			return false
		}
	}

	return true
}

func check_max_time(cls Class, params Globals) bool {
	max_time := params.max_time
	for _, constr := range cls.Constraints {
		if constr.end_t > max_time {
			return false
		}
	}

	return true
}

func check_break_overlap(c1 Constraint, break_start, break_end int) bool {

	// special case where they start at the same time
	var eq = break_start == c1.start_t && break_end == c1.end_t
	// break starts inside a class
	var overlap_1 = (break_start > c1.start_t && break_end > c1.end_t)
	// break ends inside a class
	var overlap_2 = (break_start < c1.start_t && break_end < c1.end_t)
	// class is inside break
	var overlap_3 = (break_start < c1.start_t && break_end > c1.end_t)
	// break is inside class
	var overlap_4 = (break_start > c1.start_t && break_end < c1.end_t)

	return !(overlap_1 || overlap_2 || overlap_3 || overlap_4 || eq)
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

	lec_classes_pruned := make([]Class, 0)
	rec_classes_pruned := make([]Class, 0)

	for _, class := range lec_classes {
		check := true
		if params.exclude_list != nil {
			check = check && check_class_exclude_list(class, params)
		}
		if params.min_time > 0 {
			check = check && check_min_time(class, params)
		}
		if params.max_time < 2400 {
			check = check && check_max_time(class, params)
		}
		if params.break_end != 0 {
			for _, constr := range class.Constraints {
				check = check && check_break_overlap(constr, params.break_start, params.break_end)
			}
		}
		if check {
			lec_classes_pruned = append(lec_classes_pruned, class)
		}
	}

	lec_classes = lec_classes_pruned
	if len(rec_classes) > 0 {
		for _, class := range rec_classes {
			check := true
			if params.exclude_list != nil {
				check = check && check_class_exclude_list(class, params)
			}
			if params.min_time > 0 {
				check = check && check_min_time(class, params)
			}
			if params.max_time < 2400 {
				check = check && check_max_time(class, params)
			}
			if params.break_end != 0 {
				for _, constr := range class.Constraints {
					check = check && check_break_overlap(constr, params.break_start, params.break_end)
				}
			}
			if check {
				rec_classes_pruned = append(rec_classes_pruned, class)
			}
		}
		rec_classes = rec_classes_pruned
	}

	if params.rank_by_mid_day {
		sort.SliceStable(lec_classes, func(i, j int) bool {
			a := lec_classes[i]
			b := lec_classes[j]
			score_a := 0
			score_b := 0

			for _, constr := range a.Constraints {
				score_a -= get_mid_day_score(get_med_time(constr.start_t, constr.end_t))
			}

			for _, constr := range b.Constraints {
				score_b -= get_mid_day_score(get_med_time(constr.start_t, constr.end_t))
			}

			if params.rank_by_teacher {
				for _, teacher := range params.include_list {
					if lec_classes[i].Instructor == teacher {
						score_a += 750
					}
					if lec_classes[j].Instructor == teacher {
						score_b += 750
					}
				}
			}

			return score_a > score_b
		})

		if len(rec_classes) > 0 {
			sort.SliceStable(rec_classes, func(i, j int) bool {
				// i < j
				a := rec_classes[i]
				b := rec_classes[j]
				score_a := 0
				score_b := 0

				for _, constr := range a.Constraints {
					score_a -= get_mid_day_score(get_med_time(constr.start_t, constr.end_t))
				}

				for _, constr := range b.Constraints {
					score_b -= get_mid_day_score(get_med_time(constr.start_t, constr.end_t))
				}

				if params.rank_by_teacher {
					for _, teacher := range params.include_list {
						if rec_classes[i].Instructor == teacher {
							score_a += 750
						}
						if rec_classes[j].Instructor == teacher {
							score_b += 750
						}
					}
				}

				return score_a > score_b
			})
		}
	}

	lec_course.Classes = lec_classes
	rec_course.Classes = rec_classes
	courses = append(courses, lec_course)
	if len(rec_course.Classes) > 0 {
		courses = append(courses, rec_course)
	}
	return courses
}

func parse_teacher_string(s string) []string {
	fmt.Println(s)
	split := strings.Split(s, ",")
	fmt.Println(split)
	return split
}

func main() {

	var infile string                 // input file of newline-separated class names
	var outfile string                // file printed to when print_to_outfile is enabled. default outfile.txt
	var topk int                      // how many schedules to generate
	var time_between_classes int      // minimum time between classes
	var min_time_string string        // earliest time
	var max_time_string string        // latest time
	var break_time_string string      // time for break, separated by a -
	var include_teacher_string string // teachers to rank highly
	var exclude_teacher_string string // teachers to exclude
	var mid_day_string string         // which time to cluster classes near
	var rank_by_teacher bool          // whether to rank by teachers in include_teachers_string
	var rank_by_mid_day bool          // whether to bias the classes by the mid_day variable. Defaults to True
	var print_to_outfile bool         // print the command-line output to outfile.txt. Defaults to false
	var draw_schedules bool           // make pngs of schedules. Defaults to true
	var _url = "https://classes.colorado.edu/api/?page=fose&route=search"

	flag.IntVar(&time_between_classes, "min_buffer", 0, "Minimum time between classes, in minutes. Default is 0")
	flag.IntVar(&topk, "topk", 10, "list of possible schedules to return, ranked by fit. Default is 10")

	flag.StringVar(&max_time_string, "max_time", "", "latest time for any class. 24-hr format. Ex. 20:00")
	flag.StringVar(&min_time_string, "min_time", "", "earliest time for any class. 24-hr format. Ex. 8:00.")
	flag.StringVar(&break_time_string, "break_time", "", "time reserved for a break, or lunch. 24-hr format. Ex. 10:00-11:00")
	flag.StringVar(&infile, "infile", "input.txt", "input file with each class on a new line.")
	flag.StringVar(&outfile, "outfile", "output.txt", "output file with a list of new schedules")
	flag.StringVar(&include_teacher_string, "include_teacher", "", "list of teachers to include, comma-separated, spaces after commas. Ex. E.Musk, K.Kardashian")
	flag.StringVar(&exclude_teacher_string, "exclude_teacher", "", "list of teachers to exclude, comma-separated, spaces after commas. Ex. E.Musk, K.Kardashian")
	flag.StringVar(&mid_day_string, "mid_day", "", "time of the day to prioritize classes from. 24-hr format. Ex. 12:00")

	flag.BoolVar(&rank_by_teacher, "rank_by_teacher", false, "whether to rank results by included/excluded teachers. true/false. For this to work, you must specify lists of teachers to include/exclude using --include_teacher or --exclude_teacher.")
	flag.BoolVar(&rank_by_mid_day, "rank_by_mid_day", true, "whether to prioritize classes from around the time specified by --mid_day (default is 12:00)")

	flag.Parse()

	var max_time *int = nil
	var min_time *int = nil
	var mid_day *int = nil

	var break_time *[]int = nil
	var include_teacher []string = nil
	var exclude_teacher []string = nil

	if max_time_string != "" {
		*max_time = parse_string_time(max_time_string)
	}

	if min_time_string != "" {
		*min_time = parse_string_time(min_time_string)
	}

	if mid_day_string != "" {
		*mid_day = parse_string_time(mid_day_string)
	}

	if break_time_string != "" {
		*break_time = parse_string_time_slice(break_time_string)
	}

	if include_teacher_string != "" {
		include_teacher = parse_teacher_string(include_teacher_string)
	}

	if exclude_teacher_string != "" {
		exclude_teacher = parse_teacher_string(include_teacher_string)
	}

	params := globals_init()
	params.rank_by_mid_day = rank_by_mid_day
	params.rank_by_teacher = rank_by_teacher
	params.include_list = include_teacher
	params.exclude_list = exclude_teacher
	params.time_between_classes = time_between_classes
	params.topk = topk

	if max_time != nil {
		params.max_time = *max_time
	}

	if min_time != nil {
		params.min_time = *min_time
	}

	if mid_day != nil {
		params.mid_day = *mid_day
	}

	if break_time != nil {
		params.break_start = (*break_time)[0]
	}

	if break_time != nil {
		params.break_end = (*break_time)[1]
	}

	if include_teacher != nil {
		params.include_list = include_teacher
	}

	if exclude_teacher != nil {
		params.exclude_list = exclude_teacher
	}

	// search is applied as we create list
	courses_list := create_courses_list(infile)
	courses := make([]Course, 0)

	for _, course_name := range courses_list {
		body := create_query_body(course_name)
		res_string := ping_classes(_url, body)
		fmt.Println(create_course(res_string, params))
		courses = append(courses, create_course(res_string, params)...)
	}

	schedules := search(courses, params)

	print_schedules(schedules)

	if print_to_outfile {
		print_schedules_to_file(schedules, outfile)
	}

	if draw_schedules {
		for i, schedule := range schedules {
			render(schedule, i+1)
		}
	}

}
