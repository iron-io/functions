## Image processing with Ruby Example, Upload to S3

This example will show you how to process images with a Ruby function, and upload to a S3 bucket.

```sh
# create your func.yaml file
fn init <YOUR_DOCKERHUB_USERNAME>/image-processing

# build the function
fn build

# test it
echo "http://www.sourcecertain.com/img/Example.png" | fn run

# push it to Docker Hub
fn push

# Create a route to this function on IronFunctions (assuming you have an app called `test`)
fn routes create test /image-processing

# you can now access via curl as well
curl -v -X POST http://localhost:8080/r/test/image-processing -d "https://www.nationalgeographic.com/content/dam/science/photos/000/010/1086.ngsversion.1491440409220.adapt.1900.1.jpg"
> https://iron-functions-image-resize.s3.amazonaws.com/1086.ngsversion.1491440409220.adapt.1900.1.jpg
```

