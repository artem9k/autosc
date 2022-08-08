/**
automatic scheduler for cu boulder classes (or honestly any other school)
**/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/tidwall/gjson"
)

// GLOBALS

var DAYS_IN_WEEK = 7

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

// FUNCTIONS

func get_safe_atoi(num string) int {
	var val, err = strconv.Atoi(num)
	if err != nil {
		panic(err)
	}
	return val
}

func parse_input_file(filename string) {
	//parse newline-separated input from file

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
	var credit_hours_string = gjson.Get(cart_opts_string, "credit_hrs.options.0.label").Str
	var credit_hours = get_safe_atoi(credit_hours_string)

	var new_class Class
	new_class.Instructor = instr
	new_class.Type = _type
	new_class.CreditHours = credit_hours

	constraints := make([]Constraint, 0)

	meeting_times_str.ForEach(func(key, value gjson.Result) bool {
		fmt.Println(value)
		var meet_day = get_safe_atoi(value.Get("meet_day").Str)
		var start_time = get_safe_atoi(value.Get("start_time").Str)
		var end_time = get_safe_atoi(value.Get("end_time").Str)
		new_constr := Constraint{meet_day, start_time, end_time}
		fmt.Println(new_constr)
		constraints = append(constraints, new_constr)
		return true
	})

	new_class.Constraints = constraints

	fmt.Println(new_class)

	return new_class
}

func create_course(data string) []Course {

	// return a slice with either 1 or 2 courss, depending on whether the course
	// needs a recitation

	// TODO handle edge case where no classes, error
	// TODO handle edge case where recitation + no recitation
	// TODO exclude teacher
	// TODO minimum time between classes
	// TODO latest possible time for a class
	// TODO earliest possible time for a class
	// TODO dead zone - comma separated times where no classes. Implement this as a list

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
	for i := 0; i < DAYS_IN_WEEK; i++ {
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

func DFS_iterative() []Course {
	solution := make([]Course, 0)
	// i is the Course cursor. So, for example, the first Course would have no constraints
	//i := 0
	// j is the Class cursor. So, we iterate over Classes for each i until we find one that doesn't break the constraints
	//j := 0
	for {

		return nil
	}

	return solution
}

func main() {

	var infile string
	var outfile string
	var topk int
	var class = "PHYS 1110"
	var _url = "https://classes.colorado.edu/api/?page=fose&route=search&alias=PHYS%201110"

	_url = _url + "?page=fose&route=search&"
	_url = _url + "alias=" + url.QueryEscape(class)

	flag.IntVar(&topk, "topk", 10, "list of possible schedules to return, ranked by fit. Default is 10")
	flag.StringVar(&infile, "infile", "input.txt", "input file with each class on a new line.")
	flag.StringVar(&outfile, "outfile", "output.txt", "output file with a list of new schedules")
	flag.Parse()

	var body string = "{\"other\":{\"srcdb\":\"2227\"},\"criteria\":[{\"field\":\"alias\",\"value\":\"PHYS 1110\"}]}"

	res, _ := http.Post(_url, "application/json", bytes.NewBufferString(body))
	defer res.Body.Close()
	res_body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var res_string = string(res_body)

	courses := make([]Course, 0)

	courses = append(courses, create_course(res_string)...)
	for _, i := range courses {
		fmt.Println(i.Code, i.Type)
		for _, j := range i.Classes {
			fmt.Println(j)
		}
	}
}
