package main

import (
	"fmt"
	"math"
)

const topk int = 10

func get_med_time(a, b int) int {
	return (a + b) / 2
}

func get_mid_day_score(a int) int {
	return int(math.Abs(float64(1200 - a)))
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

func pop_cls(stack []Class) ([]Class, Class) {
	l := len(stack) - 1
	return stack[:l], stack[l]
}

func push_cls(stack []Class, new_item Class) []Class {
	return append(stack, new_item)
}

func pop_int(stack []int) ([]int, int) {
	l := len(stack) - 1
	return stack[:l], stack[l]
}

func push_int(stack []int, new_item int) []int {
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

func alg_init_course_list(courses []Course) {
	for _, crs := range courses {
		for _, cls := range crs.Classes {
			cls.Discovered = false
			cls.Used = false
		}
	}
}

func alg_reset_course_list(courses []Course) {
	for _, crs := range courses {
		for _, cls := range crs.Classes {
			if !cls.Used {
				cls.Discovered = false
			}
		}
	}
}

func DFS_iterative(options []Course, solution [][]Class) [][]Class {

	if options == nil || solution == nil {
		return nil
	}

	alg_init_course_list(options)

	// the class stack is independent from the solution. We copy whatever's going on
	// in the class stack to the solution. This allows us to sample the top-k solutions,
	// instead of just one.

	class_stack := make([]Class, 0)
	level_stack := make([]int, 0)

	var v Class
	var l int
	var prev_l int
	var j int = 0 // solution cursor

	push_cls(class_stack, options[0].Classes[0])
	push_int(level_stack, 0)

	for len(class_stack) != 0 {

		class_stack, v = pop_cls(class_stack)
		level_stack, l = pop_int(level_stack)

		// special case: l is lower than prev_l, pop an element from classes
		if l < prev_l {
			solution[j], _ = pop_cls(solution[j])
		}

		if !v.Used {

			// is v valid?
			check := check_class_against_list(solution[j], v)

			if check {
				// class is good. begin pushing classes from next level
				// for class in neighbors: push class to stack
				push_cls(solution[j], v)

				if j == len(options)-1 {
					// yes
					v.Used = true

					// are we totally done?
					if l == topk-1 {
						return solution
					}

					j += 1
					alg_reset_course_list(options)
				} else {
					// label v as discovered
					v.Discovered = true
					v.Used = true
				}

			} else {
				// dont push anything

				// is v the last class?

			}

		}

	}

	return nil
}

func DFS_recursive(options []Course, solution []Class, i int) []Class {
	// search complete
	if i == len(options) {
		return solution
	}
	for j := 0; j < len(options[i].Classes); j++ {
		// compare the class with our constr
		check := check_class_against_list(solution, options[i].Classes[j])
		if check && !options[i].Classes[j].Used {
			options[i].Classes[j].Used = true
			next := DFS_recursive(options, append(solution, options[i].Classes[j]), i+1)
			if next != nil {
				return next
			} else {
				options[i].Classes[j].Used = false
			}

		}
	}
	return nil
}
func search(options []Course, params Globals) [][]Class {
	solution := make([][]Class, topk)
	for i := 0; i < topk; i++ {
		solution[i] = make([]Class, 0)
		solution[i] = DFS_recursive(options, solution[i], 0)
		if solution[i] == nil {
			fmt.Println("nil", i)
		}
	}

	return solution
}
