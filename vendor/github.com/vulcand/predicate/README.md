Predicate
=========

Predicate package used to create interpreted mini languages with Go syntax - mostly to define
various predicates for configuration, e.g. 

```
Latency() > 40 || ErrorRate() > 0.5.
```

Here's an example of fully functional predicate language to deal with division remainders:

```go
// takes number and returns true or false
type numberPredicate func(v int) bool

// Converts one number to another
type numberMapper func(v int) int

// Function that creates predicate to test if the remainder is 0
func divisibleBy(divisor int) numberPredicate {
     return func(v int) bool {
         return v%divisor == 0
     }
}

// Function - logical operator AND that combines predicates
func numberAND(a, b numberPredicate) numberPredicate {
    return func(v int) bool {
        return a(v) && b(v)
    }
}

func main(){
    // Create a new parser and define the supported operators and methods
    p, err := NewParser(Def{
        Operators: Operators{
            AND: numberAND,
        },
        Functions: map[string]interface{}{
            "DivisibleBy": divisibleBy,
        },
    })

    pr, err := p.Parse("DivisibleBy(2) && DivisibleBy(3)")
    if err == nil {
        fmt.Fatalf("Error: %v", err)
    }
    pr.(numberPredicate)(2) // false
    pr.(numberPredicate)(3) // false
    pr.(numberPredicate)(6) // true
}
```
