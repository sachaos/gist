package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/google/go-github/github"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
)

var (
	configPath = os.Getenv("HOME")
)

const (
	configName = ".gist.config"
	configType = "json"
)

func main() {
	viper.SetConfigType(configType)
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")

	app := cli.NewApp()
	app.Name = "gist"
	app.Usage = "Simple GitHub Gist command"
	app.Version = "0.0.1"

	app.Commands = cli.Commands{
		cli.Command{
			Name:   "create",
			Usage:  "Create new snippet",
			Action: Create,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "public",
					Usage: "Create public Gist",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func Create(c *cli.Context) error {
	if err := viper.ReadInConfig(); err != nil {
		if err = viper.WriteConfigAs(configPath + "/" + configName + "." + configType); err != nil {
			return err
		}
	}

	token := viper.GetString("token")
	if token == "" {
		fmt.Println("This command needs GitHub Personal API token")
		fmt.Println("Generate token from here: https://github.com/settings/tokens")
		fmt.Printf("Input Token: ")
		fmt.Scan(&token)
		viper.Set("token", token)

		if err := viper.WriteConfig(); err != nil {
			return err
		}
	}

	if len(c.Args()) != 1 {
		return errors.New("Specify filename by argument")
	}

	filename := c.Args()[0]

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)

	public := c.Bool("public")
	gist, _, err := client.Gists.Create(context.Background(), &github.Gist{
		Public: github.Bool(public),
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(filename): github.GistFile{
				Content: github.String(string(bytes)),
			},
		},
	})
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", gist.GetHTMLURL())

	return nil
}
