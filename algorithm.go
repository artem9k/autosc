package main

import (
	"fmt"
	"math"
	"reflect"
	"sort"
)

const topk int = 10

func get_med_time(a, b int) int {
	return (a + b) / 2
}

func get_mid_day_score(a int) int {
	return int(math.Abs(float64(1200 - a)))
}

func check_constraint_overlap(c1, c2 Constraint) bool {

	// special case where they start at the same time
	var eq = c2.start_t == c1.start_t && c2.end_t == c1.end_t

	var start_overlap_1 = (c2.start_t < c1.start_t && c1.start_t < c2.end_t)
	var start_overlap_2 = (c2.start_t < c2.start_t && c2.start_t < c2.end_t)
	var end_overlap_1 = (c2.start_t < c1.end_t && c1.end_t < c2.end_t)
	var end_overlap_2 = (c1.start_t < c2.end_t && c2.end_t < c1.end_t)

	return !(end_overlap_1 || end_overlap_2 || start_overlap_1 || start_overlap_2 || eq)
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

func check_solutions(solutions [][]Class, new_solution []Class) bool {
	fmt.Println(len(new_solution))
	for _, solution := range solutions {
		if len(solution) == len(new_solution) {

			sort.SliceStable(solution, func(i, j int) bool {
				a := solution[i]
				b := solution[j]
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

			sort.SliceStable(new_solution, func(i, j int) bool {
				a := new_solution[i]
				b := new_solution[j]
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

			if reflect.DeepEqual(solution, new_solution) {
				return false
			}

		}
	}
	return true
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

func DFS_recursive(options []Course, solution []Class, solutions_so_far [][]Class, i int) []Class {
	// search complete
	if i == len(options) {
		check1 := check_class_against_solutions(solution, solutions_so_far)
		if check1 {
			return solution
		} else {
			return nil
		}
	}
	for j := 0; j < len(options[i].Classes); j++ {
		// compare the class with our constr
		check2 := check_class_against_list(solution, options[i].Classes[j])
		if check2 && !options[i].Classes[j].Used {
			next := DFS_recursive(options, append(solution, options[i].Classes[j]), solutions_so_far, i+1)
			if next != nil {
				return next
			}
		}
	}
	return nil
}

func search(options []Course, params Globals) [][]Class {
	solution := make([][]Class, topk)
	for i := 0; i < topk; i++ {
		solution[i] = make([]Class, 0)
		solution[i] = DFS_recursive(options, solution[i], solution, 0)
	}
	return solution
}
