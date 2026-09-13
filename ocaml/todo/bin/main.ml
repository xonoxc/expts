(* this is a simple 
   (* this is a nested comment *)

   i can write here however much i want this is crazy commenting 
   in other languages that i have tried so far, If we just break the line comment gets broken
   specially if it's a single line comment
 comment *)

type todo_status =
  | Active
  | Done
  | InvalidStatus

let parse_todo_status (str : string) : todo_status =
  match str |> String.trim with
  | "active" -> Active
  | "done" -> Done
  | _ -> InvalidStatus

let parse_todo_status_opt (str : string) : todo_status option =
  match str |> String.trim with
  | "active" -> Some Active
  | "done" -> Some Done
  | _ -> None

type todo_data = {
  id : int;
  mutable title : string;
  mutable desc : string;
  mutable status : todo_status;
}

type update_todo_data = {
  id : int;
  mutable title : string option;
  mutable desc : string option;
  mutable status : todo_status option;
}

type todo =
  | Todo of todo_data
  | UpdateTodo of update_todo_data

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

let parse_valid_non_empty_string_opt (str : string) : string option =
  match str |> String.trim with
  | "" -> None
  | valid_str -> Some valid_str

let todos : todo_data list ref = ref []

let parse_todo
    (opt_id : string option) (title : string) (desc : string) (status : string)
    : todo =
  let valid_id = opt_id |> parse_todo_id_opt in

  match valid_id with
  | Some some_valid_id ->
      UpdateTodo
        {
          id = some_valid_id;
          title = title |> parse_valid_non_empty_string_opt;
          desc = desc |> parse_valid_non_empty_string_opt;
          status = status |> parse_todo_status_opt;
        }
  | None ->
      let new_id = !todos |> List.length in
      Todo
        ({
           id = new_id;
           title = parse_valid_non_empty_string title "title";
           desc = parse_valid_non_empty_string desc "desc";
           status = status |> parse_todo_status;
         }
          : todo_data)

type user_input =
  | AddTodo of todo_data
  | DeleteTodo of int
  | Update of int * string option * string option * todo_status option
  | ShowTodos
  | Quit
  | InvalidCmd

let status_to_string (status : todo_status) : string =
  match status with
  | Active -> "active"
  | Done -> "done"
  | InvalidStatus -> "?"

let print_todos (todo_list : todo_data list) : unit =
  match todo_list with
  | [] -> print_endline "No todos yet."
  | _ ->
      print_endline "Todos:";
      List.iter
        (fun (todo : todo_data) ->
          Printf.printf "  [%d] %s - %s (%s)\n" todo.id todo.title todo.desc
            (status_to_string todo.status))
        todo_list

let parse_cmd (cmd : string) : user_input =
  match cmd |> String.trim |> String.split_on_char ' ' with
  (* add todo command  *)
  | [ "add"; title; desc; status ] -> (
      let parsed_todo_data = parse_todo None title desc status in
      match parsed_todo_data with
      | Todo data -> AddTodo data
      | UpdateTodo _ -> InvalidCmd
      (* delete todo by Id *))
  | [ "delete"; id ] -> DeleteTodo (parse_todo_id id) (* update todo by Id *)
  | "update" :: id :: rest ->
      let title =
        match List.nth_opt rest 0 with
        | Some title -> parse_valid_non_empty_string_opt title
        | None -> None
      in
      let desc =
        match List.nth_opt rest 1 with
        | Some desc -> parse_valid_non_empty_string_opt desc
        | None -> None
      in
      let status =
        match List.nth_opt rest 2 with
        | Some status -> parse_todo_status_opt status
        | None -> None
      in
      Update (parse_todo_id id, title, desc, status)
  | [ "quit" ] -> Quit (* all other command are invalid *)
  | [ "list" ]
  | [ "show" ]
  | [ "ls" ] ->
      ShowTodos
  | _ -> InvalidCmd

let () =
  try
    while true do
      print_endline "Enter a command (add, delete, update, list, quit):";
      print_string "> ";

      let input = read_line () |> parse_cmd in

      match input with
      | AddTodo data -> todos := !todos @ [ data ]
      | DeleteTodo id ->
          todos := List.filter (fun (todo : todo_data) -> todo.id <> id) !todos
      | Update (todo_id, todo_title, todo_desc, todo_status) -> (
          match
            List.find_opt (fun (todo : todo_data) -> todo.id = todo_id) !todos
          with
          | Some todo -> (
              (match todo_title with
              | Some title -> todo.title <- title
              | None -> ());

              (match todo_desc with
              | Some desc -> todo.desc <- desc
              | None -> ());

              match todo_status with
              | Some status -> todo.status <- status
              | None -> ())
          | None -> failwith "invalid todo id not found")
      | ShowTodos -> print_todos !todos
      | Quit -> raise Exit
      | InvalidCmd ->
          print_newline ();
          print_string "Invalid command \n"
    done
  with
  | Exit -> print_endline "Exiting ...."
