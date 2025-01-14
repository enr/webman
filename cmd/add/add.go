package add

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/candrewlee14/webman/config"
	"github.com/candrewlee14/webman/multiline"
	"github.com/candrewlee14/webman/ui"
	"github.com/candrewlee14/webman/utils"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	doRefresh  bool
	switchFlag bool
)

// addCmd represents the add command
var AddCmd = &cobra.Command{
	Use:   "add [pkgs...]",
	Short: "install packages",
	Long: `
The "add" subcommand installs packages.`,
	Example: `webman add go
webman add go@18.0.0
webman add go zig rg
webman add go@18.0.0 zig@9.1.0 rg@13.0.0`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		defer os.RemoveAll(utils.WebmanTmpDir)
		// if local recipe flag is not set
		if utils.RecipeDirFlag == "" {
			// only refresh if not using local
			for _, pkgRepo := range cfg.PkgRepos {
				shouldRefresh, err := pkgRepo.ShouldRefreshRecipes(cfg.RefreshInterval)
				if err != nil {
					return err
				}
				if shouldRefresh || doRefresh {
					color.HiBlue("Refreshing package recipes for %q...", pkgRepo.Name)
					if err = pkgRepo.RefreshRecipes(); err != nil {
						color.Red("%v", err)
					}
				}
			}
		}
		pkgs := InstallAllPkgs(cfg.PkgRepos, args)
		for _, pkg := range pkgs {
			fmt.Print(pkg.InstallNotes())
		}
		if len(args) != len(pkgs) {
			return errors.New("Not all packages installed successfully")
		}
		color.Green("All %d packages are installed!", len(args))
		return nil
	},
}

func init() {
	AddCmd.Flags().BoolVar(&doRefresh, "refresh", false, "force refresh of package recipes")
	AddCmd.Flags().BoolVar(&switchFlag, "switch", false, "switch to use this new package version")
}

func cleanUpFailedInstall(pkg string, extractPath string) {
	os.RemoveAll(extractPath)
	pkgDir := filepath.Join(utils.WebmanPkgDir, pkg)
	dirs, err := os.ReadDir(pkgDir)
	if err == nil && len(dirs) == 0 {
		os.RemoveAll(pkgDir)
	}
}

func DownloadUrl(url string, filePath string, pkg string, ver string, argNum int, argCount int, ml *multiline.MultiLogger) bool {
	f, err := os.OpenFile(filePath,
		os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		return false
	}
	defer f.Close()

	r, err := http.Get(url)
	ml.Printf(argNum, "Downloading file at %s", url)
	if err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		return false
	}
	defer r.Body.Close()
	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		switch r.StatusCode {
		case 404, 403:
			ml.Printf(argNum, color.RedString("unable to find %s@%s on the web at %s", pkg, ver, url))
		default:
			ml.Printf(argNum, color.RedString("bad HTTP Response: %s", r.Status))
		}
		return false
	}
	ansiOn := ui.AreAnsiCodesEnabled()
	if !ansiOn {
		if _, err = io.Copy(f, r.Body); err != nil {
			ml.Printf(argNum, color.RedString("%v", err))
			return false
		}
		ml.Printf(argNum, `Completed downloading %s`, pkg)
		return true
	}
	colorOn := ui.AreAnsiCodesEnabled()
	saucer := "[green]▅[reset]"
	saucerHead := "[green]▅[reset]"
	saucerPadding := "[light_gray]▅[reset]"
	barStart := ""
	barEnd := ""
	barDesc := fmt.Sprintf("[cyan][%d/%d][reset] Downloading [cyan]"+pkg+"[reset] file...", argNum+1, argCount)
	if !colorOn {
		saucer = "="
		saucerHead = ">"
		saucerPadding = " "
		barDesc = fmt.Sprintf("[%d/%d] Downloading "+pkg+" file...", argNum+1, argCount)
		barStart = "["
		barEnd = "]"
	}
	bar := progressbar.NewOptions64(r.ContentLength,
		progressbar.OptionEnableColorCodes(colorOn),
		progressbar.OptionUseANSICodes(ansiOn),
		progressbar.OptionSetWriter(ioutil.Discard),
		progressbar.OptionShowBytes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription(barDesc),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        saucer,
			SaucerHead:    saucerHead,
			SaucerPadding: saucerPadding,
			BarStart:      barStart,
			BarEnd:        barEnd,
		}),
	)
	go func() {
		for !bar.IsFinished() {
			barStr := bar.String()
			ml.Printf(argNum, "%s", barStr)
			time.Sleep(100 * time.Millisecond)
		}
	}()
	if _, err = io.Copy(io.MultiWriter(f, bar), r.Body); err != nil {
		ml.Printf(argNum, color.RedString("%v", err))
		return false
	}
	return true
}
