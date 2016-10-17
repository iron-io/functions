require 'open-uri'
require 'slack_webhooks'

payload = STDIN.read

images = JSON.load(open('https://raw.githubusercontent.com/treeder/slackbots/master/guppy/commands.json'))
# Use this one to load from file: responses = JSON.load(File.open('commands.json'))

attachment = {
    "fallback" => "wat?!",
    "text" => "",
    "image_url" => "http://i.imgur.com/7kZ562z.jpg"
}

help = "Available options are:\n"
images.each_key { |k| help << "* #{k}\n" }
# puts help
# sh.set_usage(help, attachment: attachment)
# exit if sh.help?
response = {}
a = []
response[:attachments] = a

if payload.nil? || payload.strip == ""
  response[:text] = help
  a << attachment
  puts response.to_json
  exit
end

# should move the parsing in slack webhooks to a separate method
sh = SlackWebhooks::Hook.new('guppy', payload, "")

r = images[sh.text]
if r
  a << {image_url: r['image_url'], text: ""} # text seems to be required, try here: https://api.slack.com/docs/formatting/builder?msg=%7B%22response_type%22%3A%22in_channel%22%2C%22attachments%22%3A%5B%7B%22text%22%3A%22%22%2C%22image_url%22%3A%22http%3A%2F%2Fi.giphy.com%2FFwpecpDvcu7vO.gif%22%7D%5D%7D
  response[:response_type] = "in_channel"
  response[:text] = "guppy #{sh.text}" 
else
  response[:text] = help
  a << attachment
end

puts response.to_json
