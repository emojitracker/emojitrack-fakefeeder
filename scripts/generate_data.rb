#!/usr/bin/env ruby
require "net/http"
require "json"

RANKINGS_URL = "http://emojitracker.com/api/rankings"

resp = Net::HTTP.get_response(URI.parse(RANKINGS_URL))
abort "Failed to retrieve remote rankings "if resp.code != '200'
results = JSON.parse(resp.body, {:symbolize_names => true})

formatted = results.map do |t|
  %Q[\temojiRanking{char: "#{t[:char]}", id: "#{t[:id]}", name: "#{t[:name]}", score: #{t[:score]}},]
end

puts %[// Automatically generated file -- do not edit directly!
// see `scripts/generate_data.rb` if need to change.
package main

// data obtained from #{RANKINGS_URL} at #{Time.now}
var emojiRankings = []emojiRanking{
#{formatted.join("\n")}
}
]
