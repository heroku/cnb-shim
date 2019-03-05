# Cloud Native Buildpack Shim [![Build Status](https://travis-ci.com/heroku/cnb-shim.svg?token=bFx8xfjczBrYptbXskcQ&branch=master)](https://travis-ci.com/heroku/cnb-shim)

This is a Cloud Native Buildpack that acts as a shim for [Heroku Buildpacks](https://devcenter.heroku.com/articles/buildpacks).

To use it, install the target buildpack:

```sh-session
$ bin/install "path/to/buildpack.toml" "https://example.com/buildpack.tgz"
```

Then run this buildpack.