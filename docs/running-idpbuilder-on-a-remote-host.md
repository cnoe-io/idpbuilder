# Running IDPBuilder on a remote host

## Option 1: SSH Port forwarding

This option is the most flexible and involves using an ssh connection to forward traffic from local ports to the server where IDPBuilder was run.
First create your cluster on the server:

```shell
user@server:~/$ idpbuilder create
```

Once your cluster is created we need to configure our port forwards:

```shell
user@local:~/$ ssh -L 8443:server:8443 -L 32222:server:32222 user@server
```

`-L 8443:server:8443` adds portforwarding for the ingress.

`-L 32222:server:32222` adds portforwarding for the gitea ssh port.

If you want to use kubectl on your local machine first find the port the kube-api is exposed on:

```
user@server:~/$ idpbuilder get clusters
NAME       EXTERNAL-PORT   KUBE-API                  TLS     KUBE-PORT   NODES
localdev   8443            https://127.0.0.1:36091   false   6443        localdev-control-plane
```

In this case it is exposed on 36091. Then add the following to your ssh command:

`-L 36091:server:36091`

Finally copy the kube config from the server to the local machine:

```shell
user@local:~/$ mkdir -p ~/.kube
user@local:~/$ scp user@server:~/.kube/config ~/.kube/config
```

## Option 2: Changing the ingress host

If you only need remote access to the ingress you can build your remote cluster using the following options:

```shell
user@server:~/$ idpbuilder create --host SERVER.DOMAIN.NAME.HERE --use-path-routing
```

note that this doesn't work with the `--dev-password` flag.
