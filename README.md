# girelas - Git Releases Assets downloader

Download assets from releases

## usage

```
usage: girelas <rep> [--tag=<s>] [--token=<s>]

positional:
  rep                     repository name, like: owner/repository

options:
      --tag=<s>           tag name, if not specified, download latest release
      --token=<s>         access token required for private repository
  -h, --help              display this help and exit

```

`./girelas owner/packagename` - download assets from latest release

`./girelas owner/packagename --token=... --tag=v1.2.3` - download assets from private repository by specified tag
