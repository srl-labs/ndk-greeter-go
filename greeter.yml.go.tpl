# greeter agent configuration file
# for a complete list of parameters go to
# http://learn.srlinux.dev/ndk/guide/agent/#application-manager-and-application-configuration-file
# --8<-- [start:snip]
greeter:
  path: /usr/local/bin
  launch-command: {{if ne (env.Getenv "NDK_DEBUG") "" }}{{ "/debug/dlv --listen=:7000"}}{{ if ne (env.Getenv "NDK_DEBUG") "" }} {{ "--continue --accept-multiclient" }}{{ end }} {{ "--headless=true --log=true --api-version=2 exec"}} {{ end }}greeter
  version-command: greeter --version
  failure-action: wait=10
  config-delivery-format: json
  yang-modules:
    names:
      - greeter
    source-directories:
      - /opt/greeter/yang
# --8<-- [start:end]
