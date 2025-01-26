 curl -d "$(cat ./add_task.json)" http://localhost:8040/add_task

 curl -d "$(cat ./test_list.json)" http://localhost:8040/get_all_tasks

 curl -d "$(cat ./add_solution.json)" http://localhost:8040/add_solution



 curl -d "$(cat ./test_list.json)" http://localhost:8040/get_all_solutions


 curl -d "$(cat ./test_list.json)" http://localhost:8040/get_balance

 curl -d "$(cat ./test_list.json)" http://localhost:8040/get_owned_tasks
 curl -d "$(cat ./test_list.json)" http://localhost:8040/get_owned_solutions



 curl -d "$(cat ./createaccount.json)" http://localhost:8040/createaccount
