// go-netgear: Get attached devices from your Netgear router
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
    "net"
    "net/http"
    "io/ioutil"
    "bytes"
    "strings"
)

const SESSION_ID = "A7D88AE69687E58D9A00"

const SOAP_LOGIN_ACTION = "urn:NETGEAR-ROUTER:service:ParentalControl:1#Authenticate"
const SOAP_LOGIN = `\
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

const SOAP_ATTACHED_DEVICES_ACTION = "urn:NETGEAR-ROUTER:service:DeviceInfo:1#GetAttachDevice"
const SOAP_ATTACHED_DEVICES = `\
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

type AttachedDevice struct {
    Signal int `json:"signal"`
    IP net.IP `json:"ip"`
    Name string `json:"name"`
    Mac string `json:"mac"`
    Type string `json:"type"`
    LinkRate int `json:"link_rate"`
}

type Netgear struct {
    host string
    username string
    password string
    loggedIn bool
}

func (netgear *Netgear) IsLoggedIn() bool {
    return netgear.loggedIn
}

func (netgear *Netgear) Login() (bool, error) {
    message := fmt.Sprintf(SOAP_LOGIN, SESSION_ID, netgear.username, netgear.password)

    resp, err := netgear.makeRequest(SOAP_LOGIN_ACTION, message)

    if strings.Contains(resp, "<ResponseCode>000</ResponseCode>") {
        netgear.loggedIn = true
    } else {
        netgear.loggedIn = false
    }
    return netgear.loggedIn, err
}

func (netgear *Netgear) makeRequest(action string, message string) (string, error) {
    client := &http.Client{}

    url := fmt.Sprintf("http://%s:5000/soap/server_sa/", netgear.host)

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

