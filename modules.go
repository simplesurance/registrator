package main

import (
	_ "github.com/simplesurance/registrator/consul"
	_ "github.com/simplesurance/registrator/consulkv"
	_ "github.com/simplesurance/registrator/etcd"
	_ "github.com/simplesurance/registrator/skydns2"
	_ "github.com/simplesurance/registrator/zookeeper"
)
