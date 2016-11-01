import sys
sys.path.append("packages")
import fcntl
import os
import json

def non_block_read(output):
    fd = output.fileno()
    fl = fcntl.fcntl(fd, fcntl.F_GETFL)
    fcntl.fcntl(fd, fcntl.F_SETFL, fl | os.O_NONBLOCK)
    try:
        return output.read()
    except:
        return ""

try:
	content = non_block_read(sys.stdin)
	obj = json.loads(content)
	if obj["name"] != "":
		name = obj["name"]
except ValueError:
	name = "World"

print "Hello", name, "!!!"