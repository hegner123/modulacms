# CLI Planning

## Inputs
Input structs with methods

```go
type InputTypes string

var (
    textInput InputTypes = "textInput"
    textArea  InputTypes = "textArea"
    fileInput InputTypes = "fileInput""
)
type input struct{
    index   int
    key     any
    value   any
    history []any
}

func (i input) View(t InputTypes) string {
    switch t {
    case textInput:
    case textArea:
    case fileInput:

    }
}

func (i input) Update(value any) input {

}

```
