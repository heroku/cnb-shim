require 'rubygems'
require 'bundler'
require 'open-uri'
require 'net/http'
require 'shellwords'
require 'securerandom'

Bundler.require

configure { set :server, :puma }
configure { set :port, ENV['PORT'] || 5000 }

get '/v1/:namespace/:name' do
  id = "#{params['namespace']}/#{params['name']}"
  version = params["version"] || "0.1"
  name = params["name"] || id
  api = params["api"] || "0.2"
  stacks = (params["stacks"] || params["stack"] || "heroku-18,heroku-20").split(",")
  shim_dir = Dir.pwd
  url = "https://buildpack-registry.s3.amazonaws.com/buildpacks/#{id}.tgz"
  shimmed_buildpack = "#{SecureRandom.uuid}.tgz"

  Dir.mktmpdir do |dir|
    Dir.chdir(dir) do
      # setup the shim
      puts "at=shim file=#{shimmed_buildpack}"
      Dir.mkdir("bin")
      FileUtils.cp(File.join(shim_dir, "bin", "build"), "bin")
      FileUtils.cp(File.join(shim_dir, "bin", "detect"), "bin")
      FileUtils.cp(File.join(shim_dir, "bin", "release"), "bin")
      FileUtils.cp(File.join(shim_dir, "bin", "exports"), "bin")

      # write a buildpack.toml
      puts "at=descriptor file=#{shimmed_buildpack} api=#{api} id=#{id} version=#{version} name='#{name}' stacks=#{stacks.join(",")}"
      File.open("buildpack.toml", 'w') do |file|
        file.write( <<TOML )
api = "#{api}"

[buildpack]
id = "#{id}"
version = "#{version}"
name = "#{name}"
TOML

        stacks.each do |stack|
          file.write( <<TOML )

[[stacks]]
id = "#{stack}"
TOML
        end
      end

      # download and extract the target buildpack
      target_dir="target"
      Dir.mkdir(File.join(dir, target_dir))
      puts "at=download file=#{shimmed_buildpack} url='#{url}'"
      `curl --retry 3 --silent --location "#{Shellwords.escape url}" | tar xzm -C #{target_dir}`
    end

    # create a tarball of the tmpdir
    begin
      `tar cz --file=#{shimmed_buildpack} --directory=#{dir} .`
      puts "at=send file=#{shimmed_buildpack} size=#{File.size(shimmed_buildpack)}"
      send_file shimmed_buildpack, :type => "application/x-gzip"
      puts "at=success file=#{shimmed_buildpack}"
    ensure
      puts "at=cleanup file=#{shimmed_buildpack}"
      # this breaks the send_file method, so we're just not going to clean up
      # FileUtils.rm_rf(shimmed_buildpack)
    end
  end
end
