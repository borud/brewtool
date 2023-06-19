# brewtool

This is an experimental tool to make the `span.rb` file used in
<https://github.com/lab5e/homebrew-tools>. I might generalize this
a bit when i get around to it.

## Building

```shell
go build
```

## Example usage

```shell
bin/brewtool --owner lab5e --repo spancli gen --bin span --name foo --desc "Span command line client"
```
