package repl

import (
  "JFFMonkeyLang/src/lexer"
  "JFFMonkeyLang/src/parser"
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
    p := parser.New(l)

    program := p.ParseProgram()

    // 4.check error
    if len(p.Errors()) != 0 {
      printParserErrors(out, p.Errors())
      continue
    }

    // 5.print result
    io.WriteString(out, program.String())
    io.WriteString(out, "\n")
  }
}

const MONKEY_FACE = `
            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func printParserErrors(out io.Writer, errors []string) {
  io.WriteString(out, MONKEY_FACE)
  io.WriteString(out, "Woops! We ran into some monkey business here!\n")
  io.WriteString(out, " parser errors:\n")
  for _, msg := range errors {
    io.WriteString(out, "\t"+msg+"\n")
  }
}
