# Demo Backend for Course Project

This is a demonstration server for the course project.  This demonstrates some of nodes architecture features leveraging the `fastify` project.

Some of the interesting aspects include managing plugins, building servers that validate input using `json-schema`, properly managing secrets for a reverse proxy to `github apis`, swagger documentation and a few other things.

In order to execute the server in watch mode:

`yarn run watch`

Also make sure you have a personal github action token and set it up in a file named `.env` as this file will not be pushed to the github repository.

### First time installation instructions

After you pull this repo from github:

1. Create a personal access token in `github`
2. Creeate a .env file and put the following key/value pair inside `GITHUB_ACCESS_TOKEN = "THE-PERSONAL-ACCESS-TOKEN-FROM-GITHUB-FROM-STEP-1"`
3. Install all dependencies via `yarn` or `npm install`
4. Run the server via `yarn run watch`
5. Note default hard coded port is `4080`, change in `server.ts` if you want