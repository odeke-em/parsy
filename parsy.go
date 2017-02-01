package parsy

import (
	"errors"
	"log"
	"os"
	"strings"
	"sync"
)

var debug = os.Getenv("PARSY_DEBUG") != ""

func debugPrintf(fmt_ string, args ...interface{}) {
	if debug {
		log.Printf(fmt_, args...)
	}
}

// See https://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html
func isOption(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}

func isHyphen(b byte) bool { return b == '-' }

// If they don't take arguments, multiple options may follow a hyphen delimiter
// in a single token. Thus "-abc" is equivalent to "-a -b -c".

func isalphaNumeric(r byte) bool {
	return (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z')
}

// OptionNames are single alphanumeric characters
func isOptionName(arg string) bool {
	return len(arg) == 1 && isalphaNumeric(arg[0])
}

// An option and its argument may or may not appear as separate tokens.
// In other words, the whitespace separating them is optional.
// Thus "-o foo" and "-ofoo" are equivalent.

// Options typically precede

type Parser struct {
	sync.RWMutex
	shortKeys map[string]*Argument
	longKeys  map[string]*Argument

	parsedValues map[string]interface{}

	cliArgs []string

	unparsedArgs []string
}

func NewParser(cliArgs ...string) (*Parser, error) {
	parser := &Parser{
		shortKeys:    make(map[string]*Argument),
		longKeys:     make(map[string]*Argument),
		parsedValues: make(map[string]interface{}),

		cliArgs: cliArgs,
	}

	return parser, nil
}

func (p *Parser) AddArgument(arg *Argument) error {
	if err := arg.pruneAndValidate(); err != nil {
		return err
	}

	// Now let's store the values in the short and long key maps
	arg.Lock()
	shortKey, longKey := arg.Short, arg.Long
	arg.Unlock()

	if err := p.storeShortArg(shortKey, arg); err != nil {
		return err
	}

	if err := p.storeLongArg(longKey, arg); err != nil {
		return err
	}

	return nil
}

var (
	errShortKeyAlreadyDefined = errors.New("short key already defined")
	errLongKeyAlreadyDefined  = errors.New("long key already defined")
)

func (p *Parser) storeShortArg(key string, arg *Argument) error {
	if key == "" { // noop, they haven't defined the shortKey
		return nil
	}

	p.Lock()
	defer p.Unlock()
	_, exists := p.shortKeys[key]

	if exists {
		return errShortKeyAlreadyDefined
	}

	p.shortKeys[key] = arg
	return nil
}

func (p *Parser) storeLongArg(key string, arg *Argument) error {
	if key == "" { // noop, they haven't defined the longKey
		return nil
	}

	p.Lock()
	defer p.Unlock()
	_, exists := p.longKeys[key]

	if exists {
		return errLongKeyAlreadyDefined
	}

	p.longKeys[key] = arg
	return nil
}

var errShortAndLongEmpty = errors.New("short and long options cannot both be empty")

func (a *Argument) pruneAndValidate() error {
	a.Lock()
	defer a.Unlock()

	a.Short = strings.TrimSpace(a.Short)
	a.Long = strings.TrimSpace(a.Long)
	if a.Short == "" && a.Long == "" {
		return errShortAndLongEmpty
	}

	return nil
}

// Add adds a long option command
func (p *Parser) AddCommand(short, long string, typ Type, defaultValue interface{}, help string) error {
	arg := &Argument{
		Type:    typ,
		Long:    long,
		Help:    help,
		Default: defaultValue,
	}
	return p.AddArgument(arg)

}

func (p *Parser) Add(long string, typ Type, defaultValue interface{}, help string) error {
	return p.AddCommand("", long, typ, defaultValue, help)
}

func firstNonEmptyString(args ...string) string {
	for _, arg := range args {
		if arg != "" {
			return arg
		}
	}
	return ""
}

func (a *Argument) Parse(s string) (interface{}, error) {
	a.RLock()
	parseFn, err := a.Type.Parser()
	a.RUnlock()

	if err != nil {
		return nil, err
	}

	argValue := firstNonEmptyString(s, a.ArgValue)
	if argValue == "" {
		return a.Default, nil
	}

	return parseFn(s)
}

type Options struct {
	sync.RWMutex
	mapping map[string]interface{}
}

type Argument struct {
	sync.RWMutex

	Type    Type
	Default interface{}
	Help    string
	Short   string
	Long    string
	Index   int

	ArgValue string
}

func newOptions() (*Options, error) {
	opts := &Options{mapping: make(map[string]interface{})}
	return opts, nil
}

var ErrNoSuchKey = errors.New("no such key")

func (p *Parser) Get(key string) interface{} {
	retr, _ := p.Value(key)
	return retr
}

func (p *Parser) Value(key string) (interface{}, error) {
	p.RLock()
	defer p.RUnlock()

	if retr, ok := p.parsedValues[key]; ok {
		return retr, nil
	}

	return nil, ErrNoSuchKey
}

// Args is the remnant of the cli arguments
// after parsing has been performed.
func (p *Parser) Args() []string {
	p.RLock()
	defer p.RUnlock()

	return p.unparsedArgs[:]
}

func (p *Parser) Parse() error {
	if err := p.groupAndCollectValues(); err != nil {
		return err
	}

	p.RLock()
	defer p.RUnlock()

	// First step is to group and parse long with short options
	alreadyParsedArgs := make(map[string]interface{})

	mapsToParseFrom := []map[string]*Argument{
		p.shortKeys,
		p.longKeys,
	}

	for _, srcMap := range mapsToParseFrom {
		for key, arg := range srcMap {
			debugPrintf("arg.ArgValue: %s key: %s", arg.ArgValue, key)
			if _, alreadyParsed := alreadyParsedArgs[key]; alreadyParsed {
				continue
			}
			value, err := arg.Parse(arg.ArgValue)
			if err != nil {
				return err
			}
			alreadyParsedArgs[key] = value
			p.parsedValues[key] = value
		}
	}

	return nil
}

func (p *Parser) groupAndCollectValues() error {
	// The goal is to scan the args, find tokens with - and -- and categorize them
	p.Lock()
	defer p.Unlock()

	ocliArgs := p.cliArgs[:]

	var cliArgs []string
	// The first step is to remove empty tokens
	for _, arg := range ocliArgs {
		if arg != "" {
			cliArgs = append(cliArgs, arg)
		}
	}

	debugPrintf("cliArgs: %#v", cliArgs)
	if len(cliArgs) < 1 {
		cliArgs = os.Args[1:]
	}

	// TODO: Handle short options
	// For every spot that we find short options
	// This is the style of arguments:
	//  -farchive.tar <==> -f archive.tar
	//  -cvz
	//  -c -v -z

	// For each long arg, search for the occurance and parse the next tokens
	for longKey, arg := range p.longKeys {
		index := scanArgs(longKey, cliArgs, tLong)
		debugPrintf("scanArgs: %#v index: %d\n", cliArgs, index)
		if index < 0 { // Not found
			continue
		}

		// We have 3 forms:
		// --prune=false
		// --prune
		// --prune true
		token := cliArgs[index]
		ith := strings.Index(token, longKey)
		rest := token[ith+len(longKey):]
		debugPrintf("%q: index: %d ith: %d rest: %s\n", longKey, index, ith, rest)
		if len(rest) == 0 {
			nextIndex := index + 1
			if nextArgIsAValue(cliArgs, nextIndex) {
				arg.ArgValue = cliArgs[nextIndex]
				debugPrintf("arg.ArgValue: %s\n", arg.ArgValue)
				// Now mutate cliArgs essentially popping
				// that option and its argument
				cliArgs = append(cliArgs[:index], cliArgs[nextIndex+1:]...)
			}
		} else if rest[0] == '=' {
			arg.ArgValue = rest[1:]
			cliArgs = append(cliArgs[:index], cliArgs[index+1:]...)
		} else {
		}
	}

	// Don't forget to set the arguments, which are the remnants of parsing
	p.unparsedArgs = cliArgs

	return nil
}

func nextArgIsAValue(args []string, i int) bool {
	if i >= len(args) {
		return false
	}

	nextArg := args[i]
	debugPrintf("args: %v i: %d nextArg: %s\n", args, i, nextArg)
	return len(nextArg) > 0 && !isHyphen(nextArg[0])
}

const (
	tShort = false
	tLong  = true
)

func scanArgs(key string, args []string, tEnum bool) int {
	checkerFn := isLongOption
	if tEnum == tShort {
		checkerFn = isShortOption
	}

	keyLen := len(key)
	for i, arg := range args {
		ok, argSkipIndex := checkerFn(arg)
		if !ok {
			continue
		}

		rest := arg[argSkipIndex:]
		// Now for the match
		if len(rest) < keyLen || rest[:keyLen] != key {
			continue
		}

		return i
	}

	return -1
}

func isShortOption(token string) (ok bool, skipIndex int) {
	if len(token) < 2 {
		return false, -1
	}

	if !isHyphen(token[0]) {
		return false, -1
	}

	// Next character has to be an alphanumeric character
	return isalphaNumeric(token[1]), 1
}

func isLongOption(token string) (ok bool, skipIndex int) {
	if len(token) < 3 {
		return false, -1
	}

	// The first two characters have to be "--"
	for i := 0; i < 2; i++ {
		if !isHyphen(token[i]) {
			return false, -1
		}
	}

	// Next character has to be an alphanumeric character
	return isalphaNumeric(token[2]), 2
}
