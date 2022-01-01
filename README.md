# github-tools
This repo contains a github issue parser, that is useful for Enterprise Github accounts. Sometimes is needed to parse the content of the issue for some data extraction or statistics purposes. 

## Prerequisite
You need to have **Golang** installed on your machine.

## How to
1. Generate a **Personal Access Token** on your [Github](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).
2. Set the following environment variables to be able to fetch the issues in your repository that you are looking for: 
```
ACCESS_TOKEN=<The one you generated on Github>
REPO=<ORGANIZATION/REPO>
QUERY=<QUERY YOU ARE LOOKING FOR>
``` 
Example: 
```
ACCESS_TOKEN=verysecrettoken
REPO=my-cool-org/repo-name
QUERY=[Feedback] in:title
```
To know more about the query options on the Github APIs, checkout their [docs](https://docs.github.com/en/search-github/searching-on-github/searching-issues-and-pull-requests#search-by-the-title-body-or-comments). 

3. Run `go build`
4. Run `./ghissue`

### Parser
The `extractIssueToCSV` method, accepts `parser` as a function. Feel free to adapt the Parser to your needs and templates you use in your Github Issues. 

### Output
You can get your result as a `CSV` file.
