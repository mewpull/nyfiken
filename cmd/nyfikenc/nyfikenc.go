// A client program to check and handle updates from nyfiken daemon.
package main

import "os/exec"
import "os"
import "encoding/gob"
import "log"
import "flag"
import "fmt"
import "net"

import "github.com/karlek/nyfiken/settings"
import "github.com/mewkiz/pkg/bufioutil"

// command-line flags
var flagRecheck bool
var flagClearAll bool
var flagReadAll bool

func init() {
	flag.BoolVar(&flagRecheck, "f", false, "forces a recheck.")
	flag.BoolVar(&flagReadAll, "r", false, "read all updated pages in your browser.")
	flag.BoolVar(&flagClearAll, "c", false, "will clear list of updated sites.")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintln(os.Stderr, "nyfikenc [OPTION]")
	fmt.Fprintln(os.Stderr)
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
}

// Error wrapper.
func main() {
	flag.Parse()
	err := nyfikenc()
	if err != nil {
		log.Fatalln("Nyfikenc:", err)
	}
}

func nyfikenc() (err error) {

	// Connect to nyfikend.
	conn, err := net.Dial("tcp", "localhost"+settings.Global.PortNum)
	if err != nil {
		return err
	}

	// Ask for updates.
	bw := bufioutil.NewWriter(conn)
	_, err = bw.WriteLine(settings.QueryUpdates)
	if err != nil {
		return err
	}

	// Will read from network.
	dec := gob.NewDecoder(conn)

	// Decode (receive) the value.
	var ups map[settings.Update]bool
	err = dec.Decode(&ups)
	if err != nil {
		return err
	}

	// Command-line flag check
	if flagRecheck ||
		flagClearAll ||
		flagReadAll {
		if flagRecheck {
			return force(&bw)
		}
		if flagClearAll {
			return clearAll(&bw)
		}
		if flagReadAll {
			br := bufioutil.NewReader(conn)
			return readAll(&bw, &br, ups)
		}
	} else {
		// If no updates where found -> apologize.
		lenUps := len(ups)
		if lenUps == 0 {
			fmt.Println("Sorry, no updates :(")
			return err
		}
	}
	if err != nil {
		return err
	}

	for up, _ := range ups {
		fmt.Printf("%s\n", up.ReqUrl)
	}

	return nil
}

// Opens all links with browser.
func readAll(bw *bufioutil.Writer, br *bufioutil.Reader, ups map[settings.Update]bool) (err error) {
	// Ask nyfikend for browser path
	_, err = bw.WriteLine(settings.QueryAskForBrowser)
	if err != nil {
		return err
	}

	// Reads response from nyfikend
	browser, err := br.ReadLine()
	if browser == "" {
		fmt.Println("No browser path set in:", settings.ConfigPath)
		return nil
	}

	// If no updates was found, ask for forgiveness.
	if len(ups) == 0 {
		fmt.Println("Sorry, no updates :(")
		return nil
	}

	// Loop through all updates and open them with the browser
	for up, _ := range ups {
		cmd := exec.Command(browser, up.ReqUrl)
		err := cmd.Start()
		if err != nil {
			return err
		}
		err = cmd.Wait()
		if err != nil {
			return err
		}
	}

	fmt.Println("Opening all updates with:", browser)
	return nil
}

// Removes all updates.
func clearAll(bw *bufioutil.Writer) (err error) {
	// Send nyfikend a query to clear updates.
	_, err = bw.WriteLine(settings.QueryClearAll)
	if err != nil {
		return err
	}

	fmt.Println("Updates list has been cleared!")
	return nil
}

// Forces nyfikend to check all pages immediately.
func force(bw *bufioutil.Writer) (err error) {
	// Send nyfikend a query to force a recheck.
	_, err = bw.WriteLine(settings.QueryForceRecheck)
	if err != nil {
		return err
	}

	fmt.Println("All pages will now be checked!")
	return nil
}
