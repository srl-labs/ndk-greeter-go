name: greeter
prefix: ""

topology:
  nodes:
    greeter:
      kind: nokia_srlinux
      image: ghcr.io/nokia/srlinux:23.10.1
      exec:
        - touch /tmp/.ndk-dev-mode
        {{- if eq (env.Getenv "DEBUG_MODE") "true" }}
        - /debug/prepare-debug.sh
        {{- end }}
      binds:
        - ../build:/tmp/build # mount app binary
        - ../greeter.yml:/tmp/greeter.yml # agent config file to appmgr directory
        - ../yang:/opt/greeter/yang # yang modules
        - ../logs/srl:/var/log/srlinux # expose srlinux logs
        - ../logs/greeter/:/var/log/greeter # expose greeter log file
        {{- if eq (env.Getenv "DEBUG_MODE") "true" }}
        - ../debug/:/debug/
        {{- end }}