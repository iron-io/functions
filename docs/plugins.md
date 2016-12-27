# Plugins

Since Go1.8, IronFunctions also supports plugins. You can extend it with binary
plugins that builds on internal bindings covered in [extending.md](operating/extending.md).

In order to use a plugin, make sure the `.so` file is compiled and placed
somewhere accessible by IronFunctions daemon. When starting it, you will need to
set up an environment variable stating where the plugin are:

```sh
PLUGINS="commonlog.so" ./functions
```

Note that you can activate more than one plugin at a time using comma:

```sh
PLUGINS="commonlog.so,requestdump.so" ./functions
```

When started this way, IronFunctions will try to plug these extensions, which
you can actually check their success or failure in the logs:

```
INFO[0000] plugging in                                   plugin=commonlog.so
ERRO[0000] Could not fetch task                          error="Get http://127.0.0.1:8080/tasks: dial tcp 127.0.0.1:8080: getsockopt: connection refused" runner=async
INFO[0000] plugged                                       plugin=commonlog.so type=RunnerListener
INFO[0000] plugging in                                   plugin=requestdump.so
INFO[0000] plugged                                       plugin=requestdump.so type=RunnerListener
```

In order to compile a plugin, you must use the Go's correct build mode:

```sh
cd plugins/requestdump/
go build -v -buildmode=plugin
mv requestdump.so ../../
PLUGINS="requestdump.so" ./functions
```

You can have more details about this build mode in [Go's documentation](https://tip.golang.org/pkg/plugin/).