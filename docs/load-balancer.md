# IronFunctions Load Balancer

IronFunctions Load Balancer (`lb`) is a functions aware load balancer that can be configured either to improve availability or improve performance.

Its main feature is that it aware of function calls, thus it can handle load balance more intelligently.

Features:
- API for operation of Function nodes
- Automatic removal of dead Function nodes
- Persistent storage of load balance learnings (used across restarts)

## CLI operation

```sh
# ./fnlb serve &> fnlb.log
# ./fnlb node add x.x.x.x:pppp
# ./fnlb node add y.y.y.y:pppp
# ./fnlb node list
x.x.x.x:pppp - online
y.y.y.y:pppp - connecting
# ./fnlb monitor
[loads a screenful of information]
```

### Commands

* `serve` starts the load balancer in IronFunction default port (`0.0.0.0:8080`). It produces all its logs directly to stdout/stderr. If it finds `routes.json` in the local directory, it will reload all its past learnings. Remove the file a fresh restart. In case no node is available, it shall reject every request with `404 Not Found`.

* `node` is a command group that manipulates the operation of nodes for `fnlb`.

	* `add` and `del` operate the nodes. Each expects a full hostname or IP, e.g., `node.function.com:8080` or `192.168.0.1:8080`.

	* `list` show all nodes and their last known state. The possible states are: `online`, `connecting` (as soon as it is first added into the `fnlb`), `unreachable` and `removed` (in this case the node is removed, but hasn't been yet flushed out of the redirection table).

* `monitor` connects to the load balancer and gather realtime information about its internals.

## HTTP API

The load balancer reserves the `/v1/lb/...` path for its internal operations.

### `/v1/lb/nodes/{add,del}` - shows, adds and deletes nodes from the load balancer.

Input
```
"x.x.x.x:pppp"
```

Output:
* `202 Accepted` for all successful operations. In case of duplicated additions, it succeeds silently.

* `404 Not Found` in case of non-existent deletion.

* `200 OK` in case of showing nodes
```json
["x.x.x.x:pppp","y.y.y.y:pppp"]
```

### `/v1/lb/status` - status snapshot

No input necessary

Output: TBD

