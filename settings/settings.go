// Package settings contains default- and user-settings for nyfikenc/d.
package settings

import (
	"encoding/gob"
	"log"
	"os"
	"time"

	"github.com/mewkiz/pkg/errutil"
	"github.com/mewkiz/pkg/osutil"
)

// Queries sent from the client to the daemon.
const (
	QueryClearAll     = "clear all!"
	QueryForceRecheck = "recheck!"
	QueryUpdates      = "updates?"
)

// Default values.
const (
	// Default interval between updates unless overwritten in config file.
	DefaultInterval = 1 * time.Minute

	// Duration until a timeout is issued.
	TimeoutDuration = 10 * time.Second

	// Default permissions to create files: user read and write permissions.
	DefaultFilePerms   = os.FileMode(0600)
	DefaultFolderPerms = os.FileMode(0755)

	// Default newline character.
	Newline = "\n"

	// Default port number for nyfikenc/d connection.
	DefaultPortNum = ":5239"
)

// Paths to nyfiken files.
var (
	NyfikenRoot    string
	ConfigPath     string
	PagesPath      string
	CacheRoot      string
	ReadRoot       string
	UpdatesPath    string
	DebugRoot      string
	DebugCacheRoot string
	DebugReadRoot  string
)

var (
	// Updates is a map of all pages which have been updated.
	Updates map[string]bool

	// Settings which will be used unless overwritten by site-specific settings.
	Global = Prog{
		Interval:  DefaultInterval,
		FilePerms: DefaultFilePerms,
		PortNum:   DefaultPortNum,
	}

	// When Verbose is true, enable verbose output.
	Verbose bool
)

// Page is a collection of specialized settings used to eliminate
// false-positives. Page settings override program global settings.
type Page struct {
	Interval   time.Duration     // Duration of time to wait between scrapes.
	Threshold  float64           // Percentage of accepted deviation from last scrape.
	RecvMail   string            // Mail address to send a notification when a page has been updated.
	Regexp     string            // Regular expression to further specify what to select.
	Negexp     string            // Everything that matches this regular expression will be removed.
	StripFuncs []string          // Strip functions to further specify what to select.
	Header     map[string]string // HTTP headers to request targeted site with.
	Selection  string            // CSS selector string to specify what to select.
}

// Prog is the program global settings which regards all pages unless
// overwritten with page specific settings.
type Prog struct {
	Interval   time.Duration // Duration of time to wait between scrapes.
	RecvMail   string        // Mail address to send a notification when a page has been updated.
	StripFuncs []string      // Strip functions to further specify what to select.
	FilePerms  os.FileMode   // Permissions to create files with.
	PortNum    string        // On which port should the nyfikenc/d communication take place.
	Browser    string        // The path to the browser to open updates in.

	// Information about the mail address to send updates.
	SenderMail struct {
		Address    string // Mail address of the sending mail.
		Password   string // Password to that mail address.
		AuthServer string // Authorization server to the mail address.
		OutServer  string // Outgoing server to the mail address.
	}
}

// Error wrapper.
func init() {
	err := initialize()
	if err != nil {
		log.Fatalln(errutil.Err(err))
	}
}

func initialize() (err error) {
	Updates = make(map[string]bool)

	// Will set nyfiken root differently depending on operating system.
	setNyfikenRoot()
	ConfigPath = NyfikenRoot + "/config.ini"
	PagesPath = NyfikenRoot + "/pages.ini"
	UpdatesPath = NyfikenRoot + "/updates.gob"

	CacheRoot = NyfikenRoot + "/cache/"
	ReadRoot = NyfikenRoot + "/read/"
	DebugRoot = NyfikenRoot + "/debug/"
	DebugCacheRoot = NyfikenRoot + "/debug/cache/"
	DebugReadRoot = NyfikenRoot + "/debug/read/"

	// Load uncleared updates from last execution.
	err = LoadUpdates()
	if err != nil {
		return errutil.Err(err)
	}

	// Create a nyfiken config folder if it doesn't exist.
	if !osutil.Exists(NyfikenRoot) {
		err := os.Mkdir(NyfikenRoot, DefaultFolderPerms)
		if err != nil {
			return errutil.Err(err)
		}
	}

	if !osutil.Exists(CacheRoot) {
		err := os.Mkdir(CacheRoot, DefaultFolderPerms)
		if err != nil {
			return errutil.Err(err)
		}
	}

	if !osutil.Exists(ReadRoot) {
		err := os.Mkdir(ReadRoot, DefaultFolderPerms)
		if err != nil {
			return errutil.Err(err)
		}
	}

	if !osutil.Exists(DebugRoot) {
		err := os.Mkdir(DebugRoot, DefaultFolderPerms)
		if err != nil {
			return errutil.Err(err)
		}
	}

	if !osutil.Exists(DebugCacheRoot) {
		err := os.Mkdir(DebugCacheRoot, DefaultFolderPerms)
		if err != nil {
			return errutil.Err(err)
		}
	}

	if !osutil.Exists(DebugReadRoot) {
		err := os.Mkdir(DebugReadRoot, DefaultFolderPerms)
		if err != nil {
			return errutil.Err(err)
		}
	}

	return nil
}

// SaveUpdates saves uncleared updates for next execution.
func SaveUpdates() (err error) {
	f, err := os.Create(UpdatesPath)
	if err != nil {
		return errutil.Err(err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)

	err = enc.Encode(&Updates)
	if err != nil {
		return errutil.Err(err)
	}
	return nil
}

// LoadUpdates retrieves saved updates from last execution.
func LoadUpdates() (err error) {
	f, err := os.Open(UpdatesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errutil.Err(err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	err = dec.Decode(&Updates)
	if err != nil {
		return errutil.Err(err)
	}
	return nil
}
