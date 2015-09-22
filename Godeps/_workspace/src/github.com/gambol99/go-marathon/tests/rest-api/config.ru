$:.unshift File.join(File.dirname(__FILE__),'.')

require 'rack'
require 'sinatra'
require 'rest-api'

run RestAPI::Application
