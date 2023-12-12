# greeter agent configuration file
# for a complete list of parameters go to
# http://learn.srlinux.dev/ndk/guide/agent/#application-manager-and-application-configuration-file
# --8<-- [start:snip]
greeter:
  path: /usr/local/bin
  launch-command: {{if eq (env.Getenv "DEBUG_MODE") "true" }}{{ "/debug/dlv --listen=:7000"}}{{ if eq (env.Getenv "NOWAIT") "true" }} {{ "--continue --accept-multiclient" }}{{ end }} {{ "--headless=true --log=true --api-version=2 exec"}} {{ end }}/usr/local/bin/greeter
  version-command: /usr/local/bin/greeter --version
  search-command: /usr/local/bin/greeter
  failure-action: wait=10
  config-delivery-format: json
  yang-modules:
    names:
      - greeter
    source-directories:
      - /opt/greeter/yang
# --8<-- [start:end]
