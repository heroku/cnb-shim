# Cloud Native Buildpack Shim

[![CI](https://github.com/heroku/cnb-shim/actions/workflows/ci.yml/badge.svg)](https://github.com/heroku/cnb-shim/actions/workflows/ci.yml)

> [!WARNING]
> This project is not actively maintained and does not support modern Buildpack API and lifecycle versions.
>
> Please switch to native CNB implementations rather than using this shim.
>
> See [Heroku's natively supported CNB languages](https://github.com/heroku/buildpacks#supported-languages) or [search for community buildpacks](https://registry.buildpacks.io/).

This is a Cloud Native Buildpack that acts as a shim for classic [Heroku Buildpacks](https://devcenter.heroku.com/articles/buildpacks).

## Usage

This shim can be used with any buildpack in the [Heroku Buildpack Registry](https://devcenter.heroku.com/articles/buildpack-registry) by specifying a URL in the form:

```
https://cnb-shim.herokuapp.com/v1/<namespace>/<name>
```

### Example: Elixir

```
$ pack build elixir-app --buildpack https://cnb-shim.herokuapp.com/v1/hashnuke/elixir --builder heroku/buildpacks:18
```

For a complete list of available buildpacks run the following command from the [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli):

```
$ heroku buildpacks:search
```

## Applying the Shim Manually

To use the shim manually, install the target buildpack:

```sh-session
$ sbin/install "path/to/buildpack.toml" "https://example.com/buildpack.tgz"
```

Then run this buildpack.

### Example: Elixir

To use this shim with the [hashnuke/elixir](https://github.com/HashNuke/heroku-buildpack-elixir) buildpack, install [`pack` CLI](https://github.com/buildpack/pack) and run:

```
$ cd elixir-cnb

$ curl -L https://github.com/heroku/cnb-shim/releases/download/v0.1/cnb-shim-v0.1.tgz | tar xz

$ cat > buildpack.toml << TOML
api = "0.2"

[buildpack]
id = "hashnuke.elixir"
version = "0.1"
name = "Elixir"

[[stacks]]
id = "heroku-22"
TOML

$ sbin/install buildpack.toml https://buildpack-registry.s3.amazonaws.com/buildpacks/hashnuke/elixir.tgz

$ cd ~/my-elixir-app/

$ pack build elixir-app --builder heroku/buildpacks --buildpack ~/path/to/elixir-cnb
```

## License

MIT
