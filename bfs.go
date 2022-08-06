/**
automtic scheduler for cu boulder classes (or honestly any other school)
also FUCK GO for making me use all variables
**/

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
)

type RequestOther struct {
	Srcdb string
}

type RequestCriteria struct {
	Field string
	Value string
}

type RequestPayload struct {
	Other    RequestOther
	Criteria []RequestCriteria
}

type Classes struct {
	Class string
}

func print_fatal(err string) {
	fmt.Println("Fatal error: " + err)
}

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
	Instructor  string
	CreditHours int
	//Type        string
	//Constr      []Constraint
}

func (c *Class) print() {
	fmt.Printf("class: %v")
}

func rel(c1, c2 Constraint) {
	// relate two constraints.
}

func search() {
	// search for constraint overlaps over days first, then times...

}

func create_class(data map[string]interface{}) Class {
	var cls Class

	var instr = data["instr"].(string)
	var code = data["code"].(string)

	var cart_opts = data["cart_opts"].(interface{})
	var credit_hrs = cart_opts["credit_hrs"].(map[string]interface{})
	var first_option = credit_hrs["options"].([]map[string]interface{})[0]
	var credit_hours = first_option["label"].(int)

	cls.Code = code
	cls.Instructor = instr
	cls.CreditHours = credit_hours

	return cls
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

	var js map[string]interface{}
	json.Unmarshal(res_body, &js)

	// something went wrong
	if val, ok := js["fatal"]; ok {
		print_fatal(val.(string))
	}

	s := reflect.ValueOf(js["results"]).Index(0).Interface()
	i := s.(map[string]interface{})
	fmt.Println(i["code"])

	var new_class = create_class(i)
	fmt.Println(new_class)

}
