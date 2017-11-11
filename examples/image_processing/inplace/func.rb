require 'mini_magick'

payload = STDIN.read.strip
image = MiniMagick::Image.open(payload)
image.contrast
image.resize "250x200"
image.rotate "-90"
image.write STDOUT
