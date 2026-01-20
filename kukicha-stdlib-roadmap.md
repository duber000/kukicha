import slices

firstThree := slices.first(items, 3)
lastTwo := slices.last(items, 2)
tail := slices.drop(items, 3)
head := slices.dropLast(items, 1)

result := items
    |> slices.drop(2)
    |> slices.first(10)
    |> process()
