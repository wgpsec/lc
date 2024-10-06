package cmd

import "github.com/projectdiscovery/gologger"

const banner = `
    __    _      __     ________                __
   / /   (_)____/ /_   / ____/ /___  __  ______/ /
  / /   / / ___/ __/  / /   / / __ \/ / / / __  / 
 / /___/ (__  ) /_   / /___/ / /_/ / /_/ / /_/ /  
/_____/_/____/\__/   \____/_/\____/\__,_/\__,_/
`

const version = "1.1.0"
const versionDate = "2024-10-6"

func showBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t      %s %s\n\t   github.com/wgpsec/lc\n\n", version, versionDate)
}
