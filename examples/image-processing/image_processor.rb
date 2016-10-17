require 'open-uri'
require 'aws-sdk'
require 'json'
#require 'subexec'
require 'mini_magick'

@output_path = "tmp_images/"

def resize(image, h)
  original_width, original_height = image[:width], image[:height]
  h['width'] ||= original_width
  h['height'] ||= original_height
  image.resize "#{h['width']}x#{h['height']}"
  image
end

def thumbnail(image, h)
  image.combine_options do |c|
    c.thumbnail "#{h['width']}x#{h['height']}"
    c.background 'white'
    c.extent "#{h['width']}x#{h['height']}"
    c.gravity "center"
  end
  image
end

def sketch(image, h)
  image.combine_options do |c|
    c.edge "1"
    c.negate
    c.normalize
    c.colorspace "Gray"
    c.blur "0x.5"
  end
  image
end

def normalize(image, h)
  image.normalize
  image
end

def charcoal(image, h)
  image.charcoal '1'
  image
end

def level(image, h)
  image.level " #{h['black_point']},#{h['white_point']},#{h['gamma']}"
  image
end

def img_path(filename)
  "#{@output_path}#{filename}"
end

def write_image(image, filename)
  image.write img_path(filename)
  # Just in case this is on Docker, gotta change permissions
  `chmod a+rw #{img_path(filename)}`
end

def tile(h)
  file_list=[]
  image = MiniMagick::Image.open(filename)
  original_width, original_height = image[:width], image[:height]
  slice_height = original_height / h['num_tiles_height']
  slice_width = original_width / h['num_tiles_width']
  h['num_tiles_width'].times do |slice_w|
    file_list[slice_w]=[]
    h['num_tiles_height'].times do |slice_h|
      output_filename = "filename_#{slice_h}_#{slice_w}.jpg"
      image = MiniMagick::Image.open(filename)
      image.crop "#{slice_width}x#{slice_height}+#{slice_w*slice_width}+#{slice_h*slice_height}"
      write_image(image, output_filename)
      file_list[slice_w][slice_h] = output_filename
    end
  end
  file_list
end

def merge_images(col_num, row_num, file_list)
  output_filename = "merged_file.jpg"
  ilg = Magick::ImageList.new
  col_num.times do |col|
    il = Magick::ImageList.new
    row_num.times do |row|
      il.push(Magick::Image.read(file_list[col][row]).first)
    end
    ilg.push(il.append(true))
    ilg.append(false).write(output_filename)
  end
  output_filename
end

def upload_file(payload, filename)
  unless payload['disable_network']
    files = [filename].flatten
    files.each do |filepath|
      puts "Uploading the file to s3..."

      s3 = Aws::S3::Client.new(
        region: payload['aws']['region'],
        credentials: Aws::Credentials.new(payload['aws']['access_key'], payload['aws']['secret_key'])
      )

      response = s3.put_object(
        :bucket => payload['aws']['s3_bucket_name'],
        :key => filepath,
        :body => IO.read(img_path(filepath))
      ) 

      if response.successful? == true
        puts "Uploading successful."
        link = "https://s3-"+ payload['aws']['region'] +".amazonaws.com/"+payload['aws']['s3_bucket_name']+"/"+filepath 
        puts "\nYou can view the file here on s3: ", link
      else
        puts "Error uploading to s3."
      end
      puts "-"*60
    end
  end
end

def filename(payload)
  File.basename(payload['image_url'])
end

def download_image(payload)
  fname = filename(payload)
  puts "Downloading file: #{fname}"
  unless payload['disable_network']
    File.open(fname, 'wb') do |fout|
      open(payload['image_url']) do |fin|
        IO.copy_stream(fin, fout)
      end
    end
  end
  fname
end

FileUtils.mkdir_p(@output_path)
`chmod a+rw #{@output_path}`

puts "function started"

payload = STDIN.read
if payload != ""
    payload = JSON.parse(payload)
end

# p payload
file = download_image(payload)
payload['operations'].each do |op|
  puts "\nPerforming #{op['op']} with #{op.inspect}"
  output_filename = op['destination']
  image = MiniMagick::Image.open(file)
  image = self.send(op['op'], image, {}.merge(op))
  image.format op['format'] if op['format']
  write_image(image, output_filename)
  upload_file payload, output_filename
end
puts "function end"
