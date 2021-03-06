package main

import (
	"context"
	"os"
	"time"

	"fmt"
	"github.com/google/go-github/github"
)

// Git stores the information retrieved from github
type Git struct {
	Projects []*github.Project           // list of organization projects
	Repos    map[string]GitRepo          // map of repo name to repo object
	Releases []*github.RepositoryRelease // list of releases for revel/revel
}

// GitRepo stores the repo information from github
type GitRepo struct {
	Milestones []*github.Milestone
	Labels     []*github.Label
}

func loadGithub() error {
	username := "shawncatz"
	password := os.Getenv("GITHUB_TOKEN")

	tp := github.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	client = github.NewClient(tp.Client())
	ctx = context.Background()

	git = &Git{Repos: make(map[string]GitRepo)}

	projects, _, err := client.Organizations.ListProjects(ctx, ORG, &github.ProjectListOptions{})
	if err != nil {
		return err
	}

	git.Projects = projects

	for _, r := range config.Repos {
		list, _, err := client.Issues.ListMilestones(ctx, ORG, r, &github.MilestoneListOptions{State: "all"})
		if err != nil {
			return err
		}

		gr := GitRepo{}

		gr.Milestones = list

		labels, err := getLabels(r)
		if err != nil {
			return err
		}

		gr.Labels = labels

		git.Repos[r] = gr
	}

	git.Releases, _, err = client.Repositories.ListReleases(ctx, ORG, "revel", &github.ListOptions{})
	if err != nil {
		return err
	}

	return nil
}

func loadMilestone(repo, milestone string) ([]*github.Issue, error) {
	id := ""
	for _, m := range git.Repos[repo].Milestones {
		if *m.Title == milestone {
			id = fmt.Sprintf("%d", *m.Number)
			break
		}
	}

	if id == "" {
		return nil, fmt.Errorf("could not find number for milestone %s in repo %s", milestone, repo)
	}

	issues, _, err := client.Issues.ListByRepo(ctx, ORG, repo, &github.IssueListByRepoOptions{Milestone: id, State: "closed"})
	if err != nil {
		return nil, err
	}
	return issues, nil
}

// getLabels returns a combined slice of labels from paged responses
// some repos might have more labels than fit in a single page
func getLabels(repo string) ([]*github.Label, error) {
	out := []*github.Label{}
	page := 1
	lastpage := 0

	for ok := true; ok; ok = (page == lastpage) {
		labels, resp, err := client.Issues.ListLabels(ctx, ORG, repo, &github.ListOptions{Page: page})
		if err != nil {
			return nil, err
		}

		out = append(out, labels...) // ... breaks the slice into variadic parameters
		lastpage = resp.LastPage
		page++
	}

	return out, nil
}

// Project finds a project with the given name
//func (g *Git) Project(name string) *github.Project {
//	for _, p := range g.Projects {
//		if *p.Name == name {
//			return p
//		}
//	}
//
//	return nil
//}

// Milestone finds a milestone with the given repo and name
func (g *Git) Milestone(repo, name string) *github.Milestone {
	for _, m := range g.Repos[repo].Milestones {
		if *m.Title == name {
			return m
		}
	}

	return nil
}

// Label finds a label with the given repo and name
func (g *Git) Label(repo, name string) *github.Label {
	for _, m := range g.Repos[repo].Labels {
		if *m.Name == name {
			return m
		}
	}

	return nil
}

// String returns the Git object as a string
func (g *Git) String() string {
	s := ""

	//s += "Projects:\n"
	//for _, p := range g.Projects {
	//	s += fmt.Sprintf("  %d '%s'\n", p.ID, *p.Name)
	//}

	s += "Repos:\n"
	for _, r := range config.Repos {
		s += "  " + r + "\n"
		s += "    milestones:\n"
		for _, m := range g.Repos[r].Milestones {
			s += "      " + *m.Title
			if m.DueOn != nil {
				s += " " + m.DueOn.String()
			}
			s += "\n"
		}
		s += "    labels:\n"
		for _, l := range g.Repos[r].Labels {
			s += "      " + *l.Name + " " + *l.Color + "\n"
		}
	}

	return s
}

func createProject(name, desc string) (*github.Project, error) {
	project, _, err := client.Organizations.CreateProject(ctx, ORG, &github.ProjectOptions{Name: name, Body: desc})
	if err != nil {
		return nil, err
	}

	for _, c := range config.Project.Columns {
		_, _, err := client.Projects.CreateProjectColumn(ctx, project.GetID(), &github.ProjectColumnOptions{Name: c})
		if err != nil {
			return nil, err
		}
	}

	return project, nil
}

func createMilestone(repo, name, date string) (*github.Milestone, error) {
	opt := &github.Milestone{Title: &name}
	if date != "" {
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, err
		}
		opt.DueOn = &t
	}

	m, _, err := client.Issues.CreateMilestone(ctx, ORG, repo, opt)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func createLabel(repo, name, color string) (*github.Label, error) {
	opt := &github.Label{Name: &name, Color: &color}

	l, _, err := client.Issues.CreateLabel(ctx, ORG, repo, opt)
	if err != nil {
		return nil, err
	}

	return l, nil
}
