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

type Project struct {
	Name        string
	Slug        string
	Description string
	Services    struct {
		GitHub *struct {
			Id        int
			Name      string
			URL       string
			Connected bool
		} `json:"github"`
		PivotalTracker *struct {
			Id        int
			URL       string
			Connected bool
		} `json:"pivotalTracker"`
	}
}

type User struct {
	Name     string
	Email    string
	Services struct {
		GitHub *struct {
			Username    string
			AccessToken string
			Connected   bool
		} `json:"github"`
		PivotalTracker *struct {
			Id          int
			Username    string
			AccessToken string
			Connected   bool
		} `json:"pivotalTracker"`
	}
}
