<h1 align="center"> 
  CBus - Clustered DBus
  <br>
  <br>
</h1>

**Status: highly experimental**

# Motivation

I've been thinking about a clustered `systemctl` for a while now, and when I started digging into it recently I
eventually ended up looking closely at [DBus](https://www.freedesktop.org/wiki/Software/dbus/).

What I _want to achieve_ is remotely monitor and manage multiple machines.

I think DBus provides the underlying mechanism for this in the form of:

-   Properties
-   Methods
-   Signals

[Systemd](https://www.freedesktop.org/wiki/Software/systemd/dbus/) integrates deeply with DBus, registering an object
for every unit, so right out of the gate that's immensely useful without looking at what other buses/objects you might
find on a Linux system.

This is what inspired me to integrate DBus with NATS.

## Development

### Requirements

You must be familiar with and have the following installed:

-   [Direnv](https://direnv.net)
-   [Nix](https://nixos.org)

### Quick Start

After cloning the repository cd into the root directory and run:

```terminal
direnv allow
```

You will be asked to accept config for some additional substituters to which you should select yes.

Once the dev shell has finished initialising you can see the available commands by typing `menu`:

```terminal
❯ menu

[[general commands]]

  flake-linter
  menu             - prints this menu

[development]

  dev              - run local dev services
  dev-init         - re-initialise data directory
  gomod2nix        - Convert applications using Go modules -> Nix
  run-test-machine - run a test vm

[docs]

  gifs             - Generate all gifs used in docs
  vhs              - A tool for generating terminal GIFs with code

[formatting]

  fmt              - format the repo

[nats]

  nats
  nsc
```

To start the local dev services type `dev`.

![](./docs/assets/dev.gif)

You must wait until `nats-setup` has completed successfully and `nats-server` is in the running state.

After a short while the test machines (vms) should start up, and you can try retrieving some properties and invoking
some methods.

#### Invoke a method on all machines

```terminal
❯ nix run .# -- invoke org.freedesktop.systemd1 /org/freedesktop/systemd1 GetDefaultTarget
NKey: UBQUIKYGFC7CH5XMF52P2NN4ESI4XTXGAKT3WTG3XGU3352DMHGZQAX7

["default.target"]

NKey: UAILZJS66U24VALTH2D6OWVD5JTS3IDFMKOICJSOX7CHNAF3AXFVWIFX

["default.target"]

NKey: UAMLBE4LIDYXH6FO5LW6CPLEBZJ23I2OT3TPZYIJPQY5HJ46X3BU7UJK

["default.target"]
```

#### Get a property on all machines

```terminal
❯ nix run .# -- get org.freedesktop.systemd1 /org/freedesktop/systemd1/unit/basic_2etarget ActiveState
NKey: UBQUIKYGFC7CH5XMF52P2NN4ESI4XTXGAKT3WTG3XGU3352DMHGZQAX7
Signature: s

"active"

NKey: UAMLBE4LIDYXH6FO5LW6CPLEBZJ23I2OT3TPZYIJPQY5HJ46X3BU7UJK
Signature: s

"active"

NKey: UAILZJS66U24VALTH2D6OWVD5JTS3IDFMKOICJSOX7CHNAF3AXFVWIFX
Signature: s

"active"
```

#### Get a property on a list of machines by NATS NKey

```terminal
❯ nix run .# -- get org.freedesktop.systemd1 /org/freedesktop/systemd1/unit/basic_2etarget ActiveState --nkeys UBQUIKYGFC7CH5XMF52P2NN4ESI4XTXGAKT3WTG3XGU3352DMHGZQAX7,UAMLBE4LIDYXH6FO5LW6CPLEBZJ23I2OT3TPZYIJPQY5HJ46X3BU7UJK
Signature: s
NKey: UAMLBE4LIDYXH6FO5LW6CPLEBZJ23I2OT3TPZYIJPQY5HJ46X3BU7UJK

"active"

NKey: UBQUIKYGFC7CH5XMF52P2NN4ESI4XTXGAKT3WTG3XGU3352DMHGZQAX7
Signature: s

"active"
```

#### Listen for all signals from each machine

```terminal
❯ nats --context TestAdmin sub "dbus.signals.>"
12:14:06 Subscribing on dbus.signals.>
[#1] Received on "dbus.signals.UBQUIKYGFC7CH5XMF52P2NN4ESI4XTXGAKT3WTG3XGU3352DMHGZQAX7._1_1.org.freedesktop.systemd1.unit.nscd_2eservice"
Interface: org.freedesktop.DBus.Properties
Member: PropertiesChanged
NKey: UBQUIKYGFC7CH5XMF52P2NN4ESI4XTXGAKT3WTG3XGU3352DMHGZQAX7
Path: /org/freedesktop/systemd1/unit/nscd_2eservice
Sender: :1.1

["org.freedesktop.systemd1.Service",{"CleanResult":{},"ControlPID":{},"ExecMainCode":{},"ExecMainExitTimestamp":{},"ExecMainExitTimestampMonotonic":{},"ExecMainPID":{},"ExecMainStartTimestamp":{},"ExecMainStartTimestampMonotonic":{},"ExecMainStatus":{},"GID":{},"MainPID":{},"NRestarts":{},"NotifyAccess":{},"ReloadResult":{},"Result":{},"StatusErrno":{},"StatusText":{},"UID":{}},["ExecCondition","ExecConditionEx","ExecStartPre","ExecStartPreEx","ExecStart","ExecStartEx","ExecStartPost","ExecStartPostEx","ExecReload","ExecReloadEx","ExecStop","ExecStopEx","ExecStopPost","ExecStopPostEx"]]


[#2] Received on "dbus.signals.UBQUIKYGFC7CH5XMF52P2NN4ESI4XTXGAKT3WTG3XGU3352DMHGZQAX7._1_1.org.freedesktop.systemd1.unit.nscd_2eservice"
Interface: org.freedesktop.DBus.Properties
Member: PropertiesChanged
NKey: UBQUIKYGFC7CH5XMF52P2NN4ESI4XTXGAKT3WTG3XGU3352DMHGZQAX7
Path: /org/freedesktop/systemd1/unit/nscd_2eservice
Sender: :1.1

["org.freedesktop.systemd1.Unit",{"ActivationDetails":{},"ActiveEnterTimestamp":{},"ActiveEnterTimestampMonotonic":{},"ActiveExitTimestamp":{},"ActiveExitTimestampMonotonic":{},"ActiveState":{},"AssertResult":{},"AssertTimestamp":{},"AssertTimestampMonotonic":{},"ConditionResult":{},"ConditionTimestamp":{},"ConditionTimestampMonotonic":{},"FreezerState":{},"InactiveEnterTimestamp":{},"InactiveEnterTimestampMonotonic":{},"InactiveExitTimestamp":{},"InactiveExitTimestampMonotonic":{},"InvocationID":{},"Job":{},"StateChangeTimestamp":{},"StateChangeTimestampMonotonic":{},"SubState":{}},["Conditions","Asserts"]]
```

## License

This software is provided free under the [MIT Licence](https://opensource.org/licenses/MIT).

## Contact

There are a few different ways to reach me, all of which are listed on my [website](https://bmcgee.ie/).
