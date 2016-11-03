# Introduction

This guide will walk you through creating and testing a simple Lambda function.

We need the the `fnctl` tool for the rest of this guide. You can install it
by following [these instructions](https://github.com/iron-io/function/fnctl).

## Creating the function

Let's convert the `node-exec` AWS Lambda example to Docker. This simply
executes the command passed to it as the payload and logs the output.

    var exec = require('child_process').exec;
    
    exports.handler = function(event, context) {
        if (!event.cmd) {
            context.fail('Please specify a command to run as event.cmd');
            return;
        }
        var child = exec(event.cmd, function(error) {
            // Resolve with result of process
            context.done(error, 'Process complete!');
        });
    
        // Log process stdout and stderr
        child.stdout.on('data', console.log);
        child.stderr.on('data', console.error);
    };

Create an empty directory for your project and save this code in a file called
`node_exec.js`.

Now let's use `fnctl`'s Lambda functionality to create a Docker image. We can
then run the Docker image with a payload to execute the Lambda function.

    $ fnctl lambda create-function irontest/node-exec:1 nodejs node_exec.handler node_exec.js
    Image output Step 1 : FROM iron/lambda-nodejs
    ---> 66fb7af42230
    Step 2 : ADD node_exec.js ./node_exec.js
    ---> 6f922128da71
    Removing intermediate container 9644b02e95bc
    Step 3 : CMD node_exec.handler
    ---> Running in 47b2b1f3e779
    ---> 5eef8d2d3111
    Removing intermediate container 47b2b1f3e779
    Successfully built 5eef8d2d3111

As you can see, this is very similar to creating a Lambda function using the
`aws` CLI tool. We name the function as we would name other Docker images. The
`1` indicates the version. You can use any string. This way you can configure
your deployment environment to use different versions. The handler is
the name of the function to run, in the form that nodejs expects
(`module.function`). Where you would package the files into a `.zip` to upload
to Lambda, we just pass the list of files to `fnctl`. If you had node
dependencies you could pass the `node_modules` folder too.

You should now see the generated Docker image.

    $ docker images
    REPOSITORY                                      TAG    IMAGE ID         CREATED             VIRTUAL SIZE
    irontest/node-exec                              1      5eef8d2d3111     9 seconds ago       44.94 MB
    ...

## Testing the function

The `test-function` subcommand can launch the Dockerized function with the
right parameters.

    $ fnctl lambda test-function irontest/node-exec:1 --payload '{ "cmd": "echo Dockerized Lambda" }'
    Dockerized Lambda!

You should see the output. Try changing the command to `date` or something more
useful.