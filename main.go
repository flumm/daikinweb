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

var conf *Config
var configFile = "/etc/daikinweb/config.json"

func main() {
	conf = LoadConfig(configFile)
	listenAddress := ":" + fmt.Sprintf("%v", conf.Port)
	log.Println("loaded config")
	log.Println("setting up router")
	router := mux.NewRouter()
	router.HandleFunc("/units", getUnits).Methods("GET")
	router.HandleFunc("/units", controllAllUnits).Methods("PUT")
	router.HandleFunc("/units/{unit}", getUnit).Methods("GET")
	router.HandleFunc("/units/{unit}/control", setUnitControl).Methods("POST")
	router.HandleFunc("/units/{unit}/{infoType}", getUnitInfo).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(conf.WebDir)))
	log.Println("set up router")
	log.Println("starting server on", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, router))
}

func getUnits(w http.ResponseWriter, r *http.Request) {

	var data = make([]map[string]string, 0)
	var errs = make([]error, 0)
	for name, ip := range conf.Units {
		unit := daikingo.NewUnit(ip)

		info, err := unit.GetBasicInfo()
		if err != nil {
			errs = append(errs, err)
		}

		sensor, err := unit.GetSensorInfo()
		if err != nil {
			errs = append(errs, err)
		}

		control, err := unit.GetControlInfo()
		if err != nil {
			errs = append(errs, err)
		}

		for k, v := range sensor {
			info[k] = v
		}
		for k, v := range control {
			info[k] = v
		}
		info["conf_name"] = name
		data = append(data, info)
	}

	if len(errs) > 0 {
		log.Println(errs)
	}

	_ = json.NewEncoder(w).Encode(data)
}

func controllAllUnits(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	params := r.PostForm
	var result = make(map[string]map[string]string, 0)
	var errs = make([]error, 0)
	for name, ip := range conf.Units {
		unit := daikingo.NewUnit(ip)
		data, err := unit.SetControlInfo(params)

		result[name] = data
		if err != nil {
			errs = append(errs, err)
			return
		}
	}

	if len(errs) > 0 {
		log.Println(errs)
	}

	_ = json.NewEncoder(w).Encode(result)
}

func getUnit(w http.ResponseWriter, r *http.Request) {
	unitIP, ok := conf.Units[mux.Vars(r)["unit"]]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	unit := daikingo.NewUnit(unitIP)

	data, err := unit.GetBasicInfo()

	if err != nil {
		http.Error(w, "Error getting Info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(data)
}

func getUnitInfo(w http.ResponseWriter, r *http.Request) {
	unitIP, ok := conf.Units[mux.Vars(r)["unit"]]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var data map[string]string
	var err error
	unit := daikingo.NewUnit(unitIP)

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

	_ = json.NewEncoder(w).Encode(data)
}

func setUnitControl(w http.ResponseWriter, r *http.Request) {
	unitIP, ok := conf.Units[mux.Vars(r)["unit"]]
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	unit := daikingo.NewUnit(unitIP)
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	data, err := unit.SetControlInfo(r.PostForm)

	if err != nil {
		http.Error(w, "Error getting Info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(data)
}
