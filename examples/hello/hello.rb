require 'json'

name = "World"

payload = STDIN.read
if payload != ""
    puts "Received from STDIN: #{payload}"
end

# or using env vars: ENV['PAYLOAD']
payload = ENV['PAYLOAD']

puts 'ARGF: ' + payload.inspect
if payload != ""
    payload = JSON.parse(payload)
    name = payload['name']
end

puts "Hello #{name}!"
