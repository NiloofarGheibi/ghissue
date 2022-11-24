package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
	"log"
	"os"
	"regexp"
	"strings"
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
	createIssues(ctx context.Context, owner string, repo string, issues []*github.Issue) (*github.Issue, error)
	createIssue(ctx context.Context, owner string, repo string, issues *github.Issue) (*github.Issue, error)
	commentIssue(ctx context.Context, sourceOwner string, sourceRepo string, targetOwner string, targetRepo string, sourceIssue *github.Issue, targetIssue *github.Issue) error
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

// 447
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

// Create Issue in the repo
func (z *zd) createIssues(ctx context.Context, owner string, repo string, issues []*github.Issue) (*github.Issue, error) {
	for _, issue := range issues {
		i := &github.IssueRequest{
			Title: issue.Title,
			Body:  issue.Body,
			State: issue.State,
		}
		newIssue, resp, err := z.gh.Issues.Create(ctx, owner, repo, i)
		if err != nil {
			return nil, err
		}
		fmt.Println(newIssue.User)
		fmt.Println(newIssue.Title)
		fmt.Println(resp.StatusCode)
	}
	return nil, nil
}

// Create Issue in the repo
func (z *zd) createIssue(ctx context.Context, owner string, repo string, issue *github.Issue) (*github.Issue, error) {
	i := &github.IssueRequest{
		Title: issue.Title,
		Body:  issue.Body,
		State: issue.State,
	}
	newIssue, resp, err := z.gh.Issues.Create(ctx, owner, repo, i)
	if err != nil {
		return nil, err
	}
	fmt.Println(newIssue.User)
	fmt.Println(newIssue.Title)
	fmt.Println(resp.StatusCode)

	return newIssue, nil
}

// Comment Issue in the repo
func (z *zd) commentIssue(ctx context.Context, sourceOwner string, sourceRepo string, targetOwner string, targetRepo string, sourceIssue *github.Issue, targetIssue *github.Issue) error {
	comments, _, err := z.gh.Issues.ListComments(ctx, sourceOwner, sourceRepo, sourceIssue.GetNumber(), nil)
	if err != nil {
		log.Fatal(err)
	}
	body := new(string)
	*body = "Comments and Issue are copied from the Original Ticket: " + sourceIssue.GetHTMLURL() + ". From now on Team Mantle is responsible for Tracking Data Quality issues. Please use this ticket/thread for further discussions cc @" + sourceIssue.GetUser().GetLogin()
	ref := github.IssueComment{Body: body}

	_, _, err = z.gh.Issues.CreateComment(ctx, targetOwner, targetRepo, targetIssue.GetNumber(), &ref)
	if err != nil {
		log.Fatal(err)
	}

	for _, comment := range comments {
		cm := github.IssueComment{
			ID:                comment.ID,
			Body:              comment.Body,
			User:              comment.GetUser(),
			Reactions:         comment.Reactions,
			CreatedAt:         comment.CreatedAt,
			UpdatedAt:         comment.UpdatedAt,
			AuthorAssociation: comment.AuthorAssociation,
			IssueURL:          nil,
		}
		_, _, err := z.gh.Issues.CreateComment(ctx, targetOwner, targetRepo, targetIssue.GetNumber(), &cm)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (z *zd) extractIssueToCSV(writer *csv.Writer, parser Parser, issues []*github.Issue) {
	for _, issue := range issues {
		body := issue.GetBody()
		title := issue.GetTitle()
		info := parser(body)
		info = append(info, title, issue.GetHTMLURL(), issue.GetCreatedAt().String(), issue.GetClosedAt().String())
		writer.Write(info)
	}
}

func parser(input string) []string {
	// Replace your parser here
	//messageRegex := regexp.MustCompile("(Message:)(.*)")
	userNameRegex := regexp.MustCompile("(User Name:)(.*)")
	bpidRegex := regexp.MustCompile("(ID:)(.*)")
	emailRegex := regexp.MustCompile("([a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\\.[a-zA-Z0-9_-]+)")

	//userNameRegex := regexp.MustCompile("(Merchant name:)(.*)")
	//bpidRegex := regexp.MustCompile("(ID:)(.*)")
	//emailRegex := regexp.MustCompile("([a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+\\.[a-zA-Z0-9_-]+)")
	//
	//username := userNameRegex.FindString(input)
	//bpid := bpidRegex.FindString(input)
	//merchantID := strings.ReplaceAll(bpid, "ID:", "")
	//user := strings.ReplaceAll(username, "Merchant name:", "")
	//email := emailRegex.FindString(input)
	//
	//return []string{strings.TrimSpace(user), strings.TrimSpace(merchantID), strings.TrimSpace(email)}

	email := emailRegex.FindString(input)
	bpid := bpidRegex.FindString(input)
	merchantID := strings.ReplaceAll(bpid, "ID:", "")
	user := strings.ReplaceAll(userNameRegex.FindString(input), "User Name:", "")

	return []string{strings.TrimSpace(user), strings.TrimSpace(merchantID), strings.TrimSpace(email)}

}

func main() {
	token := os.Getenv("ACCESS_TOKEN")
	//sourceOwner := os.Getenv("SOURCE_OWNER")
	//sourceRepo := os.Getenv("SOURCE_REPO")

	// page=1&per_page=100&q=mantle%2Fops+is%3Aissue+Carrier%2BMapping%2BChange%3Alabel
	// https://github.bus.zalan.do/mantle/ops/issues?q=is%3Aissue+label%3A%22Carrier+Mapping+Change%22+is%3Aclosed+
	//query := "[Carrier mapping Change]] is:issue state:closed"
	query := "[Invoice Delivery] is:issue state:closed"
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

	for p := 1; p <= (150/PerPage)+1; p++ {
		opts := &github.SearchOptions{ListOptions: github.ListOptions{Page: p, PerPage: PerPage}, TextMatch: true}
		issues, err := client.fetchIssues(query, opts)
		if err != nil {
			log.Println("during fetching pages some error happened:", err)
		}

		client.extractIssueToCSV(writer, parser, issues.Issues)
		//var data [][]string
		//for _, record := range issues.Issues {
		//	row := []string{record.GetTitle(), record.GetURL(), record.GetCreatedAt().UTC().String(), record.GetClosedAt().UTC().String(), fmt.Sprintf("%f", record.GetClosedAt().UTC().Sub(record.GetCreatedAt().UTC()).Hours()/24)}
		//	data = append(data, row)
		//}
		//writer.WriteAll(data)
	}

	////opts := &github.SearchOptions{ListOptions: github.ListOptions{Page: 1, PerPage: PerPage}}
	//sIssues, err := client.fetchIssues(sourceOwner+"/"+sourceRepo+" "+query, opts)
	//log.Fatal(err)

}

func issuesAge(issues []*github.Issue) (float64, []float64) {
	ages := make([]float64, len(issues))
	for i, issue := range issues {
		fmt.Println(issue.GetTitle())
		ages[i] = issue.ClosedAt.UTC().Sub(issue.CreatedAt.UTC()).Hours()
	}

	fmt.Println(ages)

	var avg float64 = 0

	for _, age := range ages {
		avg += age
	}

	return avg / float64(len(issues)), ages
}

//ticketNumbers := []int{794, 787, 788, 785, 786, 793, 826, 827, 828, 789}
//var sIssues []*github.Issue
//var tIssues []*github.Issue
//ctx := context.Background()
//
//for _, ticket := range ticketNumbers {
//	s, err := client.getIssue(context.Background(), sourceOwner, sourceRepo, ticket)
//	if err != nil {
//		log.Fatal(err)
//	}
//	sIssues = append(sIssues, s)
//	fmt.Println(s.GetHTMLURL())
//
//	t, _ := client.createIssue(context.Background(), targetOwner, targetRepo, s)
//	fmt.Println(t.GetHTMLURL())
//	tIssues = append(tIssues, t)
//
//	err = client.commentIssue(ctx, sourceOwner, sourceRepo, targetOwner, targetRepo, s, t)
//	if err != nil {
//		fmt.Println(err)
//	}

//for i := 464; i <= 464; i++ {
//	t, err := client.getIssue(context.Background(), targetOwner, targetRepo, i)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(t.GetHTMLURL())
//	tIssues = append(tIssues, t)
//}

//ctx := context.Background()
//for _, source := range sIssues {
//	for _, target := range tIssues {
//		if source.GetTitle() == target.GetTitle() {
//			err := client.commentIssue(ctx, sourceOwner, sourceRepo, targetOwner, targetRepo, source, target)
//			if err != nil {
//				fmt.Println(err)
//			}
//		}
//	}
//}

//client.createIssue(context.Background(), targetOwner, targetRepo, issues.Issues)
//client.extractIssueToCSV(writer, parser, issues.Issues)

//for p := 2; p < (issues.GetTotal()/PerPage)+1; p++ {
//	opts = &github.SearchOptions{ListOptions: github.ListOptions{Page: p, PerPage: PerPage}}
//	issues, err := client.fetchIssues(repo+" "+query, opts)
//	if err != nil {
//		log.Println("during fetching pages some error happened:", err)
//	}
//	client.extractIssueToCSV(writer, parser, issues.Issues)
//}
