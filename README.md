# Cloud Native Buildpack Shim

This is a Cloud Native Buildpack that acts as a shim for [Heroku Buildpacks](https://devcenter.heroku.com/articles/buildpacks).

To use it, install the target buildpack:

```sh-session
$ bin/install "path/to/buildpack.toml" "https://example.com/buildpack.tgz"
```

Then run this buildpack.