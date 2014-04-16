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

package poblano

import (
	"errors"
	"fmt"
	"net/http"
)

type GitHubService struct {
	client *Client
}

func newGitHubService(client *Client) *GitHubService {
	return &GitHubService{client}
}

func (srv *GitHubService) GetPoblanoProject(repoOwner, repoName string) (*Project, *http.Response, error) {
	u := fmt.Sprintf("/api/projects?where[services.github.fullName]=%v/%v", repoOwner, repoName)
	req, err := srv.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var projects []*Project
	resp, err := srv.client.Do(req, &projects)
	if err != nil {
		return nil, resp, err
	}

	switch len(projects) {
	case 0:
		return nil, resp, errors.New("Poblano project record not found")
	case 1:
		return projects[0], resp, nil
	default:
		return nil, resp, errors.New("Poblano returned multiple project records")
	}
}

func (srv *GitHubService) GetPoblanoUser(login string) (*User, *http.Response, error) {
	u := fmt.Sprintf("/api/users?where[services.github.username]=%v", login)
	req, err := srv.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var users []*User
	resp, err := srv.client.Do(req, &users)
	if err != nil {
		return nil, resp, err
	}

	switch len(users) {
	case 0:
		return nil, resp, errors.New("Poblano user record not found")
	case 1:
		return users[0], resp, nil
	default:
		return nil, resp, errors.New("Poblano returned multiple user records")
	}
}
