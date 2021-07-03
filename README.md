# faver

A download tool for favicons

Install:

    $> go install github.com/kekscode/faver

or

    $> go build .

Usage:

    $> faver [url ...]

Example:

     $> faver https://www.gitlab.com https://www.test.de

Piping with STDIN is also supported:

    $> cat urls.txt | ./faver

Every target in `urls.txt` or STDIN in general must be a single line with `\n`
as seperator.

