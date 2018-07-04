# Sauron

Sees things

## What
URL parser, fetcher and related automation tools.

## Why
Every intelligent thing needs an eye to see other things. This project
is created to serve ongoing AI research/development for Amerikadaniste.com.

## How
You can either use GUI or API (Json) to interact and enjoy Sauron.

### GUI
High hopes for gui. Currently on roadmap but can be cutout and removed
altogether

### API
Sauron provides a Json REST API to interact with your predefined data
extraction processes.

## Setup
To use `go get` for this project's dependencies you need to be able to
use git on [github.com](https://github.com) repositories since
currently this repository and one of his dependencies is hosted there.

To remove annoying `git-credential-manager` popup and use your cute ssh
client instead, use this command: `git config --global url.
"git@github.com:".insteadOf "https://github.com/"`

## Running
Create frontend network using `docker network create frontend`

Create dns redirections for `api.localhost` and `proxy.localhost` to 127.0.0.1

Run Traefik: `docker run
-d -v /var/run/docker.sock:/var/run/docker.sock
-v ./traefik.toml:/traefik.toml
-p 80:80
-p 443:443
-l traefik.frontend.rule=Host:traefik.localhost
-l traefik.port=8080
--network frontend
--name traefik
traefik:1.3.6-alpine --docker`

Deploy stack using `docker-compose up -d`



