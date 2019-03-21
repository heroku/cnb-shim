# Cloud Native Buildpack Shim [![Build Status](https://travis-ci.com/heroku/cnb-shim.svg?token=bFx8xfjczBrYptbXskcQ&branch=master)](https://travis-ci.com/heroku/cnb-shim)

This is a Cloud Native Buildpack that acts as a shim for [Heroku Buildpacks](https://devcenter.heroku.com/articles/buildpacks).

To use it, install the target buildpack:

```sh-session
$ bin/install "path/to/buildpack.toml" "https://example.com/buildpack.tgz"
```

Then run this buildpack.

## Example: Elixir

To use this shim with the hashnuke/elixir buildpack, install [`pack` CLI](https://github.com/buildpack/pack) and run:

```
$ cd elixir-cnb

$ curl -L https://github.com/heroku/cnb-shim/releases/download/v0.0.2/cnb-shim-v0.0.2.tgz | tar xz

$ cat > buildpack.toml << TOML
> [buildpack]
> id = "hashnuke.elixir"
> version = "0.1"
> name = "Elixir"
>
> [[stacks]]
> id = "heroku-18"
TOML

$ bin/install buildpack.toml https://buildpack-registry.s3.amazonaws.com/buildpacks/hashnuke/elixir.tgz

$ cd ~/my-elixir-app/

$ pack build elixir-app --builder heroku/buildpacks --buildpack ~/path/to/elixir-cnb
```

## License

MIT
