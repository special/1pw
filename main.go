package main

import (
	"flag"
	"fmt"
	"github.com/miquella/opvault"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

var vaultPath = flag.String("vault", "~/.1pw", "path to vault")
var revealPasswords = flag.Bool("reveal", false, "reveal passwords in output")
var profileName = flag.String("profile", "default", "profile name")
var noCopyPassword = flag.Bool("no-copy", false, "don't copy password to clipboard")

func main() {
	flag.Parse()

	vault, err := opvault.Open(expandHome(*vaultPath))
	if err != nil {
		log.Fatalf("open error: %s", err)
	}

	profile, err := vault.Profile(*profileName)
	if err != nil {
		log.Fatalf("profile error: %s", err)
	}

	fmt.Print("Password: ")
	password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("password input error: %s", err)
	}
	if err := profile.Unlock(string(password)); err != nil {
		log.Fatalf("unlock error: %s", err)
	}
	fmt.Println("\n")

	items, err := profile.Items()
	if err != nil {
		log.Fatalf("items error: %s", err)
	}

	fzfArgs := []string{"--height=20%", "--min-height=15", "--header=search items"}
	if len(flag.Args()) > 0 {
		fzfArgs = append(fzfArgs, []string{"-1", "--query=" + strings.Join(flag.Args(), " ")}...)
	}

	fzf := exec.Command("fzf", fzfArgs...)
	fzfIn, err := fzf.StdinPipe()
	fzf.Stderr = os.Stderr
	fzfOut, err := fzf.StdoutPipe()

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := fzf.Run(); err != nil {
			log.Fatalf("fzf error: %s", err)
		}
	}()

	for _, item := range items {
		fmt.Fprintln(fzfIn, item.Title())
	}
	fzfIn.Close()

	sel, err := ioutil.ReadAll(fzfOut)
	if err != nil {
		log.Fatalf("fzf read error: %s", err)
	}
	selected := strings.Trim(string(sel), "\n")
	<-done

	var didClipboard bool
	for _, item := range items {
		if item.Title() != selected {
			continue
		}

		detail, err := item.Detail()
		if err != nil {
			log.Fatalf("item detail error: %s", err)
		}

		fmt.Printf("%s\n\tCategory:%s\n\tTags:%s\n", item.Title(), item.Category(), item.Tags())
		for _, field := range detail.Fields() {
			printValue := field.Value()
			if !*revealPasswords && field.Type() == opvault.PasswordFieldType {
				printValue = "********"
			}
			if !didClipboard && !*noCopyPassword && field.Type() == opvault.PasswordFieldType {
				setClipboard(field.Value())
				printValue += " [copied]"
				didClipboard = true
			}
			fmt.Printf("\t%s (%s) = %s\n", field.Name(), field.Type(), printValue)
		}

		for _, section := range detail.Sections() {
			fmt.Printf("\t%s / %s\n", section.Name(), section.Title())
			for _, field := range section.Fields() {
				printValue := field.Value()
				if !*revealPasswords && field.Kind() == opvault.ConcealedFieldKind {
					printValue = "********"
				}
				if !didClipboard && !*noCopyPassword && field.Kind() == opvault.ConcealedFieldKind {
					setClipboard(field.Value())
					printValue += " [copied]"
					didClipboard = true
				}
				fmt.Printf("\t\t%s / %s (%s) = %s\n", field.Name(), field.Title(), field.Kind(), printValue)
			}
		}
	}
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		u, err := user.Current()
		if err != nil {
			log.Fatalf("user error: %s", err)
		}
		path = filepath.Join(u.HomeDir, path[2:])
	}
	return path
}

func setClipboard(str string) error {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	in, _ := cmd.StdinPipe()

	done := make(chan error)
	go func() {
		done <- cmd.Run()
	}()

	if _, err := in.Write([]byte(str)); err != nil {
		log.Fatalf("copy error: %s", err)
	}
	in.Close()

	return <-done
}
