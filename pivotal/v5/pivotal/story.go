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

package pivotal

import (
	"fmt"
	"net/http"
)

type Story struct {
	project *Project
	id      int
}

func (story *Story) AddTask(inTask *Task) (outTask *Task, resp *http.Response, err error) {
	if inTask.Description == "" {
		return nil, nil, &ErrFieldNotSet{"description"}
	}

	u := fmt.Sprintf("projects/%v/stories/%v/tasks", story.project.id, story.id)
	req, err := story.project.client.NewRequest("POST", u, inTask)
	if err != nil {
		return nil, nil, err
	}

	var task Task
	resp, err = story.project.client.Do(req, nil)
	if err != nil {
		return nil, resp, err
	}

	return &task, resp, err
}

func (story *Story) ListTasks() ([]*Task, *http.Response, error) {
	u := fmt.Sprintf("projects/%v/stories/%v/tasks", story.project.id, story.id)
	req, err := story.project.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	var tasks []*Task
	resp, err := story.project.client.Do(req, &tasks)
	if err != nil {
		return nil, resp, err
	}

	return tasks, resp, err
}

func (story *Story) UpdateTask(inTask *Task) (outTask *Task, resp *http.Response, err error) {
	u := fmt.Sprintf("projects/%v/stories/%v/tasks/%v", story.project.id, story.id, inTask.Id)
	req, err := story.project.client.NewRequest("PUT", u, inTask)
	if err != nil {
		return nil, nil, err
	}

	var task Task
	resp, err = story.project.client.Do(req, &task)
	if err != nil {
		return nil, resp, err
	}

	return &task, resp, err
}
