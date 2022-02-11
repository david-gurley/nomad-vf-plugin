Nomad vfio-pci Passthrough Device Plugin
==================

This project provides support for allocating vfio-pci devices for passthrough
to user space proceses - initially focused on QEMU. 

Requirements
------------

- TBD 

Attributes
----------

* vendor_name 

Agent
------
valid configuration options:

Job
----
The device stanza allows the standard constraint and affinity stanzas to specify what kind of passhthrough device to use.

```
device "vfio-pci" {}
```


