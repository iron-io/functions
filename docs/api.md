
## IronFunctions Options

When starting IronFunctions, you can pass in the following configuration variables as environment variables. Use `-e VAR_NAME=VALUE` in 
docker run.  For example:

```
docker run -e VAR_NAME=VALUE ...
```

<table>
<tr>
<th>Env Variables</th>
<th>Description</th>
</tr>
<tr>
<td>DB</td>
<td>The database URL to use in URL format. See Databases below for more information. Default: BoltDB in current working directory `bolt.db`.</td>
</tr>
<tr>
<td>MQ</td>
<td>The message queue to use in URL format. See Message Queues below for more information. Default: BoltDB in current working directory `queue.db`.</td>
</tr>
<tr>
<td>API_URL</td>
<td>The primary functions api URL to pull tasks from (the address is that of another running functions process).</td>
</tr>
<tr>
<td>PORT</td>
<td>Default (8080), sets the port to run on.</td>
</tr>
<tr>
<td>NUM_ASYNC</td>
<td>The number of async runners in the functions process (default 1).</td>
</tr>
<tr>
<td>LOG_LEVEL</td>
<td>Set to `DEBUG` to enable debugging. Default is INFO.</td>
</tr>

</table>

## Databases

We currently support the following databases and they are passed in via the `DB` environment variable. For example:

```sh
docker run -e "DB=postgres://user:pass@localhost:6212/mydb" ...
```

### [Bolt](https://github.com/boltdb/bolt) (default)

URL: `bolt:///functions/data/functions.db`

Bolt is an embedded database which stores to disk. If you want to use this, be sure you don't lose the data directory by mounting
the directory on your host. eg: `docker run -v $PWD/data:/functions/data -e DB=bolt:///functions/data/bolt.db ...`

### [Redis](http://redis.io/)

URL: `redis://localhost:6379/`

Use a Redis instance as your database. Be sure to enable [peristence](http://redis.io/topics/persistence).

### [PostgreSQL](http://www.postgresql.org/)

URL: `postgres://user123:pass456@ec2-117-21-174-214.compute-1.amazonaws.com:6212/db982398`

Use a PostgreSQL database. If you're using IronFunctions in production, you should probably start here.

### What about database X?

We're happy to add more and we love pull requests, so feel free to add one! Copy one of the implementations above as a starting point. 

## Message Queues

A message queue is used to coordinate asynchronous function calls that run through IronFunctions.

We currently support the following message queues and they are passed in via the `MQ` environment variable. For example:

```sh
docker run -e "MQ=redis://localhost:6379/" ...
```

### [Bolt](https://github.com/boltdb/bolt) (default)

URL: `bolt:///titan/data/functions-mq.db`

See Bolt in databases above. The Bolt database is locked at the file level, so
the file cannot be the same as the one used for the Bolt Datastore.

### [Redis](http://redis.io/)

See Redis in databases above.

### What about message queue X?

We're happy to add more and we love pull requests, so feel free to add one! Copy one of the implementations above as a starting point. 

