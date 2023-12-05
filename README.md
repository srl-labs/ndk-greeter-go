# Greeter App in Go

Greeter is a demo Go application demonstrating the key principles of creating applications powered by the SR Linux's [NetOps Development Kit](https://learn.srlinux.dev/ndk/). Check [learn.srlinux.dev](https://learn.srlinux.dev/ndk/guide/env/go/) for a complete code walkthrough.

## Quickstart

Clone and enter the repository:

```bash
git clone https://github.com/srl-labs/ndk-greeter-go.git && \
cd ndk-greeter-go
```

Build the application and deploy it to the lab:

```
./run.sh deploy-all
```

Once the lab is deployed, the application is automatically onboarded to SR Linux.

Enter the SR Linux CLI:

```
ssh greeter
```

Once entered into the SR Linux CLI, you can finde `/greeter` context available that contains the application's configuration and operational data.

Configure the desired name:

```
--{ + running }--[  ]--
A:greeter# enter candidate

--{ + candidate shared default }--[  ]--
A:greeter# greeter name srlinux-user
```

Commit the configuration:

```
--{ +* candidate shared default }--[  ]--
A:greeter# commit now
All changes have been committed. Leaving candidate mode.
```

The application will now greet you when you list its operational state:

```
--{ + running }--[  ]--
A:greeter# info from state greeter
    greeter {
        name srlinux-user
        greeting "👋 Hello srlinux-user, I was last booted at 2023-11-26T10:24:27.374Z"
    }
```

## Shell autocompletions

To get bash autocompletions for `./run.sh` functions:

```bash
source ./run.sh
```
