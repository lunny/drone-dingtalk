// Copyright 2017 Lunny Xiao. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/lunny/dingtalk_webhook"
)

type (
	// Repo information
	Repo struct {
		Owner string
		Name  string
	}

	// Build information
	Build struct {
		Tag      string
		Event    string
		Number   int
		Commit   string
		RefSpec  string
		Branch   string
		Author   string
		Avatar   string
		Message  string
		Email    string
		Status   string
		Link     string
		Started  float64
		Finished float64
	}

	// Config for the plugin.
	Config struct {
		AccessToken string
		Message     string
		IsAtAll     bool
		Drone       bool
		Username    string
		AvatarURL   string
	}

	// Plugin values.
	Plugin struct {
		Repo    Repo
		Build   Build
		Config  Config
		Webhook *dingtalk.Webhook
	}
)

// Exec executes the plugin.
func (p *Plugin) Exec() error {
	if len(p.Config.AccessToken) == 0 {
		log.Println("missing dingtalk config")
		return errors.New("missing dingtalk config")
	}

	p.Webhook = dingtalk.NewWebhook(p.Config.AccessToken)

	if len(p.Config.Message) == 0 {
		log.Println("missing message to send")
		return errors.New("missing message to send")
	}

	if p.Config.Drone {
		err := p.Webhook.SendPayload(p.DroneTemplate())
		if err != nil {
			return err
		}
	}

	return p.Webhook.SendTextMsg(p.Config.Message, p.Config.IsAtAll)
}

// DroneTemplate is plugin default template for Drone CI.
func (p *Plugin) DroneTemplate() *dingtalk.Payload {
	description := ""
	//Color:       p.Color(),
	switch p.Build.Event {
	case "push":
		description = fmt.Sprintf("%s pushed to %s", p.Build.Author, p.Build.Branch)
	case "pull_request":
		branch := ""
		if p.Build.RefSpec != "" {
			branch = p.Build.RefSpec
		} else {
			branch = p.Build.Branch
		}
		description = fmt.Sprintf("%s updated pull request %s", p.Build.Author, branch)
	case "tag":
		description = fmt.Sprintf("%s pushed tag %s", p.Build.Author, p.Build.Branch)
	}

	return &dingtalk.Payload{
		MsgType: "actionCard",
		ActionCard: struct {
			Text           string `json:"text"`
			Title          string `json:"title"`
			HideAvatar     string `json:"hideAvatar"`
			BtnOrientation string `json:"btnOrientation"`
			SingleTitle    string `json:"singleTitle"`
			SingleURL      string `json:"singleURL"`
			Buttons        []struct {
				Title     string `json:"title"`
				ActionURL string `json:"actionURL"`
			} `json:"btns"`
		}{
			Title:          p.Build.Message,
			Text:           description,
			HideAvatar:     "0",
			BtnOrientation: "0",
			SingleTitle:    "Drone",
			SingleURL:      p.Build.Link,
		},
	}
}
