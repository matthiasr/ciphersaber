package main

import(
  "bufio"
  "crypto/rand"
  "flag"
  "fmt"
  "io"
  "os"
)

var debug bool


func initial_s_box(key []byte) [256]byte {
  var S [256]byte
  var i, j int
  var tmp byte

  if debug {
    fmt.Fprintf(os.Stderr, "Key: ")
    for _, b := range key {
      fmt.Fprintf(os.Stderr, "%x ", b)
    }
    fmt.Fprintf(os.Stderr, "\n")
  }

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

  if debug {
    fmt.Fprintf(os.Stderr, "S-Box: ")
    for _, b := range S {
      fmt.Fprintf(os.Stderr, "%x ", b)
    }
    fmt.Fprintf(os.Stderr, "\n")
  }

  return S
}

func rc4_stream(key []byte, out chan<- byte) {
  S := initial_s_box(key)
  var i, j int
  var tmp byte

  for {
    i = (i + 1) & 255
    j = (j + int(S[i])) & 255
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

func writeout(out <-chan byte, done chan<- bool) {
  stdout := bufio.NewWriter(os.Stdout)

  for b := range out {
    err := stdout.WriteByte(b)
    if err != nil {
      panic(err)
    }
  }

  stdout.Flush()
  done<- true
}

func main() {
  // decode := flag.Bool("d", false, "Decode input") /* no function because IV handling not implemented & RC4 itself is symmetric */
  flag.BoolVar(&debug, "v", false, "Verbose output")
  decode := flag.Bool("d", false, "Decode")
  flag.Parse()
  flag.Usage = func() {
    fmt.Fprintf(os.Stderr, "Usage:\n\t%s [-d] <key>", os.Args[0])
    flag.PrintDefaults()
  }

  stdin := bufio.NewReader(os.Stdin)

  in := make(chan byte)
  out := make(chan byte)
  done := make(chan bool)

  go writeout(out, done)

  IV := make([]byte, 10)
  var err error
  if *decode {
    for i, _ := range IV {
      IV[i], err = stdin.ReadByte()
      if err != nil {
        panic(err)
      }
    }
  } else {
    n, err := io.ReadFull(rand.Reader, IV)
    if n != 10 || err != nil {
      panic(err)
    }
    for _, b := range IV {
      out<- b
    }
  }

  user_key := []byte(flag.Arg(0))
  key := make([]byte, len(user_key) + 10)
  copy(key, user_key)
  copy(key[len(user_key):], IV)

  go encode(key, in, out)

  for {
    b, err := stdin.ReadByte()
    if err == io.EOF {
      break
    } else if err != nil {
      panic(err)
    }
    in<- b
  }

  close(in)
  // out will be closed by encode()
  <-done //wait for output goroutine to finalize
}

