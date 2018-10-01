import sys
sys.path.append("packages")
import os
import json

name = "World"
if not os.isatty(sys.stdin.fileno()):
    obj = json.loads(sys.stdin.read())
    if obj["name"] != "":
        name = obj["name"]

print("Hello ", name, "!!!",
      " from python version",
      str(sys.version_info.major) + '.' +
      str(sys.version_info.minor) + '.' +
      str(sys.version_info.micro))
