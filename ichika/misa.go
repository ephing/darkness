package ichika

import (
	"fmt"
	"os"
	"strings"

	"github.com/thecsw/darkness/emilia/alpha"
	"github.com/thecsw/darkness/emilia/alpha/roxy"
	"github.com/thecsw/darkness/emilia/puck"
	"github.com/thecsw/darkness/ichika/misa"
)

// MisaCommandFunc will support many different tools that darkness can support,
// such as creating gallery previews, etc. WIP.
func MisaCommandFunc() {
	misaCmd := darknessFlagset(misaCommand)

	buildGalleryPreviews := misaCmd.Bool("gallery-previews", false, "build gallery previews")
	removeGalleryPreviews := misaCmd.Bool("no-gallery-previews", false, "delete gallery previews")
	addHolosceneTitles := misaCmd.Bool("holoscene-titles", false, "add holoscene titles")
	rss := misaCmd.String("rss", "", "generate an rss file")
	rssDirectories := misaCmd.String("rss-dirs", "", "look up specific dirs")
	dryRun := misaCmd.Bool("dry-run", false, "skip writing files (but do the reading)")
	pluginName := ""
	misaCmd.StringVar(&pluginName, "plugin", "", "execute a misa plugin")

	options := getAlphaOptions(misaCmd)
	options.Dev = true

	puck.Logger.SetPrefix("Misa 🍎 ")

	if len(*rss) > 0 {
		options.Dev = false
	}
	conf := alpha.BuildConfig(options)

	if *buildGalleryPreviews {
		misa.BuildGalleryFiles(conf, *dryRun)
		os.Exit(0)
	}
	if *removeGalleryPreviews {
		misa.RemoveGalleryFiles(conf, *dryRun)
		os.Exit(0)
	}
	if *addHolosceneTitles {
		misa.UpdateHoloceneTitles(conf, *dryRun)
		os.Exit(0)
	}
	if len(*rss) > 0 {
		misa.GenerateRssFeed(conf, *rss, strings.Split(*rssDirectories, ","), *dryRun)
		os.Exit(0)
	}
	if pluginName != "" {
		if plgn, ok := conf.Runtime.PluginConfigs[pluginName]; ok {
			err := plgn.Do.(roxy.MisaDo)(plgn.Data, conf, *dryRun)
			if err != nil {
				puck.Logger.Error(err)
			}
			os.Exit(0)
		}
		puck.Logger.Fatalf("Plugin \"%s\" does not exist", pluginName)
		os.Exit(0)
	}

	if misaCmd.NFlag() == 0 {
		fmt.Println("I don't know what you want me to do, see -help")
	}
}
