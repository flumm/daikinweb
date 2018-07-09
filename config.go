/*
   This file provides a basic config file implementation.
   Copyright (C) 2018  Dominik Csapak <dominik.csapak@gmail.com>
   This program is free software; you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation; either version 2 of the License, or
   (at your option) any later version.
   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License along
   with this program; if not, write to the Free Software Foundation, Inc.,
   51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Units  map[string]string
	WebDir string
	Port   int
}

func LoadConfig(FileName string) *Config {
	var data = new(Config)

	f, err := os.Open(FileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open config, loading defaults...: ", err)
		goto defaults
	}

	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse config, loading defaults...: ", err)
		goto defaults
	}

defaults:
	if data.WebDir == "" {
		data.WebDir = "./www/"
	}

	if data.Port == 0 {
		data.Port = 8080
	}
	return data
}
