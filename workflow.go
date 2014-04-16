/*
   Copyright (C) 2013  Salsita s.r.o.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program. If not, see {http://www.gnu.org/licenses/}.
*/

package main

import (
	// Stdlib
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	// Workflow
	"cider-salsita-workflow/pivotal/v5/pivotal"
	"cider-salsita-workflow/poblano/v1/poblano"

	// Cider
	"github.com/cider/go-cider/cider/services/logging"
	"github.com/cider/go-cider/cider/services/pubsub"
)

type Workflow struct {
	directory *poblano.Client
	eventBus  *pubsub.Service
	logger    *logging.Service
}

type GithubIssueEvent struct {
	Action string `codec:"action"`
	Issue  *struct {
		Body    string `codec:"body"`
		URL     string `codec:"url"`
		HTMLURL string `codec:"html_url"`
		User    struct {
			Login string `codec:"login"`
		} `codec:"user"`
	} `codec:"issue"`
}

func (w *Workflow) AddPtTaskFromGhIssue(event pubsub.Event) {
	var (
		log    = w.logger
		caller = methodName()
	)

	// Unmarshal the event object.
	var issueEvent GithubIssueEvent
	if err := event.Unmarshal(&issueEvent); err != nil {
		log.Warnf("%s: %v", caller, err)
		return
	}

	// Only the issue opened event matters here.
	if action := issueEvent.Action; action != "opened" {
		log.Infof("%s: Actually an issue %s event, skipping...", caller, action)
		return
	}

	issue := issueEvent.Issue
	if issue.Body == "" {
		log.Infof("%s: Issue body is empty, skipping...", caller)
		return
	}

	// Look for the Pivotal Tracker story URL.
	pattern := regexp.MustCompile("https://www.pivotaltracker.com/story/show/([0-9]+)")

	match := pattern.FindStringSubmatch(issue.Body)
	if match == nil || len(match) != 2 {
		log.Infof("%s: No Pivotal Tracker URL detected, skipping...", caller)
		return
	}

	storyId, err := strconv.Atoi(string(match[1]))
	if err != nil {
		log.Warnf("%s: %v", caller, err)
		return
	}

	// Fetch Poblano records that are required.
	issueURL, err := url.Parse(issue.URL)
	if err != nil {
		log.Warnf("%s: %v", caller, err)
		return
	}

	fragments := strings.Split(issueURL.Path, "/")
	if len(fragments) != 6 {
		log.Warnf("%s: Unexpected GitHub URL encountered: %s", caller, issue.URL)
		return
	}

	gh := w.directory.GitHub

	var (
		repoOwner = fragments[2]
		repoName  = fragments[3]
	)
	log.Debugf("%s: Getting Poblano project record for repository %v...", caller, repoName)
	project, _, err := gh.GetPoblanoProject(repoOwner, repoName)
	if err != nil {
		log.Errorf("%s: %v", caller, err)
		return
	}
	log.Debugf("%s: Poblano project record received", caller)

	login := issue.User.Login
	log.Debugf("%s: Getting the Poblano user record for login %v...", caller, login)
	user, _, err := gh.GetPoblanoUser(login)
	if err != nil {
		log.Errorf("%s: %v", caller, err)
		return
	}
	log.Debugf("%s: Poblano user record received", caller)

	// Add task to the relevant PT story.
	pt := pivotal.NewClient(user.Services.PivotalTracker.AccessToken)
	story := pt.Project(project.Services.PivotalTracker.Id).Story(storyId)

	if _, err := story.AddTask(&pivotal.Task{
		Description: "GitHub issue " + issue.HTMLURL,
	}); err != nil {
		log.Errorf("%s: %v", caller, err)
		return
	}

	log.Infof("%s: Pivotal Tracker story task created for GitHub issue %s", caller, issue.HTMLURL)
}

// Helpers ---------------------------------------------------------------------

func methodName() (name string) {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		fullName := runtime.FuncForPC(pc).Name()
		parts := strings.Split(fullName, ".")
		name = parts[len(parts)-1]
	} else {
		name = "unknown method"
	}
	return
}
