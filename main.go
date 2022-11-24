package main

import (
	"context"
	"encoding/csv"
	"golang.org/x/oauth2"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v40/github"
)

const (
	GitHubEnterpriseURL = "https://github.bus.zalan.do"
	PerPage             = 100
)

type Parser func(string) []string

type ZD interface {
	fetchIssues(query string, opts *github.SearchOptions) (*github.IssuesSearchResult, error)
	fetchRepo(ctx context.Context, owner string, repo string) (*github.Repository, error)
	extractIssueToCSV(writer *csv.Writer, parser Parser, issues []*github.Issue)
	getIssue(ctx context.Context, owner string, repo string, number int) (*github.Issue, error)
}

type zd struct {
	gh *github.Client
}

func newZDGithub(client *github.Client) ZD {
	return &zd{gh: client}
}

func newStaticAuthClient(token string, url string) (*github.Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client, err := github.NewEnterpriseClient(url, url, tc)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Fetch and lists all the public topics associated with the specified GitHub topic
func (z *zd) getIssue(ctx context.Context, owner string, repo string, number int) (*github.Issue, error) {
	issue, _, err := z.gh.Issues.Get(ctx, owner, repo, number)
	return issue, err
}

// Fetch and lists all the public topics associated with the specified GitHub topic
func (z *zd) fetchIssues(query string, opts *github.SearchOptions) (*github.IssuesSearchResult, error) {
	issues, _, err := z.gh.Search.Issues(context.Background(), query, opts)

	return issues, err
}

// Fetch and lists all the public topics associated with the specified GitHub topic
func (z *zd) fetchRepo(ctx context.Context, owner string, name string) (*github.Repository, error) {
	repo, _, err := z.gh.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (z *zd) extractIssueToCSV(writer *csv.Writer, parser Parser, issues []*github.Issue) {
	for _, issue := range issues {
		body := issue.GetBody()
		info := parser(body)
		info = append(info, issue.GetHTMLURL(), issue.GetCreatedAt().String(), issue.GetClosedAt().String())
		writer.Write(info)
	}
}

func parser(input string) []string {
	// Replace your parser here
	messageRegex := regexp.MustCompile("(Message:)(.*)")
	emailRegex := regexp.MustCompile("([a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\\.[a-zA-Z0-9_-]+)")

	msg := messageRegex.FindString(input)
	message := strings.ReplaceAll(msg, "Message:", "")
	email := emailRegex.FindString(input)

	return []string{message, email}
}

func main() {
	token := os.Getenv("ACCESS_TOKEN")
	repo := os.Getenv("REPO")
	query := os.Getenv("QUERY")

	authClient, err := newStaticAuthClient(token, GitHubEnterpriseURL)
	if err != nil {
		log.Fatal(err)
	}
	client := newZDGithub(authClient)

	file, err := os.Create("result.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	opts := &github.SearchOptions{ListOptions: github.ListOptions{Page: 1, PerPage: PerPage}}
	issues, err := client.fetchIssues(repo+" "+query, opts)
	if err != nil {
		log.Fatal(err)
	}

	client.extractIssueToCSV(writer, parser, issues.Issues)

	for p := 2; p < (issues.GetTotal()/PerPage)+1; p++ {
		opts = &github.SearchOptions{ListOptions: github.ListOptions{Page: p, PerPage: PerPage}}
		issues, err := client.fetchIssues(repo+" "+query, opts)
		if err != nil {
			log.Println("during fetching pages some error happened:", err)
		}
		client.extractIssueToCSV(writer, parser, issues.Issues)
	}
}
