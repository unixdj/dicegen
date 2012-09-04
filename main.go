// Copyright 2001 Vadim Vygonets.  No rights reserved.
// This program is free software.  It comes without any warranty, to
// the extent permitted by applicable law.  You can redistribute it
// and/or modify it under the terms of the Do What The Fuck You Want
// To Public License, Version 2, as published by Sam Hocevar.  See
// the LICENSE file or http://sam.zoy.org/wtfpl/ for more details.

// Diceware8k / base64 / hex password generator
package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

var randBits struct {
	bits uint64
	have uint
}

// getBits returns n random bits.
// Breaks if n>57 (64-8+1).
func getBits(n uint) uint64 {
	if n > randBits.have {
		need := n - randBits.have
		buf := make([]byte, (need+7)/8)
		_, err := io.ReadFull(rand.Reader, buf)
		if err != nil {
			panic(err)
		}
		for _, v := range buf {
			randBits.bits |= uint64(v) << randBits.have
			randBits.have += 8
		}
	}
	res := randBits.bits & (1<<n - 1)
	randBits.bits >>= n
	randBits.have -= n
	return res
}

type engine struct {
	bits uint
	dlen int
	sep  string
	gets func(n uint64) string
}

var b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
var engines = []engine{
	{13, 5, " ", func(n uint64) string { return Dicewds8k[n] }},
	{6, 16, "", func(n uint64) string { return b64[n : n+1] }},
	{4, 16, "", func(n uint64) string { return "0123456789abcdef"[n : n+1] }},
}

func parseFlags() (*engine, int) {
	var (
		b = flag.Bool("b", false, "select base64 passwords")
		h = flag.Bool("h", false, "select hex passwords")
		e = &engines[0]
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"diceware2k/base64/hex passphrase generator\n"+
				"Usage: %s [-b|-h] [n]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "  n: password length (default: 5 words (diceware8k) or 16 characters (the rest))\n")
	}
	flag.Parse()
	if *b {
		if *h {
			fmt.Fprintf(os.Stderr, "-b or -h don't mix\n")
			flag.Usage()
			os.Exit(2)
		}
		e = &engines[1]
	} else if *h {
		e = &engines[2]
	}
	t := e.dlen
	switch flag.NArg() {
	case 0:
	case 1:
		s := flag.Arg(0)
		x, err := strconv.ParseUint(s, 10, 32)
		if err != nil || x < 1 || x > 1024 {
			fmt.Fprintf(os.Stderr,
				"%s must be a small positive integer\n", s)
			flag.Usage()
			os.Exit(2)
		}
		t = int(x)
	default:
		flag.Usage()
		os.Exit(2)
	}
	return e, t
}

func main() {
	e, t := parseFlags()
	fmt.Print(e.gets(getBits(e.bits)))
	for i := 1; i < t; i++ {
		fmt.Print(e.sep, e.gets(getBits(e.bits)))
	}
	fmt.Print("\n")
}
