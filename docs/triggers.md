# Triggers

*In proposal phase.*

Triggers are integrations that you can use in other systems to fire off functions in IronFunctions.

They are downloaded from Iron.io's TriggerHub and linked to the specified Function:

```sh
fnctl trigger link http://functions.example.com/myapp/myfunction author/trigger [--opt=...]
```

You can undo it with:
```sh
fnctl trigger unlink http://functions.example.com/myapp/myfunction author/trigger [--opt=...]
```

Each triggers knows how to correctly link itself with the function, and additional options may be added to make it happen.

You can browse the list of triggers locally with:
```sh
fnctl trigger search iron
name                      description
iron/github-webhook       configures your Github account to trigger events to IronFunctions
iron/swift-file-updated   adds a trigger to an OpenStack Swift service
[...]
```

Check more information about a specific trigger:
```sh
fnctl trigger search info github-webhook

TRIGGER: iron/github-webhook

DESCRIPTION: configures your Github account to trigger events to IronFunctions.
Detailed documentation can be found at https://fn.iron.io/hub/iron/github-webhook

OPTIONS:
- None

ENVIRONMENT VARIABLES
- GITHUB_TOKEN: the Github token to be used for setting this trigger.
- GITHUB_REPO: the Github repository to which this token must be connected to.

EXAMPLE:

GITHUB_TOKEN=x GITHUB_REPO=myrepo fnctl trigger link http://functions.example.com/myapp/myfunction iron/github-hook
```
