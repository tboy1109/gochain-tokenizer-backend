# Quickstart API Template

For rapid app creation.

[Frontend template](https://github.com/treeder/flap-ui-template)

## Getting Started

* Click `Use this template` button above
* Edit `go.mod`, change module name to your new repo name
* Create  firebase project
    * Upgrade project to pay as you go
    * Create a Firestore
    * Go to settings -> Service accounts, click "Generate new private key". This will download a JSON file. 
    * Run `base64 -w 0 account.json` to get encoded version of the file (for secrets)
    * Make `.env` file with `G_KEY` (output of above command) and `G_PROJECT_ID`
        * Or add those vars into your codespace secrets
* `make run` (boom)

## Auto Deploying

* Go to https://console.cloud.google.com/ , choose your firebase project
* Go to Cloud run -> create service
* Mostly defaults, but choose deploy from github and choose your repo
* No Google env vars required, but add any new ones you created
