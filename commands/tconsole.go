package commands

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"github.com/peterh/liner"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"github.com/mattn/go-colorable"
	//"github.com/peterh/liner"
)

var (
	onlyWhitespace = regexp.MustCompile(`^\s*$`)
	exit           = regexp.MustCompile(`^\s*exit\s*;*\s*$`)
)

// HistoryFile is the file within the data directory to store input scrollback.
const HistoryFile = "history"

// DefaultPrompt is the default prompt line prefix to use for user input querying.
const DefaultPrompt = "> "

// Config is the collection of configurations to fine tune the behavior of the
// JavaScript console.
type Config struct {
	DataDir  string       // Data directory to store the console history at
	DocRoot  string       // Filesystem path from where to load JavaScript files from
	Prompt   string       // Input prompt prefix string (defaults to DefaultPrompt)
	Prompter UserPrompter // Input prompter to allow interactive user feedback (defaults to TerminalPrompter)
	Printer  io.Writer    // Output writer to serialize any display strings to (defaults to os.Stdout)
	Preload  []string     // Absolute paths to JavaScript files to preload
}

// Console is a JavaScript interpreted runtime environment. It is a fully fledged
// JavaScript console attached to a running node via an external or in-process RPC
// client.
type Console struct {
	prompt   string       // Input prompt prefix string
	prompter UserPrompter // Input prompter to allow interactive user feedback
	histPath string       // Absolute path to the console scrollback history
	history  []string     // Scroll history maintained by the console
	printer  io.Writer    // Output writer to serialize any display strings to
}

// New initializes a JavaScript interpreted runtime environment and sets defaults
// with the config struct.
func New(config Config) (*Console, error) {
	// Handle unset config values gracefully
	if config.Prompter == nil {
		config.Prompter = Stdin
	}
	if config.Prompt == "" {
		config.Prompt = DefaultPrompt
	}
	if config.Printer == nil {
		config.Printer = colorable.NewColorableStdout()
	}
	// Initialize the console and return
	console := &Console{
		prompt:   config.Prompt,
		prompter: config.Prompter,
		printer:  config.Printer,
		histPath: filepath.Join(config.DataDir, HistoryFile),
	}
	//if err := os.MkdirAll(config.DataDir, 0700); err != nil {
	//	return nil, err
	//}
	if err := console.init(config.Preload); err != nil {
		return nil, err
	}
	return console, nil
}

// init retrieves the available APIs from the remote RPC provider and initializes
// the console's JavaScript namespaces based on the exposed modules.
func (c *Console) init(preload []string) error {
	// Configure the console's input prompter for scrollback and tab completion
	if c.prompter != nil {
		if content, err := ioutil.ReadFile(c.histPath); err != nil {
			c.prompter.SetHistory(nil)
		} else {
			c.history = strings.Split(string(content), "\n")
			c.prompter.SetHistory(c.history)
		}
		c.prompter.SetWordCompleter(c.AutoCompleteInput)
	}
	return nil
}

func (c *Console) clearHistory() {
	c.history = nil
	c.prompter.ClearHistory()
	if err := os.Remove(c.histPath); err != nil {
		fmt.Fprintln(c.printer, "can't delete history file:", err)
	} else {
		fmt.Fprintln(c.printer, "history file deleted")
	}
}

// AutoCompleteInput is a pre-assembled word completer to be used by the user
// input prompter to provide hints to the user about the methods available.
func (c *Console) AutoCompleteInput(line string, pos int) (string, []string, string) {
	// No completions can be provided for empty inputs
	if len(line) == 0 || pos == 0 {
		return "", nil, ""
	}
	// Chunck data to relevant part for autocompletion
	// E.g. in case of nested lines eth.getBalance(eth.coinb<tab><tab>
	start := pos - 1
	for ; start > 0; start-- {
		// Skip all methods and namespaces (i.e. including the dot)
		if line[start] == '.' || (line[start] >= 'a' && line[start] <= 'z') || (line[start] >= 'A' && line[start] <= 'Z') {
			continue
		}
		// Handle web3 in a special way (i.e. other numbers aren't auto completed)
		if start >= 3 && line[start-3:start] == "web3" {
			start -= 3
			continue
		}
		// We've hit an unexpected character, autocomplete form here
		start++
		break
	}
	s := []string{}
	return line[:start], s, line[pos:]
}

// Interactive starts an interactive user session, where input is propted from
// the configured user prompter.
func (c *Console) Interactive() {
	var (
		prompt    = c.prompt          // Current prompt line (used for multi-line inputs)
		indents   = 0                 // Current number of input indents (used for multi-line inputs)
		input     = ""                // Current user input
		scheduler = make(chan string) // Channel to send the next prompt on and receive the input
	)
	// Start a goroutine to listen for prompt requests and send back inputs
	go func() {
		for {
			// Read the next user input
			line, err := c.prompter.PromptInput(<-scheduler)
			if err != nil {
				// In case of an error, either clear the prompt or fail
				if err == liner.ErrPromptAborted { // ctrl-C
					prompt, indents, input = c.prompt, 0, ""
					scheduler <- ""
					continue
				}
				close(scheduler)
				return
			}
			// User input retrieved, send for interpretation and loop
			scheduler <- line
		}
	}()
	// Monitor Ctrl-C too in case the input is empty and we need to bail
	abort := make(chan os.Signal, 1)
	signal.Notify(abort, syscall.SIGINT, syscall.SIGTERM)

	// Start sending prompts to the user and reading back inputs
	for {
		// Send the next prompt, triggering an input read and process the result
		scheduler <- prompt
		select {
		case <-abort:
			// User forcefully quite the console
			fmt.Fprintln(c.printer, "caught interrupt, exiting")
			return

		case line, ok := <-scheduler:
			// User input was returned by the prompter, handle special cases
			if !ok || (indents <= 0 && exit.MatchString(line)) {
				return
			}
			if onlyWhitespace.MatchString(line) {
				continue
			}
			// Append the line to the input and check for multi-line interpretation
			input += line + "\n"

			indents = countIndents(input)
			if indents <= 0 {
				prompt = c.prompt
			} else {
				prompt = strings.Repeat(".", indents*3) + " "
			}
			// If all the needed lines are present, save the command and run
			if indents <= 0 {
				if len(input) > 0 && input[0] != ' ' {
					if command := strings.TrimSpace(input); len(c.history) == 0 || command != c.history[len(c.history)-1] {
						c.history = append(c.history, command)
						if c.prompter != nil {
							c.prompter.AppendHistory(command)
						}
					}
				}
				var cmd, arg1, arg2, arg3 string
				is := strings.Fields(input)
				for i, str := range is {
					if i == 0 {
						cmd = str
					} else if i == 1 {
						arg1 = str
					} else if i == 2 {
						arg2 = str
					} else if i == 3 {
						arg3 = str
					}
				}
				if cmd == "exit" {
					break
				}
				if cmd == "re" {
					continue
				}
				switch cmd {
				case "createNewAccount":
					createNewAccount(arg1)
					break
				case "getBalance":
					if arg1 == "" {
						fmt.Println("Please enter your address.")
						break
					}
					company := "i"
					detail := "false"
					b, err := getBalance(arg1)
					if err != nil {
						fmt.Println(err.Error())
						return
					}
					if arg2 != "" && arg2 != "i" {
						company = "f"
					}
					if arg3 != "" && arg3 != "false" {
						detail = "true"
					}
					if company == "i" {
						if detail == "true" {
							fmt.Printf("unspend:%s\n", b.UnspendAmount.String())
							fmt.Printf("unconfirmed:%s\n", b.ConfirmAmount.String())
							fmt.Printf("totalamount:%s\n", b.TotalAmount.String())
							fmt.Printf("spendamount:%s\n", b.SpendAmount.String())
						} else {
							fmt.Printf("%s\n", b.UnspendAmount.String())
						}
					} else {
						if detail == "true" {
							fmt.Printf("unspend:%f\n", b.UnspendAmount.ToCoin())
							fmt.Printf("unconfirmed:%f\n", b.ConfirmAmount.ToCoin())
							fmt.Printf("totalamount:%f\n", b.TotalAmount.ToCoin())
							fmt.Printf("spendamount:%f\n", b.SpendAmount.ToCoin())
						} else {
							fmt.Printf("%f\n", b.UnspendAmount.ToCoin())
						}
					}
					break
				//case "listAccountsBalance":
				//	listAccountsBalance(Default_minconf)
				//	break
				case "getListTxByAddr":
					if arg1 == "" {
						fmt.Println("getListTxByAddr err :Please enter your address.")
						break
					}
					filter := wallet.FilterAll
					if arg2 == "in" {
						filter = wallet.FilterIn
					} else if arg2 == "out" {
						filter = wallet.FilterOut
					}

					getListTxByAddr(arg1, filter, wallet.PageUseDefault, wallet.PageDefaultSize)
					break
				case "getBillsByAddr":
					if arg1 == "" {
						fmt.Println("getBillsByAddr err :Please enter your address.")
						break
					}
					filter := wallet.FilterAll
					if arg2 == "in" {
						filter = wallet.FilterIn
					} else if arg2 == "out" {
						filter = wallet.FilterOut
					}

					getBillsByAddr(arg1, filter, wallet.PageUseDefault, wallet.PageDefaultSize)
					break
				case "getNewAddress":
					if arg1 == "" {
						fmt.Println("getNewAddress err :Please enter your account.")
						break
					}
					getNewAddress(arg1)
					break
				case "getAddressesByAccount":
					if arg1 == "" {
						fmt.Println("getAddressesByAccount err :Please enter your account.")
						break
					}
					getAddressesByAccount(arg1)
					break
				case "getAccountByAddress":
					if arg1 == "" {
						fmt.Println("getAccountByAddress err :Please enter your address.")
						break
					}
					getAccountByAddress(arg1)
					break
				case "importPrivKey":
					if arg1 == "" {
						fmt.Println("importPrivKey err :Please enter your priKey.")
						break
					}
					importPrivKey(arg1)
					break
				case "importWifPrivKey":
					if arg1 == "" {
						fmt.Println("importwifPriKey err :Please enter your wif priKey.")
						break
					}
					importWifPrivKey(arg1)
					break
				case "dumpPrivKey":
					if arg1 == "" {
						fmt.Println("dumpPrivKey err :Please enter your address.")
						break
					}
					dumpPrivKey(arg1)
					break
				case "getAccountAndAddress":
					getAccountAndAddress()
					break
				case "sendToAddress":
					if arg1 == "" {
						fmt.Println("getAccountAndAddress err : Please enter the receipt address.")
						break
					}
					if arg2 == "" {
						fmt.Println("getAccountAndAddress err : Please enter the amount of transfer.")
						break
					}
					f32, err := strconv.ParseFloat(arg2, 32)
					if err != nil {
						fmt.Println("getAccountAndAddress err :", err.Error())
						break
					}
					sendToAddress(arg1, float64(f32))
					break
				case "updateblock":
					updateblock(0)
					break
				case "syncheight":
					syncheight()
					break
				case "unlock":
					if arg1 == "" {
						fmt.Println("unlock err : Please enter the pri password.")
						break
					}
					unlock(arg1)
					break
				case "help":
					printHelp()
					break
				default:
					fmt.Printf("Wrong command %s\n ", cmd)
					break
				}
				input = ""
			}
		}
	}
}

// countIndents returns the number of identations for the given input.
// In case of invalid input such as var a = } the result can be negative.
func countIndents(input string) int {
	var (
		indents     = 0
		inString    = false
		strOpenChar = ' '   // keep track of the string open char to allow var str = "I'm ....";
		charEscaped = false // keep track if the previous char was the '\' char, allow var str = "abc\"def";
	)

	for _, c := range input {
		switch c {
		case '\\':
			// indicate next char as escaped when in string and previous char isn't escaping this backslash
			if !charEscaped && inString {
				charEscaped = true
			}
		case '\'', '"':
			if inString && !charEscaped && strOpenChar == c { // end string
				inString = false
			} else if !inString && !charEscaped { // begin string
				inString = true
				strOpenChar = c
			}
			charEscaped = false
		case '{', '(':
			if !inString { // ignore brackets when in string, allow var str = "a{"; without indenting
				indents++
			}
			charEscaped = false
		case '}', ')':
			if !inString {
				indents--
			}
			charEscaped = false
		default:
			charEscaped = false
		}
	}

	return indents
}
