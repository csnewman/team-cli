# team-cli
Command line interface for AWS TEAM.

### Install

Install command:

```bash
go install github.com/csnewman/team-cli/cmd/team-cli@latest
```

Configure remote server:
```
team-cli configure team.you-company.com
```

### Usage

List accounts:
```
$ team-cli list-accounts

Accounts:
  [1] id="123123123123" name="example"
    - role="ReadOnlyAccess" max_duration=8 requires_approval=false
```

Request access interactively:
```
$ team-cli request

Please select the account:
  [1] id="123123123123" name="example"

Account option? 
```

Request access non-interactively:
```
$ team-cli request --account "example" --role="readonlyaccess" --duration 3 --ticket "support-123" --reason "Demo" --start "now" -y

Details:
  Account: id="123123123123" name="example"
  Role: name="ReadOnlyAccess"
  Start: now
  Ticket: "support-123"
  Justification: "Demo"

Request submitted
Request ID: 00000000-0000-0000-0000-000000000000
```

### TEAM install configuration

The default cognito client app does not allow localhost redirects upon successful authentication. `team-cli` requires
it's callback address to be added to the allowed list to be able to fetch authentication tokens.

#### Via web UI:

Add `http://localhost:43672/` to the `team06dbb7fc_app_clientWeb` app client in the Cognito `team` user pool.

![img.png](.github/callback.png)
