import json

import subprocess

def main(): 
    with open("./solutionCheck/tests.json") as f:
        tests =  json.load(f)
    res = subprocess.run(["g++", "./solutionCheck/main.cpp", "-o", "./solutionCheck/main.exe"])
    if res.stderr:
        print(f"Compilation failed: {res.stderr}")
        return -1
    for test in tests:
        inp = test['input']
        out = test['output']
        res = subprocess.run(f'echo "{inp}" | ./solutionCheck/main.exe', shell=True, stdout=subprocess.PIPE).stdout.decode()
        if res != out:
            print(f"Test did not pass\nReceived: {res}\nExpected: {out}")
            return -1
        
    print("success")
    
    return 0


if __name__ == "__main__":

    main()
