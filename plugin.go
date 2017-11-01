// Copyright 2017 Lunny Xiao. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

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
		Lang        string
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

	if p.Config.Drone {
		return p.Webhook.SendPayload(p.DroneTemplate())
	}

	if len(p.Config.Message) == 0 {
		log.Println("missing message to send")
		return errors.New("missing message to send")
	}

	return p.Webhook.SendTextMsg(p.Config.Message, p.Config.IsAtAll)
}

func (p *Plugin) getTemplate(event string) string {
	if p.Config.Lang == "zh_CN" {
		switch p.Build.Event {
		case "push":
			return `# [%s](%s)
		
![avatar](%s) %s 推送到 %s 的 %s 分支 %s`
		case "pull_request":
			return "%s 更新了 %s 合并请求 %s"
		case "tag":
			return "%s 推送了 %s 标签 %s"
		}
	} else {
		switch p.Build.Event {
		case "push":
			return `# [%s](%s)

![avatar](%s) %s pushed to %s branch %s %s`
		case "pull_request":
			return "%s updated %s pull request %s"
		case "tag":
			return "%s pushed %s tag %s"
		}
	}
	return ""
}

// DroneTemplate is plugin default template for Drone CI.
func (p *Plugin) DroneTemplate() *dingtalk.Payload {
	description := ""

	switch p.Build.Event {
	case "push":
		description = fmt.Sprintf(p.getTemplate(p.Build.Event),
			strings.TrimSpace(p.Build.Message),
			p.Build.Link,
			p.Config.AvatarURL,
			p.Build.Author,
			p.Repo.Owner+"/"+p.Repo.Name,
			p.Build.Branch,
			p.Build.Status)
	case "pull_request":
		branch := ""
		if p.Build.RefSpec != "" {
			branch = p.Build.RefSpec
		} else {
			branch = p.Build.Branch
		}
		description = fmt.Sprintf(p.getTemplate(p.Build.Event), p.Build.Author, p.Repo.Owner+"/"+p.Repo.Name, branch)
	case "tag":
		description = fmt.Sprintf(p.getTemplate(p.Build.Event), p.Build.Author, p.Repo.Owner+"/"+p.Repo.Name, p.Build.Branch)
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
