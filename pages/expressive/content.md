This week there was a discussion on the golang-nuts mailing list about an idiomatic way to update a slice of structs. For example, consider this struct representing a set of counters.

	type E struct {
		A, B, C, D int
	}

	var e = make([]E, 1000)

Updating these counters may take the form

	for i := range e {
		e[i].A += 1
		e[i].B += 2
		e[i].C += 3
		e[i].D += 4
	}

Which is good idiomatic Go code. It's pretty fast too

	BenchmarkManual   500000              4642 ns/op

However there is a problem with this example. Each access the `i`th element of `e` requires the compiler to insert an array bounds checks. You can avoid 3 of these checks by referencing the `i`th element once per iteration.

	for i := range e {
		v := &e[i]
		v.A += 1
		v.B += 2
		v.C += 3
		v.D += 4
	}

By reducing the number of subscript checks, the code now runs considerably faster.

	BenchmarkUnroll  1000000              2824 ns/op

If you are coding a tight loop, clearly this is the more efficient method, but it comes with a cost to readability, as well as a few gotchas. Someone else reading the code might be tempted to move the creation of `v` into the for declaration, or wonder why the address of `e[i]` is being taken. Both of these changes would cause an incorrect result. Obviously tests are there to catch this sort of thing, but I propose there is a better way to write this code, one that doesn't sacrifice performance, and expresses the intent of the author more clearly.

	func (e *E) update(a, b, c, d int) {
		e.A += a
		e.B += b
		e.C += c
		e.D += d
	}

	for i := range e {
		e[i].update(1, 2, 3, 4)
	}

Because `E` is a named type, we can create an `update()` method on it. Because `update` is declared on with a receiver of `*E`, the compiler automatically inserts the `(&e[i]).update()` for us. Most importantly because of the simplicity of the `update()` method itself, the inliner can roll it up into the body of the calling loop, negating the method call cost. The result is very similar to the hand unrolled version.

	BenchmarkUpdate   500000              2996 ns/op

In conclusion, as Cliff Click observed, there are [Lies, Damn Lies and Microbenchmarks](http://www.azulsystems.com/events/javaone_2002/microbenchmarks.pdf). This post is an example of a least one of the three. My intent in writing was not to spark a language war about who can update a struct the fastest, but instead argue that Go lets you write your code in a more expressive manner without having to trade off performance.

You can find the source code for the benchmarks presented in this post on Github, https://gist.github.com/1796018
