(* this is a simple 
   (* this is a nested comment *)

   i can write here however much i want this is crazy commenting 
   in other languages that i have tried so far, If we just break the line comment gets broken
   specially if it's a single line comment
 comment *)

type todo_status = Active | Done | InvalidStatus

let parse_todo_status (str : string) : todo_status =
  match str with
  | "active" -> Active
  | "done" -> Done
  | _ -> InvalidStatus

type todo_data = {
  id : int;
  mutable title : string;
  mutable desc : string;
  status : todo_status;
}

type update_todo_data = {
  mutable title : string option;
  mutable desc : string option;
  status : todo_status option;
}

type todo = Todo of todo_data | UpdateTodo of update_todo_data

let todos = ref []

type user_input =
  | AddTodo of todo_data
  | DeleteTodo of int
  | Update of string option * string option * todo_status option
  | Quit
  | InvalidCmd

let parse_todo_id (str_id : string) : int =
  match str_id |> String.trim |> int_of_string_opt with
  | Some id -> id
  | None -> failwith "[error] invalid id only integer ids are valid"

let parse_todo_id_opt (opt_id : string option) : int option =
  match opt_id with
  | Some id -> Some (parse_todo_id id)
  | None -> None

let parse_valid_non_empty_string (str : string) (field_name : string) =
  match str |> String.trim with
  | "" -> "[error] ${field_name} cannot be empty"
  | valid_str -> valid_str

type update_todo = {
  mutable title : string;
  mutable desc : string;
  status : todo_status;
}

let parse_todo (opt_id : string option) (title : string) (desc : string)
    (status : string) : todo =
  let valid_id = opt_id |> parse_todo_id_opt in
  let valid_title = title |> parse_valid_non_empty_string "title" in
  let valid_desc = desc |> parse_valid_non_empty_string "desc" in
  let valid_status = status |> parse_todo_status in

  match valid_id with
  | Some some_valid_id ->
      Todo
        {
          id = some_valid_id;
          title = valid_title;
          desc = valid_desc;
          status = valid_status;
        }
  | None ->
      UpdateTodo
        { title = valid_title; desc = valid_desc; status = valid_status }

let parse_cmd (cmd : string) : user_input =
  match cmd |> String.trim |> String.split_on_char ' ' with
  | [ "add"; id; title; desc; status ] -> (
      let parsed_todo_data = parse_todo (Some id) title desc status in
      match parsed_todo_data with
      | Todo data -> AddTodo data
      | UpdateTodo _ -> InvalidCmd)
  | [ "delete"; id ] -> DeleteTodo (parse_todo_id id)
  | [ "update"; title; desc; status ] -> (
      let update_todo_data = parse_todo None title desc status in
      match update_todo_data with
      | UpdateTodo data -> UpdateTodo data
      | Todo _ -> InvalidCmd)
  | [ "quit" ] -> Quit
  | _ -> InvalidCmd

let () =
  while true do
    let input = read_line () |> parse_cmd in

    match input with
    | AddTodo -> add_todo ()
  done
