## Image processing with Ruby Example

This example will show you how to process images with a Ruby function.

```sh
# create your func.yaml file
fn init <YOUR_DOCKERHUB_USERNAME>/image-processing

# build the function
fn build

# test it
echo "http://www.sourcecertain.com/img/Example.png" | fn run > image.png

# push it to Docker Hub
fn push

# Create a route to this function on IronFunctions (assuming you have an app called `test`)
fn routes create test /image-processing

# you can now access via curl as well
curl -v -X POST http://localhost:8080/r/test/image-processing -d "http://www.sourcecertain.com/img/Example.png" > image.png
```

