package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	pkt := &pktlines{
		in:  os.Stdin,
		out: os.Stdout,
	}
	// S: PKT-LINE(version=1\0push-options ...)
	version, err := pkt.nextLine()
	if err != nil {
		log.Fatal(err)
	}
	if !strings.HasPrefix(version, "version=1") {
		log.Fatalf("unsupported version: `%s`", version)
	}

	// H: PKT-LINE(version=1\0push-options ...)
	if err := pkt.write("version=1\000push-options"); err != nil {
		log.Fatal(err)
	}

	// S: PKT-LINE(<old-oid> <new-oid> <ref-name>)
	reflines, err := pkt.nextLine()
	if err != nil {
		log.Fatal(err)
	}

	// S: PKT-LINE(push-options)
	// ignore for test
	if _, err := pkt.nextLine(); err != nil {
		log.Fatal(err)
	}

	reflines = strings.TrimSpace(reflines)
	if split := strings.Split(reflines, "\n"); len(split) > 1 {
		log.Fatal("unexpected reflines size")
	} else {
		reflines = split[0]
	}

	args := strings.Fields(reflines)
	oldref := args[0]
	newref := args[1]
	refname := args[2]

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	git, err := exec.LookPath("git")
	if err != nil {
		panic(err)
	}
	var stderr bytes.Buffer
	cmd := exec.Cmd{
		Path:   git,
		Args:   []string{"git", "update-ref", refname, newref, oldref},
		Env:    os.Environ(),
		Dir:    wd,
		Stdin:  nil,
		Stdout: nil,
		Stderr: &stderr,
	}
	if err := cmd.Run(); err != nil {
		pkt.write(fmt.Sprintf("ng %s update-ref error: %s", refname, err, string(stderr.Bytes())))
	} else {
		pkt.write(fmt.Sprintf("ok %s", refname))
	}
}

func debug(msg string) {
	fmt.Fprintln(os.Stderr, "[proc-receive]: "+msg)
}

type pktlines struct {
	in  io.Reader
	out io.Writer
}

func (p *pktlines) nextLine() (string, error) {
	var res []byte
	blen := make([]byte, 4)

	for {
		if r, err := p.in.Read(blen); err != nil {
			return "", fmt.Errorf("read len: %w", err)
		} else if r < 4 {
			return "", fmt.Errorf("too few len bytes %d", r)
		}

		if blen[0] == '0' && blen[1] == '0' && blen[2] == '0' && blen[3] == '0' {
			break
		}

		size, err := strconv.ParseInt(string(blen), 16, 32)
		if err != nil {
			return "", fmt.Errorf("read len value: %w", err)
		}
		buf := make([]byte, int(size)-4)
		if r, err := p.in.Read(buf); err != nil {
			return "", fmt.Errorf("read line: %w", err)
		} else if r < len(buf) {
			return "", fmt.Errorf("wrong line length: expect %d but was %d", len(buf), r)
		}

		if len(res) != 0 {
			res = append(res, '\n')
		}
		res = append(res, buf...)
	}
	debug("S: " + humanize(res))
	return string(res), nil
}

func humanize(bin []byte) string {
	var sb strings.Builder
	for _, b := range bin {
		if b == '\000' {
			sb.WriteString("\\0")
		} else {
			sb.WriteByte(b)
		}
	}
	return sb.String()
}

var (
	bflush = [4]byte{'0', '0', '0', '0'}
)

func (p *pktlines) write(buf string) error {
	res := []byte(fmt.Sprintf("%04x", len(buf)+4))
	res = append(res, []byte(buf)...)
	debug("H: " + humanize(res))
	res = append(res, bflush[:]...)
	if w, err := p.out.Write(res); err != nil {
		return fmt.Errorf("write error: %w", err)
	} else if w < len(res) {
		return fmt.Errorf("too few bytes written: sent %d written %d", len(res), w)
	}
	return nil
}
