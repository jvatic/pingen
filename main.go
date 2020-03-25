package main

import (
	"bufio"
	crypto_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/flynn/flynn/pkg/shutdown"
	docopt "github.com/flynn/go-docopt"
)

var r *rand.Rand

func init() {
	var b [8]byte
	crypto_rand.Read(b[:])
	r = rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(b[:]))))
}

func main() {
	defer shutdown.Exit()

	log.SetFlags(0)

	usage := `
usage: pingen <length> [<blacklist>...]
       pingen -h | --help

Options:
	-h, --help  Show this screen
`[1:]
	args, err := parseArgs(usage)
	if err != nil {
		log.Fatal(err)
	}
	var pin []string
	for {
		pin = randomPin(args.Length)
		if !hasPin(args.Blacklist, pin) {
			break
		}
	}
	fmt.Println(strings.Join(pin, " "))
}

type Args struct {
	Length    int
	Blacklist Blacklist
}

type Blacklist map[string]struct{}

func parseArgs(usage string) (*Args, error) {
	cliArgs, err := docopt.Parse(usage, nil, true, "0.1.0", true)
	if err != nil {
		return nil, err
	}
	blacklist := make(Blacklist)
	paths, ok := cliArgs.All["<blacklist>"].([]string)
	if !ok {
		paths = []string{}
	}
	for _, p := range paths {
		if err := parseBlacklist(p, blacklist); err != nil {
			log.Println(err)
		}
	}
	args := &Args{
		Blacklist: blacklist,
	}

	length, err := strconv.Atoi(cliArgs.String["<length>"])
	if err != nil {
		return nil, err
	}
	args.Length = length
	return args, nil
}

func parseBlacklist(path string, blacklist Blacklist) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		blacklist[strings.TrimSpace(s.Text())] = struct{}{}
	}
	if err := s.Err(); err != nil {
		return err
	}
	return nil
}

func hasPin(blacklist Blacklist, pin []string) bool {
	_, ok := blacklist[strings.Join(pin, "")]
	return ok
}

func randomPin(length int) []string {
	pin := make([]string, length)
	for i := 0; i < length; i++ {
		pin[i] = fmt.Sprintf("%d", r.Intn(10))
	}
	return pin
}
