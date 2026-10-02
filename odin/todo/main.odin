package main

import "core:fmt"
import "core:os"
import "core:strconv"
import "core:strings"

todo :: struct {
	id:    int,
	title: string,
	desc:  string,
	done:  bool,
}


FIELD_INDEXES := [?]string{"title", "desc", "done"}


Choice :: enum {
	Add    = 1,
	Delete = 2,
	List   = 3,
}


string_or_bool :: union {
	string,
	bool,
}

main :: proc() {
	todos := make([dynamic]todo, 0, 10)


	buf: [256]byte

	outer_loop: for {
		fmt.println("enter your choice:")
		fmt.println("1 - Add")
		fmt.println("2 - Delete")
		fmt.println("3 - List")

		bytes_read, err := os.read(os.stdin, buf[:])

		if err != nil {
			if bytes_read == 0 {
				break outer_loop
			}
			fmt.println("cannot read input")
			continue
		}

		str_choice := strings.trim_space(string(buf[:bytes_read]))
		if str_choice == "" {
			continue
		}


		int_val, valid := strconv.parse_int(str_choice)
		if !valid {
			fmt.println("invalid choice choose again..")
			continue
		}

		choice := Choice(int_val)


		switch choice {
		case .Add:
			todo := todo {
				id = len(todos) + 1,
			}


			i := 0

			field_loop: for i < 3 {
				buf: [256]byte

				if i == 2 {
					fmt.printf("%s (1 for true/0 for false) :", FIELD_INDEXES[i])
				} else {
					fmt.printf("%s :", FIELD_INDEXES[i])
				}

				bytes_read, err := os.read(os.stdin, buf[:])

				if err != nil || bytes_read == 0 {
					continue outer_loop
				}


				val: string_or_bool

				val = strings.clone(strings.trim_space(string(buf[:bytes_read])))

				if val == "" {
					fmt.println("empty argument")
					continue field_loop
				}

				if i == 2 {
					b, valid := strconv.parse_bool(val.(string))
					if !valid {
						fmt.println("invalid arguement use 1 or 0")
						continue field_loop
					}

					val = b
				}


				if i == 0 {
					todo.title = val.(string)
				} else if i == 1 {
					todo.desc = val.(string)
				} else {
					todo.done = val.(bool)
				}

				i += 1

			}

			append(&todos, todo)


			fmt.println("todo added !!!")

		case .Delete:
			fmt.print("todo id: ")

			buf: [256]byte

			bytes_read, err := os.read(os.stdin, buf[:])
			if err != nil || bytes_read == 0 {
				continue outer_loop
			}

			int_choice, valid := strconv.parse_int(strings.trim_space(string(buf[:bytes_read])))
			if !valid {
				fmt.println("invalid integer")
				continue outer_loop
			}


			for todo, idx in todos {
				if todo.id == int_choice {
					ordered_remove(&todos, idx)
					break
				}
			}

			fmt.println("todo removed...")

		case .List:
			if len(todos) == 0 {
				fmt.println("No todos.")
				continue
			}


			for i := 0; i < len(todos); i += 1 {
				curr_todo := todos[i]

				fmt.println("======= todos ========")
				fmt.println("todo id:", curr_todo.id)
				fmt.println("title", curr_todo.title)
				fmt.println("desc", curr_todo.desc)

				if curr_todo.done {
					fmt.println("status :: done")
				} else {
					fmt.println("status :: not done")
				}

				fmt.println("\n")
			}

		case:
			fmt.println("invalid choice choose again..")

		}

	}


}
