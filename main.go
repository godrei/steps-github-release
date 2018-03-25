package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type config struct {
	APIToken      string `env:"api_token,required"`
	RepositoryURL string `env:"repository_url,required"`
	Tag           string `env:"tag,required"`
	Commit        string `env:"commit,required"`
	Name          string `env:"name,required"`
	Body          string `env:"body,required"`
	Draft         string `env:"draft,opt[yes,no]"`
}

// formats:
// https://hostname/owner/repository.git
// git@hostname:owner/repository.git
// ssh://git@hostname:port/owner/repository.git
func parseRepo(url string) (host string, owner string, name string) {
	url = strings.TrimSuffix(url, ".git")

	var repo string
	switch {
	case strings.HasPrefix(url, "https://"):
		url = strings.TrimPrefix(url, "https://")
		idx := strings.Index(url, "/")
		host, repo = url[:idx], url[idx+1:]
	case strings.HasPrefix(url, "git@"):
		url = url[strings.Index(url, "@")+1:]
		idx := strings.Index(url, ":")
		host, repo = url[:idx], url[idx+1:]
	case strings.HasPrefix(url, "ssh://"):
		url = url[strings.Index(url, "@")+1:]
		host = url[:strings.Index(url, ":")]
		repo = url[strings.Index(url, "/")+1:]
	}

	split := strings.Split(repo, "/")
	return host, split[0], split[1]
}

func main() {
	var cfg config
	if err := stepconf.Parse(&cfg); err != nil {
		log.Errorf("Error: %s\n", err)
		os.Exit(1)
	}
	stepconf.Print(cfg)

	ctx := context.Background()
	token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.APIToken})
	authClient := oauth2.NewClient(ctx, token)
	client := github.NewClient(authClient)

	draft := (cfg.Draft == "yes")
	release := &github.RepositoryRelease{
		TagName:         &cfg.Tag,
		TargetCommitish: &cfg.Commit,
		Name:            &cfg.Name,
		Body:            &cfg.Body,
		Draft:           &draft,
	}

	_, owner, name := parseRepo(cfg.RepositoryURL)
	newRelease, _, err := client.Repositories.CreateRelease(ctx, owner, name, release)
	if err != nil {
		log.Errorf("Failed to create release: %s", err)
		os.Exit(1)
	}

	printableRelease := newRelease.String()

	b, err := json.MarshalIndent(newRelease, "", " ")
	if err == nil {
		printableRelease = string(b)
	}

	fmt.Println()
	log.Infof("release created:")
	log.Printf(printableRelease)
}
