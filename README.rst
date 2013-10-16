go-eval
=======

This is a fork of https://github.com/sbinet/go-eval

This fork is used by https://github.com/kyleburton/go-abtab to provide expression support in the @abgrep@ utility.  The changes in the fork focus on allowing variables to be injected into the evaluation runtime.

Installation::

  go get github.com/kyleburton/go-eval/pkg/eval
  go get github.com/kyleburton/go-eval/cmd/go-eval

Usage::

  $ go-eval
  > hello := "world"
  > println(hello)
  world
  >

Documentation:

  http://gopkgdoc.appspot.com/pkg/github.com/sbinet/go-eval/pkg/eval
