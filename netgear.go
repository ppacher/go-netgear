// go-netgear: Go API to retrieve devices attached to modern Netgear routers
//
// This is a golang rewrite of the python module pynetgear developed by balloob:
// <https.//github.com/balloob/pynetgear>
//
// website: https://github.com/nethack42/go-netgear
// author:  Patrick Pacher <patrick.pacher@gmail.com>
//
// Copyright 2015 Patrick Pacher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//     Unless required by applicable law or agreed to in writing, software
//     distributed under the License is distributed on an "AS IS" BASIS,
//     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//     See the License for the specific language governing permissions and
//     limitations under the License.
//
//
package netgear

import (
    "fmt"
    "regexp"
    "net/http"
    "io/ioutil"
    "bytes"
    "strings"
)

const sessionID = "A7D88AE69687E58D9A00"

const soapActionLogin = "urn:NETGEAR-ROUTER:service:ParentalControl:1#Authenticate"
const soapLoginMessage = `\
<?xml version="1.0" encoding="utf-8" ?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
<SOAP-ENV:Header>
<SessionID xsi:type="xsd:string" xmlns:xsi="http://www.w3.org/1999/XMLSchema-instance">%s</SessionID>
</SOAP-ENV:Header>
<SOAP-ENV:Body>
<Authenticate>
  <NewUsername>%s</NewUsername>
  <NewPassword>%s</NewPassword>
</Authenticate>
</SOAP-ENV:Body>
</SOAP-ENV:Envelope>
`

const soapActionGetAttachedDevices = "urn:NETGEAR-ROUTER:service:DeviceInfo:1#GetAttachDevice"
const soapGetAttachedDevicesMesssage = `\
<?xml version="1.0" encoding="utf-8" standalone="no"?>
<SOAP-ENV:Envelope xmlns:SOAPSDK1="http://www.w3.org/2001/XMLSchema" xmlns:SOAPSDK2="http://www.w3.org/2001/XMLSchema-instance" xmlns:SOAPSDK3="http://schemas.xmlsoap.org/soap/encoding/" xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
<SOAP-ENV:Header>
<SessionID>%s</SessionID>
</SOAP-ENV:Header>
<SOAP-ENV:Body>
<M1:GetAttachDevice xmlns:M1="urn:NETGEAR-ROUTER:service:DeviceInfo:1">
</M1:GetAttachDevice>
</SOAP-ENV:Body>
</SOAP-ENV:Envelope>
`

// AttachedDevice represents a network device attached to the Netgear router via 
// a wired or wireless link
type AttachedDevice struct {
    Signal string `json:"signal"`
    IP string `json:"ip"`
    Name string `json:"name"`
    Mac string `json:"mac"`
    Type string `json:"type"`
    LinkRate string `json:"link_rate"`
}

// Netgear describes a modern Netgear router providing a SOAP interface at port
// 5000
type Netgear struct {
    host string
    username string
    password string
    loggedIn bool
    regex *regexp.Regexp
}

// IsLoggedIn returns true if the session has been authenticated against the 
// Netgear Router or false otherwise.
func (netgear *Netgear) IsLoggedIn() bool {
    return netgear.loggedIn
}

// Login authenticates the session against the Netgear router
// On success true and nil should be returned. Otherwise false and
// the related error are returned
func (netgear *Netgear) Login() (bool, error) {
    message := fmt.Sprintf(soapLoginMessage, sessionID, netgear.username, netgear.password)

    resp, err := netgear.makeRequest(soapActionLogin, message)

    if strings.Contains(resp, "<ResponseCode>000</ResponseCode>") {
        netgear.loggedIn = true
    } else {
        netgear.loggedIn = false
    }
    return netgear.loggedIn, err
}

func (netgear *Netgear) getUrl() string {
    return fmt.Sprintf("http://%s:5000/soap/server_sa/", netgear.host)
}

func (netgear *Netgear) makeRequest(action string, message string) (string, error) {
    client := &http.Client{}

    url := netgear.getUrl()

    req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(message)))
    if err != nil {
        return "", err
    }
    req.Header.Add("SOAPAction", action)

    response, err := client.Do(req)
    if err != nil {
        return "", err
    }

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return "", err
    }

    return string(body), err
}


// GetAttachedDevices queries the Netgear router for attached network
// devices and returns a list of them. If an error occures an empty list
// and the respective error is returned.
func (netgear *Netgear) GetAttachedDevices() ([]AttachedDevice, error) {
    var result []AttachedDevice

    message := fmt.Sprintf(soapGetAttachedDevicesMesssage, sessionID)
    resp, err := netgear.makeRequest(soapActionGetAttachedDevices, message)

    if strings.Contains(resp, "<ResponseCode>000</ResponseCode>") {
        re := netgear.regex.FindStringSubmatch(resp)
        if len(re) < 2 {
            err = fmt.Errorf("Invalid response code")
            return result, err
        }
        devices := re[1]
        fields := strings.Split(devices, ";")

        countItems := int((len(fields)-1)/6)

        for i := 0; i < countItems ; i++ {
            device := AttachedDevice{
                Signal: fields[i*6+0],
                IP: fields[i*6+1],
                Name: fields[i*6+2],
                Mac: fields[i*6+3],
                Type: fields[i*6+4],
                LinkRate: fields[i*6+5],
            }
            result = append(result, device)
        }

    }
    return result, err
}

// NewRouter returns a new and already initialized Netgear router instance
// However, the Netgear SOAP session has not been authenticated at this point.
// Use Login() to authenticate against the router
func NewRouter(host, username, password string) *Netgear {
    router := &Netgear{
        host: host,
        username: username,
        password: password,
        regex : regexp.MustCompile("<NewAttachDevice>(.*)</NewAttachDevice>"),
    }
    return router
}

