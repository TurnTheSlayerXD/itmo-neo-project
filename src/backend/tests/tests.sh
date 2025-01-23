 curl -d "$json" http://localhost:8040/add_task

 curl -d "$(cat ./test_list.json)" http://localhost:8040/get_all_tasks

 curl -d "$json" http://localhost:8040/add_solution



 curl -d "$(cat ./test_list.json)" http://localhost:8040/get_all_solutions

