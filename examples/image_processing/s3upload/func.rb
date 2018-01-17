require 'mini_magick'
require 'aws-sdk-s3'

file_uri = STDIN.read.strip
image = MiniMagick::Image.open(file_uri)
image.contrast
image.resize "250x200"
image.rotate "-90"

s3 = Aws::S3::Client.new
bucket = ENV['AWS_S3_BUCKET']

obj = s3.put_object( bucket: bucket,
                     key: File.basename(file_uri),
                     body: image.tempfile,
                     acl: "public-read",
                     cache_control: "max-age=604800")

# Unfortunately put_object returns `put object output`, not an object.
# So we create another reference here. Probably there is a better way to do this in S3 API.
obj = Aws::S3::Object.new bucket_name: bucket, key: File.basename(file_uri)
puts obj.public_url
