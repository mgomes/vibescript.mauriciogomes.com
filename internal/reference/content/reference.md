Vibescript is a scoped extension language: scripts extend a Go application
the way Lua extends a game, written by users or agents and running inside
bounds the host defines. This page covers the language and the runtime that
bounds it: literals through classes, typing, concurrency, and the sandbox's
host configuration. For the full embedding API, capability adapters, and
method-level coverage, see the
[upstream docs](https://github.com/mgomes/vibescript/tree/master/docs).

## Basics {#basics}

### Source structure {#source-structure}

Files are UTF-8 text, typically with a `.vibe` extension. A `#` starts a
comment that runs to end-of-line.

Top-level declarations are functions, classes, modules, and enums. Executable
top-level statements form the default script body when a file is run directly,
and form a module initializer when a file is loaded with `require`. Hosts that
embed Vibescript usually call a named function instead, which is how the
runnable examples on this site work.

Statements are separated by newlines or semicolons, and expressions can be
used as statements.

### Values & literals {#literals}

```vibe
missing = nil
active = true
count = 42
price = 19.99
label = :active
greeting = "hello #{name}"
tags = ["alpha", "beta"]
config = {retries: 3, verbose: true}
window = 1..5
grace = 2.days
```

The literal categories are `nil`, `true`, `false`; integers and floats;
strings; symbols (`:name`, or quoted as `:"with-punctuation"`); arrays;
hashes; ranges; and duration literals such as `5.minutes` or `2.days`.

Ranges with `..` include the final endpoint, ranges with `...` exclude it.
Ruby's open-ended ranges are supported: `arr[1..]` takes everything from index
1 on, `s[..2]` the leading characters, and `when 3..` matches three and up.
Open ranges cannot be iterated: `each`, `map`, `to_a`, `for`, and friends
reject them up front rather than running into the sandbox quotas. Unlike Ruby,
descending ranges iterate: `(5..1).to_a` is `[5, 4, 3, 2, 1]` rather than an
empty array.

### Numbers {#numbers}

Integers are arbitrary precision, as in Ruby: arithmetic that leaves the
signed 64-bit range promotes transparently, and a result that fits 64 bits
again returns to the compact fast representation. There is a single `int`
type; scripts never observe a separate "bignum" kind. A few surfaces
deliberately stay within 64 bits and raise a clear error for larger values:
range endpoints, iteration counts (`times`, `upto`, `downto`, `step`),
`Money`/`Duration`/`Time` arithmetic, and argument positions that denote
indexes, counts, sizes, or precisions.

```vibe
big = 2 ** 100          # => 1267650600228229401496703205376, exact
readable = 1_000_000    # underscores are visual separators
hex = 0xDEAD_BEEF       # 0x, 0b, 0o, and 0d base prefixes
sci = 1.5e-2            # floats may use scientific notation
kilo = 1e3              # any literal with an exponent is a float: 1000.0
```

Numeric literals accept underscores between digits in any base. An exponent
marker (`e`/`E`) takes an optional sign and one or more digits; a literal
whose exponent overflows the 64-bit float range saturates to `Infinity`. A
numeric literal may not directly abut an identifier: `123abc` and `1.5x` are
parse errors rather than two tokens, matching Ruby. A bare leading zero
(`010`) stays decimal rather than being read as legacy octal.

### Strings & symbols {#strings}

Double-quoted strings support `#{...}` interpolation. Each interpolation
holds one expression, converted with the same string form used by `to_s`. The
expression may contain its own double-quoted strings and even nested
interpolations; the interpolation extends to its matching `}`. Escape a
literal marker as `\#{...}`. Single-quoted strings do not interpolate.

```vibe
def describe(name)
  "#{name || "guest"} checked in"
end
```

Symbols are usually written bare (`:name`), but a quoted form lets a symbol
hold punctuation, spaces, or be empty: `:"foo-bar"`, `:'foo bar'`, `:""`.
Quoted symbols use the same escapes as the matching string quote, and
interpolation is not supported in symbol literals.

In a hash literal, a bare label makes a symbol key and a quoted label makes a
string key: `{ name: 1 }` is read back with `h[:name]`, while `{ "name": 1 }`
is read back with `h["name"]`. The quoted form is the only literal syntax for
a string-keyed hash, which is what `JSON.parse` returns. Ruby's hash rocket
syntax (`=>`) is not supported.

Strings are immutable values. Reading with `[]` mirrors Ruby's `String#[]`
and `Array#[]`, including negative indexes, `value[start, length]`, and
`value[range]` slices.

### Variables & assignment {#variables}

Variables are dynamically bound by assignment. Parallel and destructuring
assignment split array values across targets:

```vibe
a, b = [1, 2]
first, *middle, last = [1, 2, 3, 4]
x, (y, z) = [1, [2, 3]]
```

Missing values bind as `nil`, extra values are ignored unless captured by a
`*rest` target, and scalar right-hand values are treated as one value. A bare
`*` is an anonymous rest target that discards what it captures. It can sit
at the front, middle, or end (`*, last = [1, 2, 3]`).

Index assignment works on mutable collections, and array targets accept a
negative index that counts back from the end:

```vibe
items = [1, 2, 3]
items[0] = 10
items[-1] = 30
```

Compound assignment is supported for variables, member targets, and index
targets with `+=`, `-=`, `*=`, `/=`, `%=`, and `**=`:

```vibe
total += amount
items[0] *= 2
record[:score] **= 2
```

## Functions {#functions}

### Defining functions {#defining-functions}

Define functions with `def`/`end`. The last evaluated expression is the
return value, and `return` exits early:

```vibe
def add(a, b)
  a + b
end
```

Parameters and returns take optional type annotations, checked as runtime
contracts at typed boundaries:

```vibe
def charge(amount: int, currency: string = "USD") -> hash
  {amount: amount, currency: currency}
end
```

### Parameter forms {#parameters}

A parameter's spelling chooses how it receives a value. The token after the
colon disambiguates the keyword and typed forms:

| Form | Meaning |
| --- | --- |
| `name` | required positional parameter |
| `name = default` | optional positional parameter |
| `name: Type` | typed positional parameter |
| `name: Type = default` | typed positional parameter with a default |
| `name:` | required keyword-only parameter |
| `name: default` | optional keyword-only parameter |
| `*rest` | captures extra positional arguments |
| `**rest` | captures extra keyword arguments |
| `&block` | captures a passed block |

A keyword-only parameter is bound only by a matching keyword label; it never
accepts a positional argument. The optional form supplies its default when
the label is omitted, and a later default may reference an earlier parameter:

```vibe
def connect(host:, port: 8080, scheme: "https", timeout: port * 2)
  "#{scheme}://#{host}:#{port}"
end

def demo
  connect(host: "example.com")            # port 8080, scheme "https"
  connect(host: "example.com", port: 443) # overrides port
end
```

Because `name: Type` declares a typed positional parameter, a bare identifier
after the colon resolves as a type name, not a keyword default: write
`a: int` for a typed positional and `a: 0` for an optional keyword. A default
that is *only* a reference to an earlier parameter must be parenthesized:
`timeout: port * 2` is a default, but `timeout: port` reads as a type, so
write `timeout: (port)`.

### Function values {#function-values}

A function referenced by name (without calling it) is a value that can be
passed around and invoked. Direct `fn(...)` invocation and Ruby-style
`fn.call(...)` behave identically, forwarding positional arguments, keyword
arguments, and an optional block:

```vibe
def inc(n)
  n + 1
end

def apply_twice(fn, value)
  fn.call(fn(value))
end

def demo
  apply_twice(inc, 40) # => 42
end
```

The only member exposed on a function value is `call`. A zero-arity function
is auto-invoked when referenced by name, so it cannot currently be passed as
a function value.

## Calls & Blocks {#calls}

### Method calls {#method-calls}

Calls support positional and keyword arguments, and may omit parentheses when
all arguments stay on the same line:

```vibe
def demo(fees, amount)
  fees.apply(amount)
  fees.apply amount
  render status: "ok"
end
```

Positional arguments must come before keyword labels: `collect(first: 1,
"tail")` is a parse error, while `collect("head", first: 1)` is accepted.

Label arguments bind as keyword arguments when the callee accepts them. When
a function has a positional hash parameter instead, the same source form is
passed as a final options hash, in both parenless and parenthesized form:

```vibe
def configure(opts)
  opts[:retries]
end

def demo
  configure(retries: 3)  # => 3
  configure retries: 3   # same call
end
```

Constructor calls (`Klass.new(...)`) and method calls (`receiver.method(...)`)
keep strict parenthesized keyword binding: a parenthesized keyword with no
matching keyword parameter does not collapse into a positional options hash.

A local variable already holding the value of a keyword can be passed with
the shorthand `greet(name:)`, which is `greet(name: name)`.

### Splats & parenless arguments {#splats}

Ruby-style call splats expand prepared argument lists in place: `f(*args)`
spreads an array into positional arguments, `f(**opts)` spreads a hash into
keyword arguments (string or symbol keys; later duplicates win), and both
combine freely with regular arguments and blocks:

```vibe
def sum3(a, b, c)
  a + b + c
end

def demo
  args = [2, 3]
  sum3(1, *args) # => 6
end
```

The expansion happens before binding, so arity, keyword, and type errors
match the equivalent literal call.

In parenless calls, spacing disambiguates: `f *args` splats, while `a * b`
and `a*b` stay multiplication. The same rule lets a regex or an array literal
be a parenless command argument (`match /ID-[0-9]+/`, `puts [3, 1, 2].sort`).
A sigil detached from a non-local callee but flush against its operand
opens an argument, and every other spacing keeps the operator reading.

### Blocks {#blocks}

Blocks are lightweight lambdas passed with `do ... end` or braces. Block
parameters default to `nil` when fewer values are provided, and can
destructure the yielded value exactly like assignment destructuring:

```vibe
def active_names(players)
  players
    .select do |player|
      player[:active]
    end
    .map do |player|
      player[:name]
    end
end

def firsts(rows)
  rows.map do |(head, *)|
    head
  end
end
```

A function runs its caller's block with `yield`, and `block_given?` reports
whether the current call was given one:

```vibe
def fetch(default)
  if block_given?
    yield
  else
    default
  end
end

def demo
  fetch("none")             # => "none"
  fetch("none") { "value" } # => "value"
end
```

As in Ruby, an explicit `return` inside a block returns from the method whose
body created the block, ending iteration immediately. `ensure` blocks on the
way out still run. A block without a parameter list infers implicit
parameters (`it`, `_1`..`_9`).

### Procs & lambdas {#lambdas}

Blocks become first-class values through `Proc.new { ... }`, `proc { ... }`,
`lambda { ... }`, and the stabby lambda `->(args) { ... }`. All four produce
values invoked with `.call`:

```vibe
def demo
  double = ->(n) { n * 2 }
  add = lambda do |a, b|
    a + b
  end
  add.call(double.call(20), 2) # => 42
end
```

Procs and lambdas differ exactly as in Ruby. A **proc** keeps block
semantics: missing arguments pad to `nil`, extra arguments are dropped, a
single array argument auto-splats, and `return` in the body returns from the
method that created it. A **lambda** behaves like an anonymous method: arity
is strict, and `return`, `break`, and `next` are all local to the lambda.
`fn.lambda?` reports which semantics a callable carries.

The ampersand converts a value into the call's block: `m(&blk)` forwards a
captured block, proc, function value, or bound method, and `m(&:name)` is
the symbol-to-proc shorthand:

```vibe
def shout(words)
  words.map(&:upcase)  # => ["A", "B"] for ["a", "b"]
end

def total(numbers)
  numbers.reduce(&:+)  # => 6 for [1, 2, 3]
end
```

The `&` argument must be last, appears at most once, and cannot be combined
with a literal block.

### Safe navigation {#safe-navigation}

`receiver&.member` reads a member or calls a method only when the receiver is
not `nil`; otherwise the whole access short-circuits to `nil` without
evaluating arguments or block:

```vibe
def demo(user)
  user&.name
  user&.profile("public")
  user&.profile&.name
end
```

The operator guards only its immediate access. In `user&.profile.name` the
trailing `.name` still dispatches on whatever `user&.profile` returned, so
guard each link in a chain. Safe navigation cannot be used as an assignment
target; `user&.name = "Ada"` is a parse error.

## Operators {#operators}

### Operator families {#operator-families}

- Arithmetic: `+`, `-`, `*`, `/`, `%`, `**`
- Comparison: `==`, `!=`, `<`, `<=`, `>`, `>=`, `<=>`
- Case equality: `===`
- Regex match: `text =~ /re/` (character index of the first match, or `nil`)
  and `text !~ /re/` (`true` when the pattern does not match)
- Boolean: `&&`, `||`, unary `!`
- Collection: `array << value` (append), `array & other` (intersection)
- Unary sign: prefix `-` negates a number; prefix `+` is the identity on
  numbers and strings
- Conditional: `condition ? when_true : when_false`

The Ruby word forms `and`, `or`, and `not` are **not** boolean operators in
Vibescript. They are ordinary identifiers, so they can be used as method
names, function names, and hash labels. Use `&&`, `||`, and `!`.

`array << value` appends in place and returns the receiver, exactly like
Ruby's shovel. `array & other` returns a new array holding the elements
common to both, duplicates removed, left order preserved. Following Ruby,
`+` binds tighter than `<<`, which binds tighter than `&`.

### Comparison & case equality {#comparison}

The spaceship `<=>` returns `-1`, `0`, or `1` for ordered operands and `nil`
when the operands cannot be ordered (different kinds, money in different
currencies, or a `NaN` on either side). The relational operators `<`, `<=`,
`>`, `>=` instead raise on incomparable operands, matching Ruby's
`ArgumentError`.

`===` treats its left operand as a matcher, mirroring `case`/`when`. A range
matcher checks membership, a regex matcher tests a string, and every other
matcher falls back to `==`:

```vibe
def demo
  (1..3) === 2       # => true
  /el+/ === "hello"  # => true
  1 === 1.0          # => false; int and float stay distinct kinds
end
```

### Precedence & continuation {#precedence}

Precedence follows conventional arithmetic/boolean ordering. `**` is
right-associative and binds more tightly than unary `-`, so `-2 ** 2` is
`-(2 ** 2)`. Integer powers stay `int` for non-negative exponents, promoting
to arbitrary precision past 64 bits; mixed numeric powers and negative
integer exponents return `float`.

Division follows Ruby: integer division by zero (`1 / 0`) raises, while
float division by zero (`1.0 / 0`) follows IEEE 754 and yields `Infinity`,
`-Infinity`, or `NaN`; inspect those with `Float#nan?`, `Float#infinite?`,
and `Float#finite?`. `&&` binds tighter than `||`, and ternary conditionals
sit below `||`, associate to the right, and evaluate only the selected
branch.

A leading `+` or `-` on a fresh line follows Vibescript's
indented-continuation rule, which intentionally differs from Ruby: flush
against its operand it begins a new statement, while separated by whitespace
it continues the previous line as a binary operator, so multi-line
arithmetic can be indented under its first operand:

```vibe
def demo(total, amount)
  total
    + amount
end
```

## Control Flow {#control-flow}

### Conditionals {#conditionals}

`if` / `elsif` / `else` and `unless` / `else` are statements and
value-producing expressions. When no branch matches and there is no `else`,
the expression returns `nil`:

```vibe
def label(score)
  if score >= 90
    "great"
  elsif score >= 80
    "passing"
  else
    "retry"
  end
end
```

Short statements can use modifier conditionals and ternaries:

```vibe
def demo(active, suspended)
  status = "open" unless suspended
  active ? "active" : "inactive"
end
```

### case / when {#case-when}

`case` evaluates to the matching branch expression (or `nil` when nothing
matches and no `else` is provided). `when` candidates use `===` semantics:
ranges test membership, regexes test strings, and other values use equality.
Use `then` for compact single-line bodies:

```vibe
def label(score)
  case score
  when 100 then "perfect"
  when 90, 95 then "great"
  when 80..99 then "passing"
  else "ok"
  end
end
```

Targetless `case` evaluates each `when` expression as a predicate in order:

```vibe
def bucket(score)
  case
  when score == 100 then "perfect"
  when score >= 80 then "passing"
  else "ok"
  end
end
```

### Loops {#loops}

`while` and `until` loop over a condition; `for ... in` iterates arrays,
ranges, and hashes. `do` may be used as an optional body separator after the
condition or iterable:

```vibe
def countdown(n)
  out = []
  while n > 0 do
    out << n
    n -= 1
  end
  out
end

def sum_first_five
  total = 0
  for n in 1..5
    total += n
  end
  total
end
```

A `for` loop over a hash binds a two-element `[key, value]` pair per
iteration, visiting entries in insertion order. `break` and `next` target the
nearest active loop, raise when used outside any loop, and cannot cross call
boundaries. Short statements also come in modifier-loop form:

```vibe
def demo(i)
  i += 1 while i < 3
  i -= 1 until i <= 0
  i
end
```

Every loop iteration consumes a step against the sandbox's step quota, so an
accidental infinite loop terminates with a quota error instead of hanging
the host.

### Error handling {#errors}

Raise explicit failures with `raise`, and handle them with
`begin` / `rescue` / `ensure`. A rescue clause can bind the error and read
its message:

```vibe
def safe_divide(a, b)
  begin
    a / b
  rescue RuntimeError => err
    "failed: #{err.message}"
  ensure
    cleanup
  end
end
```

### Guard clauses {#guards}

`return` works with modifier conditionals for early exits:

```vibe
def ship(order)
  return "missing" unless order
  return "empty" if order[:items] == []
  "shipped"
end
```

## Types & Structure {#types}

### Classes {#classes}

Class declarations group behavior and state. Instances are created with
`.new`, and methods take the same signatures as functions:

```vibe
class Counter
  def bump(value: int) -> int
    value + 1
  end
end

def demo
  Counter.new.bump(1) # => 2
end
```

Inheritance is not supported. Instance variables (`@name`), class variables
(`@@count`), accessors, mixins, and visibility are covered in the upstream
[classes guide](https://github.com/mgomes/vibescript/blob/master/docs/classes.md).

### Modules {#modules}

Module declarations group module functions and constants under a namespace.
Their instance-style methods mix into classes with `include` (instance
methods) or `extend` (class methods):

```vibe
module Billing
  LIMIT = 5

  def self.code
    "ok"
  end
end

def demo
  Billing.code   # => "ok"
  Billing::LIMIT # => 5
end
```

`module` is contextual, not reserved: it starts a declaration only when
followed by a constant name. Declarations may nest (`Outer::Inner`), and
modules cannot be instantiated.

Load shared code from other files with `require`. File-based modules are
distinct from in-source `module` declarations, and resolution is governed by
the host's `Config.ModulePaths` and policy lists:

```vibe
def demo(input)
  helpers = require("public/helpers", as: "helpers")
  helpers.normalize(input)
end
```

### Enums {#enums}

Enums declare nominal state sets, with members accessed via `::`:

```vibe
enum Status
  Draft
  Published
end

def demo
  Status::Draft
end
```

Coercion, equality, and serialization behavior are covered in the upstream
[enums guide](https://github.com/mgomes/vibescript/blob/master/docs/enums.md).

### Gradual typing {#typing}

Typing is optional and gradual: annotate parameters and returns where
helpful, and rely on runtime contract checks at typed boundaries. The checker
also infers local types to catch known contradictions before execution;
unannotated values simply remain gradual.

Type names are case-insensitive: `int`, `float`, `number`, `string`, `bool`,
`nil`, `duration`, `time`, `money`, `array`, `hash`/`object`, `range`,
`function`, top-level enum names, and `any`. Append `?` for nullable
(`string?`, `int?`), join alternatives with `|` (`int | string`), and
parameterize containers with `array<T>` and `hash<K, V>`.

Shape types describe hash payloads field by field. A `?` on a field name
marks it optional, and a trailing `...` leaves the shape open to extra keys:

```vibe
def apply_bonus(payload: { id: string, points: int }) -> { id: string, points: int }
  { id: payload[:id], points: payload[:points] + 5 }
end

def register(user: { name: string, age?: int, ... }) -> string
  user[:name]
end
```

## Runtime {#runtime}

### Built-ins {#builtins}

Notable built-in facilities include assertion and conversion helpers,
`Time`, `Duration`, and `Money` values with quota-guarded arithmetic, and the
`JSON` and `Regex` utility families. Durations read naturally off integers:

```vibe
def demo
  5.minutes
  2.days
end
```

The full API surface (string, array, hash, and range methods included) is
documented in the upstream
[builtins](https://github.com/mgomes/vibescript/blob/master/docs/builtins.md)
and
[stdlib](https://github.com/mgomes/vibescript/blob/master/docs/stdlib_core_utilities.md)
guides.

### Tasks & concurrency {#tasks}

`Tasks` runs independent named functions concurrently while the runtime keeps
the work bounded and scoped. Concurrency is **structured**: a task cannot
outlive the `Tasks.run` or `Tasks.map` scope that created it, leaving a scope
waits for every spawned task, and failures report through `task.value` or at
scope exit.

`Tasks.map` runs each input through the same named function and returns
results in input order (not completion order). Limit fanout for one call with
`max:`:

```vibe
def score_user(user)
  user[:score] * user[:weight]
end

def score_users(users)
  Tasks.map(users, max: 2, with: :score_user)
end
```

`Tasks.run` is the manual scope: `tasks.spawn(:function_name, arg, key: value)`
starts a named function and returns a handle, and `task.value` waits for that
task and returns its result or raises its error. The block's value is the
scope's value:

```vibe
def prepare_user(user)
  "prepared:" + user[:id]
end

def prepare_pair(first, second)
  Tasks.run(max: 2) do |tasks|
    left = tasks.spawn(:prepare_user, first)
    right = tasks.spawn(:prepare_user, second)

    [left.value, right.value]
  end
end
```

Scope exit waits automatically, so `tasks.wait` is only needed as an explicit
barrier when later code in the same block must wait for spawned work.

Tasks run through fresh execution state. They inherit the parent call's
capabilities, globals, strict-effects policy, and cancellation context, but
they do not share mutable locals or captured block state. Arguments,
results, and inherited globals are **cloned** across the task boundary, and
must be data-only: functions, blocks, capabilities, and cyclic structures
cannot cross it. Results retained by task handles count against the parent's
memory quota while the scope is alive.

The host bounds all fanout: `DefaultTaskConcurrency` applies when a script
omits `max:`, and `MaxTaskConcurrency` caps script-provided values. A
request above the cap raises (`Tasks.map max 99 exceeds host maximum 64`)
rather than being silently clamped.

### Sandbox & quotas {#sandbox}

Vibescript is safe by default. Every run is bounded by three quotas:
**steps**, **memory**, and **recursion depth**. Every loop iteration
(even with an empty body), call, and allocation is charged against them.
Splat-expanded arguments cost the same as literal ones, so no call shape
escapes accounting. When a script exceeds a quota it terminates with a clear
error (`step quota exceeded`, `memory quota exceeded`, `recursion depth
exceeded`) instead of degrading the host.

Recursion is deliberately never unlimited: the interpreter recurses on the
host's Go stack, so even the most generous profile keeps a finite cap that
fails cleanly on runaway recursion rather than crashing the process.

Scripts cannot reach the filesystem, network, or clock on their own. Side
effects enter only through what the host passes in: data globals, and typed
**capability adapters** that validate arguments and results at the boundary
(everything crossing it must be data-only; callables are rejected). With
`StrictEffects` enabled, even host-seeded globals must be data-only, forcing
every side effect through an auditable adapter. Host context cancellation
propagates into running scripts, including spawned tasks.

The [playground on this site](/) enforces a deliberately tight budget on
every run; the exact values are shown on the homepage.

### Host configuration {#host-config}

Hosts embed the interpreter by constructing an engine from `vibes.Config`;
the zero value is a working, conservatively-limited sandbox:

```go
engine, err := vibes.NewEngine(vibes.Config{
    StepQuota:              20_000,
    MemoryQuotaBytes:       256 << 10, // 256 KiB
    RecursionLimit:         32,
    StrictEffects:          true,
    DefaultTaskConcurrency: 4,
    MaxTaskConcurrency:     16,
    ModulePaths:            []string{"/srv/vibes/modules"},
})
```

| Field | Default | Controls |
| --- | --- | --- |
| `StepQuota` | 1,000,000 | execution steps per call; loops, calls, and allocations all charge it |
| `MemoryQuotaBytes` | 16 MiB | live interpreter memory per call |
| `RecursionLimit` | 256 | call depth; always finite, even in the most generous profile |
| `StrictEffects` | `false` | require host globals to be data-only; side effects go through capability adapters |
| `ModulePaths` | none | directories searched by `require` |
| `ModuleAllowList` / `ModuleDenyList` | none | policy filter on which modules may load |
| `OutputWriter` / `ErrorWriter` | unset | destinations for `puts`/`print`/`p` and `warn`; unset makes those builtins raise |
| `RandomReader` / `RandomReadFunc` | `crypto/rand` | the randomness source scripts observe |
| `MaxSourceBytes` | 1 MiB | largest source a single compile accepts |
| `MaxCachedModules` | 1,000 | compiled-module cache bound |
| `DefaultTaskConcurrency` | 4 | task fanout when a script omits `max:` |
| `MaxTaskConcurrency` | 64 | hard cap on script-requested fanout; above it raises |
| `DevMode` | `false` | development-only module reloading on file change |

Quota fields read zero as "use the default" and `vibes.Unlimited` as
"disable this quota". Instead of tuning each knob, a host can apply a named
profile, a coherent budget bundle. In ascending generosity:

| Profile | Steps | Memory | Recursion |
| --- | --- | --- | --- |
| `low` | 1,000,000 | 16 MiB | 256 |
| `medium` | 20,000,000 | 128 MiB | 1,000 |
| `high` | 200,000,000 | 512 MiB | 4,000 |
| `xhigh` | unlimited | unlimited | 10,000 |

`vibes.ProfileHigh.ApplyTo(&cfg)` writes only the three quota fields, and
`vibes.QuotaProfileByName("medium")` resolves a user-supplied name. The
`vibes` CLI runs on the same ladder and defaults to `xhigh`, since it runs your
own scripts and is not a sandbox. Capability adapters, module policy, and
per-call options (`CallOptions.Globals`, `CallOptions.Capabilities`) are
covered in the upstream
[integration guide](https://github.com/mgomes/vibescript/blob/master/docs/integration.md)
and
[host cookbook](https://github.com/mgomes/vibescript/blob/master/docs/host_cookbook.md).
