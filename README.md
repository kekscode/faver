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

Known issues:

Only one way to detect favicon links is supported right now which does not work with every website.
