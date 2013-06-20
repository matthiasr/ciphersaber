package main

import(
  "bufio"
  "flag"
  "fmt"
  "io"
  "os"
)

func initial_s_box(key []byte) [256]byte {
  var S [256]byte
  var i, j int
  var tmp byte
  for i, _ = range S {
    S[i] = byte(i & 255)
  }

  j = 0
  for i = 0; i <= 255; i++ {
    j = (j + int(S[i]) + int(key[i % len(key)])) & 255
    tmp = S[i]
    S[i] = S[j]
    S[j] = tmp
  }
  return S
}

func rc4_stream(key []byte, out chan<- byte) {
  S := initial_s_box(key)
  var i, j int
  var tmp byte

  for {
    i = (i + 1) & 255
    j = (j + int(S[j])) & 255
    tmp = S[i]
    S[i] = S[j]
    S[j] = tmp
    out<- S[(S[i] + S[j]) & 255]
  }
}

func encode(key []byte, in <-chan byte, out chan<- byte) {
  rc4 := make(chan byte)
  go rc4_stream(key, rc4)

  for b := range in {
    out<- ( b ^ <-rc4 )
  }
  close(out)
}

func writeout(out <-chan byte) {
  stdout := bufio.NewWriter(os.Stdout)

  defer func() {
    stdout.Flush()
  }()

  for b := range out {
    err := stdout.WriteByte(b)
    if err != nil {
      panic(err)
    }
  }
}

func main() {
  // decode := flag.Bool("d", false, "Decode input") /* no function because IV handling not implemented & RC4 itself is symmetric */
  flag.Parse()
  flag.Usage = func() {
    fmt.Fprintf(os.Stderr, "Usage:\n\t%s [-d] <key>", os.Args[0])
    flag.PrintDefaults()
  }
  key := []byte(flag.Arg(0))

  in := make(chan byte)
  out := make(chan byte)
  go encode(key, in, out)
  go writeout(out)

  stdin := bufio.NewReader(os.Stdin)

  for {
    b, err := stdin.ReadByte()
    if err == io.EOF {
      break
    } else if err != nil {
      panic(err)
    }
    in<- b
  }
}

