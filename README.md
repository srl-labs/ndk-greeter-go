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

Configure the name of the person to greet:

```
