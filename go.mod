module github.com/david-gurley/nomad-vf-plugin

go 1.17

require (
	github.com/david-gurley/host v1.0.33
	github.com/hashicorp/go-hclog v0.9.1
	github.com/hashicorp/nomad v0.10.0-beta1.0.20191119152219-a9490506dc2a
	github.com/kr/pretty v0.1.0
)

require (
	github.com/LK4D4/joincontext v0.0.0-20171026170139-1724345da6d5 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/armon/go-metrics v0.0.0-20190430140413-ec5e00d3c878 // indirect
	github.com/david-gurley/pci v0.1.5 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/gorhill/cronexpr v0.0.0-20180427100037-88b0669f7d75 // indirect
	github.com/hashicorp/consul/api v1.2.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.1.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/go-plugin v1.0.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.1 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/hcl2 v0.0.0-20191002203319-fb75b3253c80 // indirect
	github.com/hashicorp/raft v1.1.1 // indirect
	github.com/hashicorp/serf v0.8.2 // indirect
	github.com/hashicorp/vault/api v1.0.4 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/jaypipes/ghw v0.8.0 // indirect
	github.com/jaypipes/pcidb v0.6.0 // indirect
	github.com/jinzhu/copier v0.3.2 // indirect
	github.com/kr/text v0.1.0 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/safchain/ethtool v0.0.0-20210803160452-9aa261dae9b1 // indirect
	github.com/shirou/gopsutil v0.0.0-00010101000000-000000000000 // indirect
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/ugorji/go v0.0.0-00010101000000-000000000000 // indirect
	github.com/vishvananda/netlink v1.1.0 // indirect
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df // indirect
	github.com/vmihailenco/msgpack v3.3.3+incompatible // indirect
	github.com/zclconf/go-cty v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20191029031824-8986dd9e96cf // indirect
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859 // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/appengine v1.4.0 // indirect
	google.golang.org/genproto v0.0.0-20190404172233-64821d5d2107 // indirect
	google.golang.org/grpc v1.22.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	howett.net/plist v0.0.0-20181124034731-591f970eefbb // indirect
)

// don't use shirou/gopsutil, use the hashicorp fork
replace github.com/shirou/gopsutil => github.com/hashicorp/gopsutil v2.17.13-0.20190117153606-62d5761ddb7d+incompatible

// don't use ugorji/go, use the hashicorp fork
replace github.com/ugorji/go => github.com/hashicorp/go-msgpack v0.0.0-20190927123313-23165f7bc3c2

// fix the version of hashicorp/go-msgpack to 96ddbed8d05b
replace github.com/hashicorp/go-msgpack => github.com/hashicorp/go-msgpack v0.0.0-20191101193846-96ddbed8d05b
