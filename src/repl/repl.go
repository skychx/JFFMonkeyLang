package repl

import (
  "JFFMonkeyLang/src/lexer"
  "JFFMonkeyLang/src/token"
  "bufio"
  "fmt"
  "io"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
  scanner := bufio.NewScanner(in)

  for {
    fmt.Fprint(out, PROMPT)
    // 1.read from command line input
    scanned := scanner.Scan()

    // 2.check input
    if !scanned {
      return
    }

    // 3.input data format to string
    line := scanner.Text()
    l := lexer.New(line)

    // 4.print all tokens until token.Type equal EOF
    for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
      fmt.Fprintf(out, "%+v\n", tok)
    }
  }
}
