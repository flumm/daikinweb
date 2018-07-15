/*
   Daikin Web provides an api proxy/web interface for Daikin Emura FTXG-L AC units.
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
	"github.com/flumm/daikingo"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configDir = "/etc/daikinweb/"
const configFile = configDir + "config.json"

var conf *Config
var units map[string]*daikingo.Unit = make(map[string]*daikingo.Unit, 0)

func main() {
	conf = LoadConfig(configFile)
	// init units
	for k, v := range conf.Units {
		unit := daikingo.NewUnit(v)
		units[k] = unit
	}
	listenAddress := ":" + fmt.Sprintf("%v", conf.Port)
	log.Println("loaded config")
	log.Println("setting up router")
	router := mux.NewRouter()
	router.HandleFunc("/units", getUnits).Methods("GET")
	router.HandleFunc("/units", controllAllUnits).Methods("PUT")
	router.HandleFunc("/units/{unit}", getUnit).Methods("GET")
	router.HandleFunc("/units/{unit}/control", setUnitControl).Methods("PUT")
	router.HandleFunc("/units/{unit}/{infoType}", getUnitInfo).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(conf.WebDir)))
	log.Println("set up router")
	log.Println("starting server on", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, router))
}

type APIResult struct {
	Data interface{} `json:"data"`
	Err  interface{} `json:"err,omitempty"`
}

func getUnits(w http.ResponseWriter, r *http.Request) {

	var data = make([]map[string]string, 0)
	var errs = make([]string, 0)
	for name, unit := range units {
		unitData := make(map[string]string)

		info, err := unit.GetBasicInfo()
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			for k, v := range info {
				unitData[k] = v
			}
		}

		sensor, err := unit.GetSensorInfo()
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			for k, v := range sensor {
				unitData[k] = v
			}
		}

		control, err := unit.GetControlInfo()
		if err != nil {
			errs = append(errs, err.Error())
		} else {
			for k, v := range control {
				unitData[k] = v
			}
		}

		unitData["name"] = name
		data = append(data, unitData)
	}

	result := APIResult{data, nil}
	if len(errs) > 0 {
		result.Err = errs
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Println(err)
	}
}

func controllAllUnits(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	params := r.PostForm
	var data = make(map[string]map[string]string)
	var errs = make([]error, 0)
	for name, unit := range units {
		info, err := unit.SetControlInfo(params, true)

		data[name] = info
		if err != nil {
			errs = append(errs, err)
			return
		}
	}

	result := APIResult{data, nil}
	if len(errs) > 0 {
		result.Err = errs
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		log.Println(err)
	}
}

func getUnit(w http.ResponseWriter, r *http.Request) {
	unit, ok := units[mux.Vars(r)["unit"]]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	data, err := unit.GetBasicInfo()

	if err != nil {
		http.Error(w, "Error getting Info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println(err)
	}
}

func getUnitInfo(w http.ResponseWriter, r *http.Request) {
	unit, ok := units[mux.Vars(r)["unit"]]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var data map[string]string
	var err error

	switch mux.Vars(r)["infoType"] {
	case "basic":
		data, err = unit.GetBasicInfo()
	case "sensor":
		data, err = unit.GetSensorInfo()
	case "control":
		data, err = unit.GetControlInfo()
	case "model":
		data, err = unit.GetModelInfo()
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Error getting Info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println(err)
	}
}

func setUnitControl(w http.ResponseWriter, r *http.Request) {
	unit, ok := units[mux.Vars(r)["unit"]]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	data, err := unit.SetControlInfo(r.PostForm, true)

	if err != nil {
		http.Error(w, "Error setting control: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println(err)
	}
}
