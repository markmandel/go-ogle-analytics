## Go-ogle Analytics

Track and monitor your Go programs for free with Google Analytics

The `ga` package is essentially a Go wrapper around the [Google Analytics - Measurement Protocol](https://developers.google.com/analytics/devguides/collection/protocol/v1/reference)

**Warning** :warning:

1. This package is 95% generated from [Google Analytics - Measurement Protocol](https://developers.google.com/analytics/devguides/collection/protocol/v1/reference) so it may contain bugs - please report them.
1. Although the protocol provides types, currently they're not being used
	* GA uses the presence of a property to determine whether or not it's set. So with boolean values for example,
	there's 3 states, `true`, `false` and *unset*. This makes it difficult when coming modelling this in Go land since `bool`s
	must be either `true` or `false` - there is no *unset*.
1. This package is beta software and may change due to **.1** and/or **.2**.
1. GA allows "10 million hits per month per property" and will reject requests after that

### Install

```
go get -v github.com/jpillora/go-ogle-analytics
```

### Usage

1. Log into GA and create a new property and note its Tracker ID

1. Create a `ga-test.go` file

	``` go
	package main

	import "github.com/jpillora/go-ogle-analytics"

	func main() {
		client, err := ga.NewClient("UA-XXXXXX-Y") //Tracker ID
		if err != nil {
			panic(err)
		}

		err = client.Send(&ga.Event{
			Category: "Foo",
			Action:   "Bar",
			Label:    "Bazz",
		})

		if err != nil {
			panic(err)
		}

		println("Event fired!")
	}
	```

1. In GA, go to Real-time > Events

1. Run `ga-test.go`

	```
	$ go run ga-test.go
	Event fired!
	```

1. Watch as your event appears

	![foo-ga](https://cloud.githubusercontent.com/assets/633843/5979585/023fc580-a8fd-11e4-803a-956610bcc2e2.png)

### Documentation

#### http://godoc.org/github.com/jpillora/go-ogle-analytics

#### MIT License

Copyright © 2015 &lt;dev@jpillora.com&gt;

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.