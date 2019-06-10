module github.com/ledboot/knative-kong-ingress

go 1.12

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20180904163835-0709b304e793
	golang.org/x/net => github.com/golang/net v0.0.0-20190108225652-1e06a53dbb7e
	golang.org/x/sys => github.com/golang/sys v0.0.0-20181116152217-5ac8a444bdc5
	golang.org/x/text => github.com/golang/text v0.3.0
	k8s.io/api => github.com/kubernetes/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apiextensions-apiserver => github.com/apiextensions-apiserver v0.0.0-20190221221350-bfb440be4b87
	k8s.io/apimachinery => github.com/kubernetes/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go => github.com/kubernetes/client-go v10.0.0+incompatible
	k8s.io/kube-openapi => github.com/kubernetes/kube-openapi v0.0.0-20190603182131-db7b694dc208
	sigs.k8s.io/controller-runtime v0.1.11 => github.com/kubernetes-sigs/controller-runtime v0.1.11
	sigs.k8s.io/structured-merge-diff => github.com/kubernetes-sigs/structured-merge-diff v0.0.0-20190525122527-15d366b2352e
)

require (
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30 // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.3.0 // indirect
	github.com/google/go-containerregistry v0.0.0-20190531175139-2687bd5ba651 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/hbagdi/go-kong v0.5.0
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/knative/pkg v0.0.0-20190607011741-d8a2a739d6b3
	github.com/knative/serving v0.6.0
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/operator-framework/operator-sdk v0.8.1
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/prometheus/client_golang v0.9.3 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20190602205700-9b8cae951d65
	k8s.io/apimachinery v0.0.0-20190606174813-5a6182816fbf
	k8s.io/client-go v11.0.0+incompatible // indirect
	k8s.io/kube-openapi v0.0.0-20190603182131-db7b694dc208 // indirect
	k8s.io/utils v0.0.0-20190529001817-6999998975a7 // indirect
	sigs.k8s.io/controller-runtime v0.1.11
	sigs.k8s.io/yaml v1.1.0 // indirect
)
