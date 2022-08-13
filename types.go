package main

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

	// algorithm-specific stuff
	Discovered bool
	Used       bool
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
